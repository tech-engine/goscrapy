package logger

import (
	"io"
	"os"
	"sync"

	"github.com/tech-engine/goscrapy/internal/types"
	"github.com/tech-engine/goscrapy/pkg/core"
)

// This will reroute logs from different goroutines to the same writer
// may improve later if needed
type sharedWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (s *sharedWriter) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.w.Write(p)
}

func (s *sharedWriter) SetWriter(w io.Writer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.w = w
}

type loggerOpts struct {
	level core.LogLevel
	w     io.Writer
	name  string
}

func WithLevel(level core.LogLevel) types.OptFunc[loggerOpts] {
	return func(o *loggerOpts) {
		o.level = level
	}
}

func WithWriter(w io.Writer) types.OptFunc[loggerOpts] {
	return func(o *loggerOpts) {
		o.w = w
	}
}

func WithName(name string) types.OptFunc[loggerOpts] {
	return func(o *loggerOpts) {
		o.name = name
	}
}

func defaultOpts() loggerOpts {
	opts := loggerOpts{
		level: core.LevelInfo,
		w:     os.Stderr,
	}

	value, ok := os.LookupEnv("GOSCRAPY_LOG_LEVEL")
	if ok && value != "" {
		opts.level = ParseLevel(value)
	}

	return opts
}

func NewLogger(optFuncs ...types.OptFunc[loggerOpts]) *logger {
	opts := defaultOpts()

	for _, fn := range optFuncs {
		fn(&opts)
	}

	if opts.level == core.LevelNone {
		opts.w = noopWriter{}
	}

	return &logger{
		w:     &sharedWriter{w: opts.w},
		level: opts.level,
		name:  opts.name,
	}
}

func EnsureLogger(loggerIn core.ILogger) core.ILogger {
	if loggerIn != nil {
		return loggerIn
	}
	// we fallback to default logger
	return NewLogger()
}
