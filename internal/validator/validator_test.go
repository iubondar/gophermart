package validator

import (
	"testing"
)

func TestValidateLuhn(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid card number", "4003600000000014", true},
		{"Invalid card number", "4111111111111112", false},
		{"Empty string", "", false},
		{"Non-digit characters", "4111-1111-1111-1111", false},
		{"Single digit", "5", false},
		{"Single digit positive", "0", true},
		{"Two digits", "12", false},
		{"Two digits valid", "18", true},
		{"Long valid number", "4532015112830366", true},
		{"Long invalid number", "4532015112830367", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateLuhn(tt.input); got != tt.expected {
				t.Errorf("ValidateLuhn(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
