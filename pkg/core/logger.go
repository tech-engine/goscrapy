package core

import "io"

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelNone
)

type ILogger interface {
	Debug(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)
	Debugf(template string, args ...any)
	Infof(template string, args ...any)
	Warnf(template string, args ...any)
	Errorf(template string, args ...any)
	// must return a logger pointing to the same writer as that of parent
	WithName(name string) ILogger
}

// IConfigurableLogger is the framework-level interface that allows output redirection.
type IConfigurableLogger interface {
	ILogger
	WithWriter(io.Writer) ILogger
}
