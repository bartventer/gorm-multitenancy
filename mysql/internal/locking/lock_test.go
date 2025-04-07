package locking

import (
	"crypto/sha1"
	"encoding/hex"
	"testing"
)

func TestEncodeKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Key within 64 characters",
			input:    "short_key",
			expected: "short_key",
		},
		{
			name:  "Key exceeding 64 characters",
			input: "this_is_a_very_long_key_that_exceeds_the_sixty_four_character_limit_and_needs_to_be_hashed",
			expected: func() string {
				hash := sha1.Sum([]byte("this_is_a_very_long_key_that_exceeds_the_sixty_four_character_limit_and_needs_to_be_hashed"))
				return hex.EncodeToString(hash[:]) // SHA-1 hash is always 40 characters
			}(),
		},
		{
			name:     "Empty key",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := encodeKey(tt.input)
			if result != tt.expected {
				t.Errorf("encodeKey(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
