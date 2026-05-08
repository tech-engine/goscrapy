package cli

import (
	"bufio"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetGoVersionFromMod(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{"Valid go version", "module test\n\ngo 1.26\n", "1.26"},
		{"Valid go version with patch", "module test\n\ngo 1.26.1\n", "1.26.1"},
		{"With inline comment", "module test\n\ngo 1.26 // minimum version\n", "1.26"},
		{"With preceding spaces", "module test\n  go 1.25\n", "1.25"},
		{"No version", "module test\n", ""},
		{"Empty file", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, getGoVersionFromMod(tt.content))
		})
	}
}

func TestIsSupportedGoVersion(t *testing.T) {
	tests := []struct {
		name     string
		current  string
		min      string
		expected bool
	}{
		{"Exact match", "1.26", "1.26", true},
		{"Newer patch", "1.26.1", "1.26", true},
		{"Newer minor", "1.27", "1.26", true},
		{"Older version", "1.25", "1.26", false},
		{"Older patch", "1.26.0", "1.26.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isSupportedGoVersion(tt.current, tt.min))
		})
	}
}

func TestConfirmEOF(t *testing.T) {
	// swap global reader
	oldStdinReader := stdinReader
	defer func() { stdinReader = oldStdinReader }()

	r, w, err := os.Pipe()
	assert.NoError(t, err)

	stdinReader = bufio.NewReader(r)

	// close pipe to trigger eof
	_ = w.Close()

	// fallback to default on eof
	assert.True(t, confirm("test", true))
	assert.False(t, confirm("test", false))
}
