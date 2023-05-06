package main

import (
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
