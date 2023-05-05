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
