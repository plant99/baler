package baler

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func UnConvert(sourceDir string, destinationDir string) error {
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

		scanner := bufio.NewScanner(file)
		var destinationFile *os.File
		var destinationFileName string

		buf := make([]byte, 0, 64*1024)
		// TODO: should come from MaxOutputFileSize
		// TODO: very long lines break this
		scanner.Buffer(buf, 5*1024*1024)

		for scanner.Scan() {
			line := scanner.Text()

			// check for filename tag
			if strings.HasPrefix(line, "// filename: ") {
				if destinationFile != nil {
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

				destinationFile, err = os.OpenFile(destinationPath, os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					return fmt.Errorf("failed to create file: %s. error: %w", destinationFileName, err)
				}
				scanner.Scan()
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
