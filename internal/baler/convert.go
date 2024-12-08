package baler

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"unicode/utf8"
)

// TODO: figure out how to store and update this information alt. global variable
var fileCounter = 0

func checkValidUTF8(fileName string) (bool, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return false, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 6*1024*1024)
	var content []byte
	for lineCount := 0; lineCount < 10 && scanner.Scan(); lineCount++ {
		content = append(content, scanner.Bytes()...)
		content = append(content, '\n')
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("error scanning file: %v", err)
	}
	return utf8.Valid(content), nil
}

func convertDirectoryAndSaveToFile(sourcePath string, destinationDir string) error {
	// sort entries alphabetically
	entries, err := os.ReadDir(sourcePath)
	if err != nil {
		return err
	}
	// create new file in destinationPath
	outputFileName := fmt.Sprintf("output_%s.txt", strconv.Itoa(fileCounter))
	destinationFileName := filepath.Join(destinationDir, outputFileName)

	// TODO: work with explicit mode (which doesn't assume things)
	filesToIgnore := []string{".DS_Store", "package-lock.json"}
	directoriesToIgnore := []string{"node_modules", ".git", ".expo"}
	f, err := os.OpenFile(destinationFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	// TODO: how would clean-up work in recursion
	defer f.Close()

	// absolute path preparation
	absSourcePath, err := filepath.Abs(sourcePath)
	if err != nil {
		return err
	}

	// the loop
	for _, entry := range entries {
		absPath := filepath.Join(absSourcePath, entry.Name())

		// for each file write the file into converted text file
		// for each directory, run convertDirectoryAndSaveToFile recursively
		if !entry.IsDir() {
			// ignore logic
			if slices.Contains(filesToIgnore, entry.Name()) {
				continue
			}

			// check valid utf-8
			validUtf, err := checkValidUTF8(absPath)
			if err != nil {
				return err
			}
			if !validUtf {
				// baler only supports text files
				continue
			}
			// read entry to data
			data, err := os.ReadFile(absPath)
			if err != nil {
				return err
			}
			file_name_comment := []byte("\n// filename: " + entry.Name() + "\n\n")
			data = append(file_name_comment, data...)
			// before writing this data make sure it doesn't exceed the filesize.
			currentDestinationFileInfo, err := f.Stat()
			if err != nil {
				return err
			}
			currentDestinationFileSize := currentDestinationFileInfo.Size()
			// 5242880 = ~5MB
			if currentDestinationFileSize+int64(len(data)) > int64(5242880) {
				fileCounter += 1
				f.Close()
				// update reference to file
				/*
					TODO: Ideally the function call could be idempotent if 'READ' status
					is maintained somewhere.
					In which case, we could just increment fileCounter and call the
					function again.
				*/
				outputFileName = fmt.Sprintf("output_%s.txt", strconv.Itoa(fileCounter))
				destinationFileName = filepath.Join(destinationDir, outputFileName)
				f, err = os.OpenFile(destinationFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					return err
				}
			}
			if _, err := f.Write(data); err != nil {
				return err
			}
		} else {
			if slices.Contains(directoriesToIgnore, entry.Name()) {
				continue
			}
			// process directory
			convertDirectoryAndSaveToFile(absPath, destinationDir)
		}
	}
	return nil
}

func Convert(path string, limit uint32) error {
	pathFixed := "/Users/shiv/code/experimental/heli-exp"
	// check if path exists
	_, err := os.Stat(pathFixed)
	if err != nil {
		// TODO: create custom internal errors
		return err
	}
	// create temporary directory
	dname, err := os.MkdirTemp("", "outputdir")
	if err != nil {
		return err
	}
	err = convertDirectoryAndSaveToFile(pathFixed, dname)
	if err != nil {
		return err
	}
	return nil
}
