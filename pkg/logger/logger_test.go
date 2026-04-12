// Note: generated tests
package logger

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tech-engine/goscrapy/pkg/core"
)

func TestNewLogger(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(
		WithLevel(core.LevelDebug),
		WithWriter(&buf),
		WithName("test"),
	)

	l.Debug("debug message")
	l.Info("info message")

	output := buf.String()
	assert.Contains(t, output, "DEBUG")
	assert.Contains(t, output, "[test]")
	assert.Contains(t, output, "debug message")
	assert.Contains(t, output, "info message")
}

func TestLoggerDefaults(t *testing.T) {
	l := NewLogger()
	assert.Equal(t, core.LevelInfo, l.level)
	assert.Equal(t, os.Stderr, l.w.w)
}

func TestLoggerLevelNone(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(
		WithLevel(core.LevelNone),
		WithWriter(&buf),
	)

	l.Info("this should be discarded")
	l.Error("this should also be discarded")

	assert.Empty(t, buf.String(), "Output should be empty for LevelNone")

	assert.Equal(t, noopWriter{}, l.w.w, "Writer should be noopWriter for LevelNone")
}

func TestLoggerLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(
		WithLevel(core.LevelWarn),
		WithWriter(&buf),
	)

	l.Info("should not see this")
	l.Warn("should see this")

	output := buf.String()
	assert.NotContains(t, output, "should not see this")
	assert.Contains(t, output, "should see this")
}

func TestLoggerWithWriter(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	l := NewLogger(WithWriter(&buf1), WithLevel(core.LevelInfo))

	l.Info("message 1")
	assert.Contains(t, buf1.String(), "message 1")

	l.WithWriter(&buf2)
	l.Info("message 2")
	assert.Contains(t, buf2.String(), "message 2")
}

func TestLoggerEnvVar(t *testing.T) {
	os.Setenv("GOSCRAPY_LOG_LEVEL", "DEBUG")
	defer os.Unsetenv("GOSCRAPY_LOG_LEVEL")

	var buf bytes.Buffer
	l := NewLogger(WithWriter(&buf))

	l.Debug("env debug")

	assert.Contains(t, buf.String(), "env debug")
}

func TestLoggerWithName(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(WithWriter(&buf), WithLevel(core.LevelInfo))
	l2 := l.WithName("sub")

	l2.Info("sub message")

	assert.Contains(t, buf.String(), "[sub]")
	assert.Contains(t, buf.String(), "sub message")
}
