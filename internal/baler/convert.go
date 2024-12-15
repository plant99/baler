package baler

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"unicode/utf8"
)

// TODO: figure out how to store and update this information alt. global variable
var fileCounter = 0

type ConvertConfig struct {
	MaxInputFileLines uint64
	MaxInputFileSize  uint64
	MaxOutputFileSize uint64
	ExclusionPatterns *[]string
}

type ValidationResult struct {
	IsValidUTF8  bool
	IsValidLines bool
	IsValidSize  bool
	// artifacts
	Size uint64
}

func customScanner(file *os.File) *bufio.Scanner {
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 6*1024*1024)
	return scanner
}

func validateFile(fileName string, config *ConvertConfig) (*ValidationResult, error) {
	// TODO: check if it's feasible to return early
	isValidUTF8 := true
	isValidLines := true
	isValidSize := true
	// checks without opening the file
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	if fileInfo.Size() > int64(config.MaxInputFileSize) {
		isValidSize = false
	}
	// checks including reads of the file
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := customScanner(file)
	lineCount := uint32(0)

	for scanner.Scan() {
		lineCount++
		if lineCount <= 10 && !utf8.Valid(scanner.Bytes()) {
			isValidUTF8 = false
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning file: %v", err)
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

func shouldIgnore(relativePath string, patternList *[]string) (bool, error) {
	for _, pattern := range *patternList {
		matches, err := path.Match(pattern, relativePath)
		if err != nil {
			return false, err
		}
		if matches {
			return true, nil
		}
	}
	return false, nil
}

func copyContent(srcPath string, destFile *os.File, srcRelativePath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file %w", err)
	}
	defer srcFile.Close()

	reader := bufio.NewReader(srcFile)
	writer := bufio.NewWriter(destFile)

	if _, err := writer.WriteString(fmt.Sprintf("\n\n// filename: %s\n\n", srcRelativePath)); err != nil {
		return fmt.Errorf("failed to write filename comment: %w", err)
	}
	if _, err = io.Copy(writer, reader); err != nil {
		return fmt.Errorf("error copying file: %w", err)
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("error flushing writer: %w", err)
	}
	return nil
}
func convertDirectoryAndSaveToFile(absProcessingDirPath string, sourcePath string, destinationDir string, config *ConvertConfig) error {
	// sort entries alphabetically
	entries, err := os.ReadDir(absProcessingDirPath)
	if err != nil {
		return err
	}
	absSourcePath, err := filepath.Abs(sourcePath)
	if err != nil {
		return err
	}
	// reference to file in destinationPath
	outputFileName := fmt.Sprintf("output_%s.txt", strconv.Itoa(fileCounter))
	destinationFileName := filepath.Join(destinationDir, outputFileName)

	f, err := os.OpenFile(destinationFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	// TODO: how would clean-up work in recursion
	defer f.Close()

	// iterate through entries
	for _, entry := range entries {
		absPath := filepath.Join(absProcessingDirPath, entry.Name())
		relPath, err := filepath.Rel(absSourcePath, absPath)
		if err != nil {
			return err
		}

		// ignore logic
		if ignore, err := shouldIgnore(relPath, config.ExclusionPatterns); err != nil {
			return err
		} else if ignore {
			continue
		}

		// for each file write the file into converted text file
		// for each directory, run convertDirectoryAndSaveToFile recursively
		if !entry.IsDir() {
			// file validation before processing
			validationResult, err := validateFile(absPath, config)
			if err != nil {
				return err
			}
			if !validationResult.IsValidLines || !validationResult.IsValidSize || !validationResult.IsValidUTF8 {
				// TODO: log
				continue
			}
			// check if entry + existing sink file exceeds size limit
			// if so, increment file name counter and set it as sink
			currentDestinationFileInfo, err := f.Stat()
			if err != nil {
				return err
			}
			currentDestinationFileSize := currentDestinationFileInfo.Size()
			// 5242880 = 5 * 1024 * 1024
			if currentDestinationFileSize+int64(validationResult.Size) > int64(config.MaxOutputFileSize) {
				// close reference to old file
				f.Close()

				// update reference to new file
				/*
					TODO: Ideally the function call could be idempotent if 'READ' status
					is maintained somewhere.
					In which case, we could just increment fileCounter and call the
					function again.
				*/
				fileCounter += 1
				outputFileName = fmt.Sprintf("output_%s.txt", strconv.Itoa(fileCounter))
				destinationFileName = filepath.Join(destinationDir, outputFileName)
				f, err = os.OpenFile(destinationFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					return err
				}
				defer f.Close()
			}
			// perform copy
			if err = copyContent(absPath, f, relPath); err != nil {
				return err
			}

		} else {
			// process directory
			// TODO: Could there theoretically be infinite open references to files?
			convertDirectoryAndSaveToFile(absPath, sourcePath, destinationDir, config)
		}
	}
	return nil
}

func Convert(inputPath string, outputPath string, config *ConvertConfig) error {
	// check if input, output paths exists
	if _, err := os.Stat(inputPath); err != nil {
		return err
	}
	if _, err := os.Stat(outputPath); err != nil {
		return err
	}
	absInputPath, err := filepath.Abs(inputPath)
	if err != nil {
		return err
	}
	if err := convertDirectoryAndSaveToFile(absInputPath, inputPath, outputPath, config); err != nil {
		return err
	}
	return nil
}

func UnConvert(outputPath string, inputPath string) error {
	return nil
}
