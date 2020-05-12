package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSkipHeader(t *testing.T) {
	for _, tt := range []struct {
		input    []byte
		expected []byte
	}{
		{
			input:    []byte("hello"),
			expected: []byte("hello"),
		},
		{
			input:    []byte("hello\n"),
			expected: []byte(""),
		},
		{
			input:    []byte("\n\n"),
			expected: []byte("\n"),
		},
		{
			input:    []byte("ABC\nWake up\n"),
			expected: []byte("Wake up\n"),
		},
		{
			input:    []byte{},
			expected: []byte{},
		},
	} {
		actual := skipHeader(tt.input)

		assert.Equal(t, tt.expected, actual)
	}
}
