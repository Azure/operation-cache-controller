package ptr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// test case for Of

func TestOf(t *testing.T) {
	t.Run("string to ptr", func(t *testing.T) {
		assert := assert.New(t)
		assert.Equal(*Of("a"), "a")
		assert.Equal(*Of("b"), "b")
	})

	t.Run("int to ptr", func(t *testing.T) {
		assert := assert.New(t)
		assert.Equal(*Of(1), 1)
		assert.Equal(*Of(2), 2)
	})

	t.Run("bool to ptr", func(t *testing.T) {
		assert := assert.New(t)
		assert.Equal(*Of(true), true)
		assert.Equal(*Of(false), false)
	})

	t.Run("float to ptr", func(t *testing.T) {
		assert := assert.New(t)
		assert.Equal(*Of(1.1), 1.1)
		assert.Equal(*Of(2.2), 2.2)
	})

}

// test case for Deref

func TestDeref(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "string",
			input:    "a",
			expected: "a",
		},
		{
			name:     "int",
			input:    1,
			expected: 1,
		},
		{
			name:     "bool true",
			input:    true,
			expected: true,
		},
		{
			name:     "float",
			input:    1.1,
			expected: 1.1,
		},
		{
			name:     "bool false",
			input:    false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(Deref(Of(tt.input)), tt.expected)
		})
	}
}
