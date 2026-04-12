package logger

import (
	"io"

	"github.com/tech-engine/goscrapy/pkg/core"
)

type noopLogger struct{}

func (n *noopLogger) Debug(args ...any)                   {}
func (n *noopLogger) Info(args ...any)                    {}
func (n *noopLogger) Warn(args ...any)                    {}
func (n *noopLogger) Error(args ...any)                   {}
func (n *noopLogger) Debugf(fmt string, args ...any)      {}
func (n *noopLogger) Infof(fmt string, args ...any)       {}
func (n *noopLogger) Warnf(fmt string, args ...any)       {}
func (n *noopLogger) Errorf(fmt string, args ...any)      {}
func (n *noopLogger) WithName(name string) core.ILogger   { return n }
func (n *noopLogger) WithWriter(w io.Writer) core.ILogger { return n }

func NewNoopLogger() core.IConfigurableLogger {
	return &noopLogger{}
}

type noopWriter struct{}

func (nw noopWriter) Write(p []byte) (int, error) {
	return len(p), nil
}
