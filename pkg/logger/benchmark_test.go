// Note: generated benchmarks
package logger

import (
	"io"
	"strings"
	"testing"

	"github.com/tech-engine/goscrapy/pkg/core"
)

var levelMap = map[string]core.LogLevel{
	"DEBUG": core.LevelDebug,
	"INFO":  core.LevelInfo,
	"WARN":  core.LevelWarn,
	"ERROR": core.LevelError,
	"NONE":  core.LevelNone,
}

func BenchmarkLoggerInfo(b *testing.B) {
	l := NewLogger(WithLevel(core.LevelInfo), WithWriter(io.Discard))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("benchmark message")
	}
}

func BenchmarkLoggerParallel(b *testing.B) {
	l := NewLogger(WithLevel(core.LevelInfo), WithWriter(io.Discard))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info("benchmark parallel message")
		}
	})
}

func BenchmarkLoggerFormattedParallel(b *testing.B) {
	l := NewLogger(WithLevel(core.LevelInfo), WithWriter(io.Discard))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Infof("benchmark parallel formatted message: %s=%d", "key", 123)
		}
	})
}

func BenchmarkLoggerDisabled(b *testing.B) {
	l := NewLogger(WithLevel(core.LevelNone))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("benchmark message")
	}
}

func BenchmarkNewLogger(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewLogger(WithLevel(core.LevelInfo), WithWriter(io.Discard))
	}
}

func BenchmarkNewLoggerDisabled(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewLogger(WithLevel(core.LevelNone))
	}
}

func BenchmarkLogger_Current(b *testing.B) {
	w := &sharedWriter{}
	w.SetWriter(noopWriter{})

	l := &logger{
		w:     w,
		level: 0,
		name:  "engine",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.Info("user", i, "logged in")
	}
}

func BenchmarkParseLevelSwitch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ParseLevel("DEBUG")
		_ = ParseLevel("INFO")
		_ = ParseLevel("WARN")
		_ = ParseLevel("ERROR")
		_ = ParseLevel("NONE")
	}
}

func BenchmarkParseLevelMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = levelMap[strings.ToUpper("DEBUG")]
		_ = levelMap[strings.ToUpper("INFO")]
		_ = levelMap[strings.ToUpper("WARN")]
		_ = levelMap[strings.ToUpper("ERROR")]
		_ = levelMap[strings.ToUpper("NONE")]
	}
}
