package main

import (
	"os"
	"strings"
	"testing"
)

func TestParseAILine(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedStr string
		expectedInt int
		expectedFlt float64
		expectedErr string
	}{
		{
			name:        "Valid Input",
			input:       "AI:test,42,0.42",
			expectedStr: "test",
			expectedInt: 42,
			expectedFlt: 0.42,
			expectedErr: "",
		},
		{
			name:        "Invalid Prefix",
			input:       "AIt:test,42,0.42",
			expectedErr: "Invalid string field in line",
		},
		{
			name:        "Invalid Integer",
			input:       "AI:test,abc,0.42",
			expectedErr: "Invalid integer field in line",
		},
		{
			name:        "Negative Integer",
			input:       "AI:test,-5,0.42",
			expectedErr: "Negative integer field in line",
		},
		{
			name:        "Invalid Float",
			input:       "AI:test,42,xyz",
			expectedErr: "Invalid float field in line",
		},
		{
			name:        "Float Out of Range",
			input:       "AI:test,42,1.2",
			expectedErr: "Invalid float range in line",
		},
		{
			name:        "Empty String",
			input:       "AI:,42,0.42",
			expectedErr: "Empty string field in line",
		},
		{
			name:        "Invalid Fields Count",
			input:       "AI:test,42",
			expectedErr: "Invalid number of fields in line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str, num, flt, err := parseAILine(tt.input)

			if tt.expectedErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.expectedErr, err)
				}
				if err != nil && (str != "gpt-3.5-turbo" || num != 400 || flt != 0.7) {
					t.Errorf("Expected '%s, %d, %f', got '%v, %v, %v'", "gpt-3.5-turbo", 400, 0.7, str, num, flt)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if str != tt.expectedStr {
					t.Errorf("Expected string '%s', got '%s'", tt.expectedStr, str)
				}

				if num != tt.expectedInt {
					t.Errorf("Expected integer %d, got %d", tt.expectedInt, num)
				}

				if flt != tt.expectedFlt {
					t.Errorf("Expected float %f, got %f", tt.expectedFlt, flt)
				}
			}
		})
	}
}
func TestUnescape(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "hello\\tworld",
			expected: "hello\tworld",
		},
		{
			input:    "hello\\\\world",
			expected: "hello\\world",
		},
		{
			input:    `hello\nworld`,
			expected: "hello\nworld",
		},
		{
			input:    `hello\\nworld`,
			expected: `hello\nworld`,
		},
		{
			input:    `hello\\"`,
			expected: `hello\"`,
		},
		{
			input:    `hello \world"`,
			expected: "hello \\world\"",
		},
	}

	for _, tt := range tests {
		result := unescape(tt.input)
		if result != tt.expected {
			t.Errorf("unescape(%s): expected %s, but got %s", tt.input, tt.expected, result)
		}
	}
}
func TestFindLastAILine(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name: "No AI line",
			text: `This is a test
Another line
`,
			expected: `This is a test
Another line
`,
		},
		{
			name: "One AI line",
			text: `This is a test
AI: model=base, max_tokens=50, temperature=0.5
Another line
`,
			expected: "This is a test\n",
		},
		{
			name: "Multiple lines with AI line",
			text: `This is a test
Another line
AI: model=base, max_tokens=50, temperature=0.5
`,
			expected: `This is a test
Another line
`,
		},
		{
			name: "Multiple AI lines",
			text: `AI: model=base, max_tokens=50, temperature=0.5
This is a test
AI: model=custom, max_tokens=20, temperature=0.8
`,
			expected: `AI: model=base, max_tokens=50, temperature=0.5
This is a test
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			part1, aiLine := findLastAILine(tc.text)
			if part1 != tc.expected {
				t.Errorf("Expected part1 to be %q, but was %q", tc.expected, part1)
			}
			if aiLine == "" && strings.Contains(tc.text, "AI:") {
				t.Errorf("Expected aiLine to be non-empty")
			}
		})
	}
}
func TestReplaceExtension(t *testing.T) {
	// Test with filename with extension
	inputFilename := "example.txt"
	newExt := "html"
	expectedOutput := "example.html"
	output := replaceExtension(inputFilename, newExt)
	if output != expectedOutput {
		t.Errorf("Expected output to be %s, but got %s", expectedOutput, output)
	}

	// Test with filename without extension
	inputFilename = "example"
	newExt = "html"
	expectedOutput = "example.html"
	output = replaceExtension(inputFilename, newExt)
	if output != expectedOutput {
		t.Errorf("Expected output to be %s, but got %s", expectedOutput, output)
	}

	// Test with new extension with dot prefix
	inputFilename = "example.txt"
	newExt = ".html"
	expectedOutput = "example.html"
	output = replaceExtension(inputFilename, newExt)
	if output != expectedOutput {
		t.Errorf("Expected output to be %s, but got %s", expectedOutput, output)
	}

	// Test with empty new extension
	inputFilename = "example.txt"
	newExt = ""
	expectedOutput = "example"
	output = replaceExtension(inputFilename, newExt)
	if output != expectedOutput {
		t.Errorf("Expected output to be %s, but got %s", expectedOutput, output)
	}
}
func TestCopyFile(t *testing.T) {
	// create a temporary file for testing
	srcFile, err := os.CreateTemp("", "test_src")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(srcFile.Name())
	if _, err := srcFile.WriteString("test string"); err != nil {
		t.Fatal(err)
	}
	if err := srcFile.Close(); err != nil {
		t.Fatal(err)
	}

	// create a temporary file for the destination
	destFile, err := os.CreateTemp("", "test_dest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(destFile.Name())

	// run the copy function
	err = copyFile(srcFile.Name(), destFile.Name())
	if err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	// check that the contents of the destination file match the source file
	destContents, err := os.ReadFile(destFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	if string(destContents) != "test string" {
		t.Fatalf("destination file contents do not match source file")
	}
}
func TestOverwriteFile(t *testing.T) {
	// setup test environment
	filename := "test.txt"
	ext := ".bak"
	backupFilename := replaceExtension(filename, ext)
	content := "hello world"
	if err := os.WriteFile(filename, []byte("original"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	defer os.Remove(filename)
	defer os.Remove(backupFilename)

	// call function
	err := overwriteFile(filename, ext, content)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// verify file content
	b, err := os.ReadFile(filename)

	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(b) != content {
		t.Errorf("unexpected file content: got %v, want %v", string(b), content)
	}

	// verify backup file
	b, err = os.ReadFile(backupFilename)
	if err != nil {
		t.Fatalf("failed to read backup file: %v", err)
	}
	if string(b) != "original" {
		t.Errorf("unexpected backup file content: got %v, want %v", string(b), "original")
	}
}
func TestCheckFileArgs(t *testing.T) {
	testCases := []struct {
		Name         string
		Filenames    []string
		Expected     []string
		ExpectedErrs []error
	}{
		{
			Name:         "All files exist",
			Filenames:    []string{"file1.txt", "file2.txt", "file3.txt"},
			Expected:     []string{"file1.txt", "file2.txt", "file3.txt"},
			ExpectedErrs: nil,
		},
		{
			Name:         "Some files do not exist",
			Filenames:    []string{"file1.txt", "file2.txt", "file3.txt", "file4.txt"},
			Expected:     []string{"file1.txt", "file2.txt", "file3.txt", "file4.txt"},
			ExpectedErrs: nil,
		},
		{
			Name:         "All files do not exist",
			Filenames:    []string{"file5.txt", "file6.txt", "file7.txt"},
			Expected:     []string{"file5.txt", "file6.txt", "file7.txt"},
			ExpectedErrs: nil,
		},
	}

	// create expected files
	for _, file := range testCases[0].Filenames {
		_, err := os.Create(file)
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			actual, actualErrs := checkFileArgs(testCase.Filenames)
			if len(actual) != len(testCase.Expected) {
				t.Errorf("Expected %v, but got %v", testCase.Expected, actual)
			}

			if len(actualErrs) != len(testCase.ExpectedErrs) {
				t.Errorf("Expected %v, but got %v", testCase.ExpectedErrs, actualErrs)
			}
		})
	}

	// delete expected files
	for _, file := range testCases[0].Filenames {
		_ = os.Remove(file)
	}
	for _, file := range []string{"file4.txt", "file5.txt", "file6.txt", "file7.txt"} {
		_ = os.Remove(file)
	}
}

func TestProcessAuthorComments(t *testing.T) {
	prefix := "#"
	text := `
# This is a comment
#OUT
Not a comment
#IN
# Another comment
Still not a comment`

	expected := `
Still not a comment`

	result := processAuthorComments(text, prefix)

	if !strings.EqualFold(expected, result) {
		t.Errorf("Expected '%s' but got '%s'", expected, result)
	}
}
