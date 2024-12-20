package baler

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func removeTrailingNewline(file *os.File) *BalerError {
	info, err := file.Stat()
	if err != nil {
		return NewIOError(
			fmt.Sprintf("failed to get file information for %s", file.Name()),
			err,
		)
	}

	if info.Size() > 0 {
		lastByte := make([]byte, 1)
		if _, err := file.Seek(-1, io.SeekEnd); err != nil {
			return NewIOError(
				fmt.Sprintf("failed to seek file %s", file.Name()),
				err,
			)
		}
		if _, err := file.Read(lastByte); err != nil {
			return NewIOError(
				fmt.Sprintf("failed to read last byte of %s", file.Name()),
				err,
			)
		}

		// If last byte is newline, remove it
		// This path is ALWAYS reached
		if lastByte[0] == '\n' {
			if err := file.Truncate(info.Size() - 1); err != nil {
				return NewIOError(
					fmt.Sprintf("failed to truncate file: %s", file.Name()),
					err,
				)
			}
		}
	}
	return nil
}

func UnConvert(sourceDir string, destinationDir string, config *BalerConfig) *BalerError {
	if _, err := os.Stat(sourceDir); err != nil {
		return NewValidationError(
			fmt.Sprintf("source directory doesn't exist: %s", sourceDir),
			err,
		)
	}
	if _, err := os.Stat(destinationDir); err != nil {
		return NewValidationError(
			fmt.Sprintf("destination directory doesn't exist: %s", destinationDir),
			err,
		)
	}
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return NewValidationError(
			fmt.Sprintf("unable to list source directory: %s", sourceDir),
			err,
		)
	}
	if len(entries) == 0 {
		return NewValidationError("no files to process", nil)
	}
	var sourcePaths []string
	for _, entry := range entries {
		sourcePaths = append(sourcePaths, filepath.Join(sourceDir, entry.Name()))
	}
	for _, path := range sourcePaths {
		file, err := os.Open(path)
		if err != nil {
			return NewIOError(
				fmt.Sprintf("failed to open source file: %s", path),
				err,
			)
		}
		defer file.Close()

		var destinationFile *os.File
		var destinationFileName string

		scanner := customScanner(file, config)
		for scanner.Scan() {
			line := scanner.Text()

			if strings.HasPrefix(line, config.FileDelimiter) {
				if destinationFile != nil {
					// there is ALWAYS a trailing \n from convert function
					// to the generated baler file
					// Either we keep a buffered line to write without this \n
					// if prefix is encountered. Or we do this hack to remove \n
					// after the file is written.
					if balerErr := removeTrailingNewline(destinationFile); balerErr != nil {
						return balerErr
					}
					if err := destinationFile.Close(); err != nil {
						return NewIOError(
							fmt.Sprintf("unable to close file reference for %s", destinationFileName),
							err,
						)
					}
				}

				destinationFileName = strings.TrimPrefix(line, "// filename: ")
				destinationFileName = strings.TrimSpace(destinationFileName)

				destinationPath := filepath.Join(destinationDir, destinationFileName)
				if err := os.MkdirAll(filepath.Dir(destinationPath), 0755); err != nil {
					return NewIOError(
						fmt.Sprintf("failed to create directory for path: %s", destinationPath),
						err,
					)
				}
				// the read mode is required to remove the extra '\n'
				// enhancement: it's possible to create isolated file references with different permissions
				destinationFile, err = os.OpenFile(destinationPath, os.O_CREATE|os.O_RDWR, 0644)
				if err != nil {
					return NewIOError(
						fmt.Sprintf("failed to create file: %s", destinationPath),
						err,
					)
				}
				continue
			}

			if destinationFile != nil {
				if _, err := destinationFile.WriteString(line + "\n"); err != nil {
					return NewIOError(
						fmt.Sprintf("failed to write to file: %s", destinationFileName),
						err,
					)
				}
			}
		}
		if destinationFile != nil {
			if err := destinationFile.Close(); err != nil {
				return NewIOError(
					fmt.Sprintf("unable to close file reference: %s", destinationFileName),
					err,
				)
			}
		}

		if err := scanner.Err(); err != nil {
			return NewIOError(
				fmt.Sprintf("error scanning file: %s", path),
				err,
			)
		}
	}
	return nil
}
