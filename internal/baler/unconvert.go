package baler

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func removeTrailingNewline(file *os.File) error {
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file stats: %w", err)
	}

	if info.Size() > 0 {
		lastByte := make([]byte, 1)
		if _, err := file.Seek(-1, io.SeekEnd); err != nil {
			return fmt.Errorf("failed to seek file: %w", err)
		}
		if _, err := file.Read(lastByte); err != nil {
			return fmt.Errorf("failed to read last byte: %w", err)
		}

		// If last byte is newline, remove it
		// This path is ALWAYS reached
		if lastByte[0] == '\n' {
			if err := file.Truncate(info.Size() - 1); err != nil {
				return fmt.Errorf("failed to truncate file: %w", err)
			}
		}
	}
	return nil
}

func UnConvert(sourceDir string, destinationDir string, config *BalerConfig) error {
	if _, err := os.Stat(sourceDir); err != nil {
		return fmt.Errorf("source directory not found: %w", err)
	}
	if _, err := os.Stat(destinationDir); err != nil {
		return fmt.Errorf("destination directory not found: %w", err)
	}
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return fmt.Errorf("unable to list source dir: %w", err)
	}
	if len(entries) == 0 {
		return fmt.Errorf("found no files to process.")
	}
	var sourcePaths []string
	for _, entry := range entries {
		sourcePaths = append(sourcePaths, filepath.Join(sourceDir, entry.Name()))
	}
	for _, path := range sourcePaths {
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open source file: %w", err)
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
					if err := removeTrailingNewline(destinationFile); err != nil {
						return fmt.Errorf("unable to remove trailing \n: %w", err)
					}
					if err := destinationFile.Close(); err != nil {
						return fmt.Errorf("unable to close file reference: %w", err)
					}
				}

				destinationFileName = strings.TrimPrefix(line, "// filename: ")
				destinationFileName = strings.TrimSpace(destinationFileName)

				destinationPath := filepath.Join(destinationDir, destinationFileName)
				if err := os.MkdirAll(filepath.Dir(destinationPath), 0755); err != nil {
					return fmt.Errorf("failed to create directory for path: %s . Error: %w", destinationPath, err)
				}
				// the read mode is required to remove the extra '\n'
				// enhancement: it's possible to create isolated file references with different permissions
				destinationFile, err = os.OpenFile(destinationPath, os.O_CREATE|os.O_RDWR, 0644)
				if err != nil {
					return fmt.Errorf("failed to create file: %s. error: %w", destinationFileName, err)
				}
				continue
			}

			if destinationFile != nil {
				if _, err := destinationFile.WriteString(line + "\n"); err != nil {
					return fmt.Errorf("failed to write to file %s: %w", destinationFileName, err)
				}
			}
		}
		if destinationFile != nil {
			if err := destinationFile.Close(); err != nil {
				return fmt.Errorf("unable to close file reference: %w", err)
			}
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("error scanning file %s: %w", path, err)
		}
	}
	return nil
}
