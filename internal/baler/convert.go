package baler

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"
)

type ValidationResult struct {
	IsValidUTF8  bool
	IsValidLines bool
	IsValidSize  bool
	// artifacts
	Size uint64
}

func customScanner(file *os.File, config *BalerConfig) *bufio.Scanner {
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	var maxBufSize uint64
	if config.MaxBufferSize > 0 {
		maxBufSize = config.MaxBufferSize
	} else {
		maxBufSize = config.MaxInputFileSize
	}
	scanner.Buffer(buf, int(maxBufSize))
	return scanner
}

func validateFile(fileName string, config *BalerConfig) (*ValidationResult, *BalerError) {
	// TODO: check if it's feasible to return early
	isValidUTF8 := true
	isValidLines := true
	isValidSize := true
	// checks without opening the file
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return nil, NewIOError(fmt.Sprintf("failed to get file info for: %s", fileName), err)
	}

	if fileInfo.Size() > int64(config.MaxInputFileSize) {
		isValidSize = false
	}
	// checks including reads of the file
	file, err := os.Open(fileName)
	if err != nil {
		return nil, NewIOError(fmt.Sprintf("unable to open: %s", fileName), err)
	}
	defer file.Close()
	scanner := customScanner(file, config)
	lineCount := uint32(0)

	for scanner.Scan() {
		lineCount++
		if lineCount <= 10 && !utf8.Valid(scanner.Bytes()) {
			isValidUTF8 = false
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, NewIOError(fmt.Sprintf("error scanning file: %s", fileName), err)
	}
	if lineCount > uint32(config.MaxInputFileLines) {
		isValidLines = false
	}
	return &ValidationResult{
		IsValidUTF8:  isValidUTF8,
		IsValidLines: isValidLines,
		IsValidSize:  isValidSize,
		Size:         uint64(fileInfo.Size()),
	}, nil
}

func shouldIgnore(relativePath string, patternList *[]string) (bool, *BalerError) {
	for _, pattern := range *patternList {
		matches, err := path.Match(pattern, relativePath)
		if err != nil {
			return false, NewValidationError(
				fmt.Sprintf(
					"error matching path with pattern:  %s %s",
					pattern,
					relativePath,
				),
				err,
			)
		}
		if matches {
			return true, nil
		}
	}
	return false, nil
}

func copyContent(srcPath string, destFile *os.File, srcRelativePath string, fileDelimiter string) *BalerError {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return NewIOError("failed to open source file", err)
	}
	defer srcFile.Close()

	reader := bufio.NewReader(srcFile)
	writer := bufio.NewWriter(destFile)
	if _, err := writer.WriteString(fmt.Sprintf("\n%s%s\n", fileDelimiter, srcRelativePath)); err != nil {
		return NewIOError("failed to write filename comment", err)
	}
	if _, err = io.Copy(writer, reader); err != nil {
		return NewIOError("error copying file", err)
	}
	if err := writer.Flush(); err != nil {
		return NewIOError("error flushing writer", err)
	}
	return nil
}

func getValidIncreasedFileCounter(outputDir string, fileCounter int) (int, *BalerError) {
	// this function checks the next output_<integer>.txt in outputDir
	// such that integer > fileCounter
	// This function runs one per output file
	var nextBigInteger = fileCounter + 1
	// avoid collision
	fileList, err := os.ReadDir(outputDir)
	if err != nil {
		return fileCounter, NewIOError("unable to read output directory for file suffix", err)
	}
	// map of existing counters
	existingCounters := make(map[int]bool)
	for _, file := range fileList {
		name := file.Name()
		if !file.IsDir() && strings.HasPrefix(name, "output_") && strings.HasSuffix(name, ".txt") {
			numStr := name[7 : len(name)-4]
			if num, err := strconv.Atoi(numStr); err != nil {
				return fileCounter, NewValidationError("encountered invalid file in output directory", err)
			} else {
				existingCounters[num] = true
			}
		}
	}
	for {
		if _, exists := existingCounters[nextBigInteger]; !exists {
			break
		}
		nextBigInteger++
	}
	return nextBigInteger, nil
}

func convertDirectoryAndSaveToFile(absProcessingDirPath string, sourcePath string, destinationDir string, config *BalerConfig) (*[]string, *BalerError) {
	var fileCounter = 0
	processingStack := []string{absProcessingDirPath}
	filesProcessed := &[]string{}

	absSourcePath, err := filepath.Abs(sourcePath)
	if err != nil {
		return &[]string{}, NewValidationError(
			fmt.Sprintf("error getting absolute path for sourcePath: %s", sourcePath),
			err,
		)
	}
	// reference to file in destinationPath
	outputFileName := fmt.Sprintf("output_%s.txt", strconv.Itoa(fileCounter))
	destinationFileName := filepath.Join(destinationDir, outputFileName)

	destinationFile, err := os.OpenFile(destinationFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return &[]string{}, NewIOError(
			fmt.Sprintf("unable to open file: %s", destinationFileName),
			err,
		)
	}
	defer destinationFile.Close()

	for len(processingStack) > 0 {
		currentDir := processingStack[len(processingStack)-1]
		processingStack = processingStack[:len(processingStack)-1]

		entries, err := os.ReadDir(currentDir)
		if err != nil {
			return &[]string{}, NewIOError(fmt.Sprintf("unable to read directory: %s", currentDir), err)
		}
		// iterate through entries
		for _, entry := range entries {
			absPath := filepath.Join(currentDir, entry.Name())
			relPath, err := filepath.Rel(absSourcePath, absPath)
			if err != nil {
				return &[]string{}, NewIOError(fmt.Sprintf("unable to get relative filepath for %s", absPath), err)
			}

			// ignore logic
			if ignore, balerErr := shouldIgnore(relPath, config.ExclusionPatterns); balerErr != nil {
				return &[]string{}, balerErr
			} else if ignore {
				if config.Verbose {
					config.Logger.Info("Skipping excluded file: " + relPath)
				}
				continue
			}

			// for each file write the file into converted text file
			// for each directory, append to processingStack
			if !entry.IsDir() {
				// file validation before processing
				validationResult, balerErr := validateFile(absPath, config)
				if balerErr != nil {
					return &[]string{}, balerErr
				}
				if !validationResult.IsValidLines || !validationResult.IsValidSize || !validationResult.IsValidUTF8 {
					// TODO: log
					continue
				}
				// check if entry + existing sink file exceeds size limit
				// if so, increment file name counter and set it as sink
				currentDestinationFileInfo, err := destinationFile.Stat()
				if err != nil {
					return &[]string{}, NewIOError(
						fmt.Sprintf("unable to get information on %s", destinationFileName),
						err,
					)
				}
				currentDestinationFileSize := currentDestinationFileInfo.Size()
				if currentDestinationFileSize+int64(validationResult.Size) > int64(config.MaxOutputFileSize) {
					// close reference to old file
					destinationFile.Close()

					// update reference to new file
					/*
						TODO: Ideally the function call could be idempotent if 'READ' status
						is maintained somewhere.
						In which case, we could just increment fileCounter and call the
						function again.
					*/
					fileCounter, balerErr := getValidIncreasedFileCounter(destinationDir, fileCounter)
					if balerErr != nil {
						return &[]string{}, balerErr
					}
					outputFileName = fmt.Sprintf("output_%s.txt", strconv.Itoa(fileCounter))
					destinationFileName = filepath.Join(destinationDir, outputFileName)
					destinationFile, err = os.OpenFile(destinationFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					if err != nil {
						return &[]string{}, NewIOError(
							fmt.Sprintf("unable to open file %s", destinationFileName),
							err,
						)
					}
					defer destinationFile.Close()
				}
				// perform copy
				if balerErr = copyContent(absPath, destinationFile, relPath, config.FileDelimiter); balerErr != nil {
					return &[]string{}, balerErr
				}

			} else {
				processingStack = append(processingStack, absPath)
			}
			*filesProcessed = append(*filesProcessed, relPath)
			if config.Verbose {
				config.Logger.Info("Successfully processed file: " + relPath)
			}
		}
	}
	return filesProcessed, nil
}

func Convert(inputPath string, outputPath string, config *BalerConfig) (*[]string, *BalerError) {
	// check if input, output paths exists
	if _, err := os.Stat(inputPath); err != nil {
		return &[]string{}, NewIOError(
			fmt.Sprintf("unable to get information on %s. Are you sure this path exists? ", inputPath),
			err,
		)
	}
	if _, err := os.Stat(outputPath); err != nil {
		return &[]string{}, NewIOError(
			fmt.Sprintf("unable to get information on %s. Are you sure this path exists? ", outputPath),
			err,
		)
	}
	absInputPath, err := filepath.Abs(inputPath)
	if err != nil {
		return &[]string{}, NewIOError(
			fmt.Sprintf("unable to get absolute path for %s", inputPath),
			err,
		)
	}
	processedPaths, balerErr := convertDirectoryAndSaveToFile(absInputPath, inputPath, outputPath, config)
	if balerErr != nil {
		return &[]string{}, balerErr
	}
	return processedPaths, nil
}
