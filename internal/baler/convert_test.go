package baler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type mockLogger struct {
	infoMessages  []string
	warnMessages  []string
	errorMessages []string
}

func (l *mockLogger) Info(msg string)  { l.infoMessages = append(l.infoMessages, msg) }
func (l *mockLogger) Warn(msg string)  { l.warnMessages = append(l.warnMessages, msg) }
func (l *mockLogger) Error(msg string) { l.errorMessages = append(l.errorMessages, msg) }

func setupTestDir(t *testing.T) (string, func()) {
	dir, err := os.MkdirTemp("", "baler-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(dir)
	}

	return dir, cleanup
}

func createTestFile(t *testing.T, dir, name, content string) string {
	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	return path
}

func TestValidateFile(t *testing.T) {
	testDir, cleanup := setupTestDir(t)
	defer cleanup()

	tests := []struct {
		name           string
		content        string
		maxSize        uint64
		maxLines       uint64
		expectValid    bool
		expectedResult *ValidationResult
	}{
		{
			name:        "Valid small file",
			content:     "test\nfile\ncontent",
			maxSize:     1024,
			maxLines:    10,
			expectValid: true,
			expectedResult: &ValidationResult{
				IsValidUTF8:  true,
				IsValidLines: true,
				IsValidSize:  true,
			},
		},
		{
			name:        "File exceeding max lines",
			content:     "1\n2\n3\n4\n5\n6",
			maxSize:     1024,
			maxLines:    3,
			expectValid: true,
			expectedResult: &ValidationResult{
				IsValidUTF8:  true,
				IsValidLines: false,
				IsValidSize:  true,
			},
		},
		{
			name:        "File exceeding max size",
			content:     "large content",
			maxSize:     5,
			maxLines:    10,
			expectValid: true,
			expectedResult: &ValidationResult{
				IsValidUTF8:  true,
				IsValidLines: true,
				IsValidSize:  false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := createTestFile(t, testDir, "test.txt", tt.content)
			config := &BalerConfig{
				MaxInputFileSize:  tt.maxSize,
				MaxInputFileLines: tt.maxLines,
				Logger:            &NoopLogger{},
			}

			result, err := validateFile(filePath, config)
			if err != nil && tt.expectValid {
				t.Errorf("Expected valid file, got error: %v", err)
			}

			if result != nil {
				if result.IsValidLines != tt.expectedResult.IsValidLines {
					t.Errorf("Expected IsValidLines=%v, got %v",
						tt.expectedResult.IsValidLines, result.IsValidLines)
				}
				if result.IsValidSize != tt.expectedResult.IsValidSize {
					t.Errorf("Expected IsValidSize=%v, got %v",
						tt.expectedResult.IsValidSize, result.IsValidSize)
				}
				if result.IsValidUTF8 != tt.expectedResult.IsValidUTF8 {
					t.Errorf("Expected IsValidUTF8=%v, got %v",
						tt.expectedResult.IsValidUTF8, result.IsValidUTF8)
				}
			}
		})
	}
}

func TestShouldIgnore(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		patterns     []string
		expectIgnore bool
		expectError  bool
	}{
		{
			name: "Should ignore node_modules",
			// Note that path.Matches only needs to be checked for
			// root level ignore, and not children, so ** need not work
			// in our usecase
			path:         "node_modules/package",
			patterns:     []string{"node_modules/*"},
			expectIgnore: true,
			expectError:  false,
		},
		{
			name:         "Should ignore node_modules",
			path:         "node_modules/package",
			patterns:     []string{"node_modules*"},
			expectIgnore: false,
			expectError:  false,
		},
		{
			name:         "Should not ignore regular file",
			path:         "src/main.go",
			patterns:     []string{"node_modules*", "*.tmp"},
			expectIgnore: false,
			expectError:  false,
		},
		{
			name:         "Invalid pattern should error",
			path:         "test.go",
			patterns:     []string{"[invalid"},
			expectIgnore: false,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ignore, err := shouldIgnore(tt.path, &tt.patterns)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if ignore != tt.expectIgnore {
				t.Errorf("Expected ignore=%v but got %v", tt.expectIgnore, ignore)
			}
		})
	}
}

func TestIntegration(t *testing.T) {
	// TODO:
	// 	add tests for unhappy paths,
	//  take common parts out to run multiple testcases
	sourceDir, sourceCleanup := setupTestDir(t)
	defer sourceCleanup()

	destDir, destCleanup := setupTestDir(t)
	defer destCleanup()

	files := map[string]string{
		"main.go":              "package main\nfunc main() {}\n",
		"lib/helper.go":        "package lib\nfunc Helper() {}\n",
		"node_modules/test.js": "console.log('test')",
		"test/large_file.txt":  string(make([]byte, 1024*1024)),
	}

	for path, content := range files {
		fullPath := filepath.Join(sourceDir, path)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		err = os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}
	logger := &mockLogger{}
	config := &BalerConfig{
		MaxInputFileSize:  2 * 1024 * 1024,
		MaxInputFileLines: 1000,
		MaxOutputFileSize: 3 * 1024 * 1024,
		ExclusionPatterns: &[]string{"node_modules*"},
		FileDelimiter:     "// filename: ",
		Logger:            logger,
		Verbose:           true,
	}

	processedFiles, balerErr := Convert(sourceDir, destDir, config)
	if balerErr != nil {
		t.Fatalf("Convert failed: %v", balerErr)
	}

	if len(*processedFiles) == 0 {
		t.Error("Expected processed files but got none")
	}
	outputFiles, err := os.ReadDir(destDir)
	if err != nil {
		t.Fatalf("Failed to read destination directory: %v", err)
	}

	if len(outputFiles) == 0 {
		t.Error("No output files generated")
	}

	// Verify excluded files
	for _, file := range *processedFiles {
		cleanPath := filepath.Clean(file)
		cleanPrefix := filepath.Clean("node_modules")
		relPath := filepath.ToSlash(cleanPath)
		if strings.HasPrefix(relPath, cleanPrefix) {
			t.Error("node_modules should have been excluded")
		}
	}

	unconvertDir, unconvertCleanup := setupTestDir(t)
	defer unconvertCleanup()

	unconvertConfig := &BalerConfig{
		MaxInputFileSize: config.MaxOutputFileSize,
		MaxBufferSize:    config.MaxBufferSize,
		FileDelimiter:    config.FileDelimiter,
		Logger:           logger,
		Verbose:          true,
		Operation:        OperationUnconvert,
	}

	balerErr = UnConvert(destDir, unconvertDir, unconvertConfig)
	if balerErr != nil {
		t.Fatalf("Unconvert failed: %v", balerErr)
	}

	// unconverted files match original files
	for path, expectedContent := range files {
		if strings.HasPrefix(path, "node_modules/") {
			continue
		}

		unconvertedPath := filepath.Join(unconvertDir, path)
		content, err := os.ReadFile(unconvertedPath)
		if err != nil {
			t.Errorf("Failed to read unconverted file %s: %v", unconvertedPath, err)
			continue
		}

		expectedContent = strings.ReplaceAll(expectedContent, "\r\n", "\n")
		gotContent := strings.ReplaceAll(string(content), "\r\n", "\n")

		if gotContent != expectedContent {
			t.Errorf("Content mismatch for %s\nExpected:\n%s\nGot:\n%s",
				path, expectedContent, gotContent)
		}
	}

	err = filepath.Walk(unconvertDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(unconvertDir, path)
		if err != nil {
			return err
		}
		if strings.HasPrefix(relPath, "node_modules") {
			t.Errorf("Excluded file/directory was recreated: %s", relPath)
		}
		return nil
	})
	if err != nil {
		t.Errorf("Failed to walk unconverted directory: %v", err)
	}

	if len(logger.infoMessages) == 0 {
		t.Error("Expected info messages in verbose mode")
	}
}
