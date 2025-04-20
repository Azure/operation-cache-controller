package rand

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name   string
		length int
		want   struct {
			length  int
			charset string
		}
	}{
		{
			name:   "generate string of length 0",
			length: 0,
			want: struct {
				length  int
				charset string
			}{
				length:  0,
				charset: letterBytes,
			},
		},
		{
			name:   "generate string of length 10",
			length: 10,
			want: struct {
				length  int
				charset string
			}{
				length:  10,
				charset: letterBytes,
			},
		},
		{
			name:   "generate string of length 32",
			length: 32,
			want: struct {
				length  int
				charset string
			}{
				length:  32,
				charset: letterBytes,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate multiple strings to ensure randomness
			generated := make(map[string]bool)
			for i := 0; i < 100; i++ {
				result := GenerateRandomString(tt.length)

				// Check length
				assert.Equal(t, tt.want.length, len(result), "generated string length should match requested length")

				// Check that string only contains valid characters
				for _, char := range result {
					assert.Contains(t, tt.want.charset, string(char), "generated string should only contain valid characters")
				}

				// Track uniqueness
				generated[result] = true
			}

			// For non-zero length strings, check that we're getting different values (randomness check)
			if tt.length > 0 {
				// With 100 generations, we should get at least 95 unique strings for any non-zero length
				// This is a probabilistic test, but the chance of failure is extremely low with proper randomness
				assert.Greater(t, len(generated), 95, "should generate mostly unique strings")
			}
		})
	}
}
