package logger

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/tech-engine/goscrapy/pkg/core"
)

var bytePool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, 256)
		return &b
	},
}

type logger struct {
	w     *sharedWriter
	level core.LogLevel
	name  string
}

var levelBytes = [...][]byte{
	core.LevelDebug: []byte("DEBUG "),
	core.LevelInfo:  []byte("INFO  "),
	core.LevelWarn:  []byte("WARN  "),
	core.LevelError: []byte("ERROR "),
}

// Allows switching underlying writer
func (l *logger) WithWriter(w io.Writer) core.ILogger {
	l.w.SetWriter(w)
	return l
}

func (l *logger) Debug(v ...any) {
	if l.level > core.LevelDebug {
		return
	}
	l.print(core.LevelDebug, "🔍", v...)
}

func (l *logger) Info(v ...any) {
	if l.level > core.LevelInfo {
		return
	}
	l.print(core.LevelInfo, "🕷️", v...)
}

func (l *logger) Warn(v ...any) {
	if l.level > core.LevelWarn {
		return
	}
	l.print(core.LevelWarn, "⚠️", v...)
}

func (l *logger) Error(v ...any) {
	if l.level > core.LevelError {
		return
	}
	l.print(core.LevelError, "❌", v...)
}

func (l *logger) Debugf(format string, v ...any) {
	if l.level > core.LevelDebug {
		return
	}
	l.printf(core.LevelDebug, "🔍", format, v...)
}

func (l *logger) Infof(format string, v ...any) {
	if l.level > core.LevelInfo {
		return
	}
	l.printf(core.LevelInfo, "🕷️", format, v...)
}

func (l *logger) Warnf(format string, v ...any) {
	if l.level > core.LevelWarn {
		return
	}
	l.printf(core.LevelWarn, "⚠️", format, v...)
}

func (l *logger) Errorf(format string, v ...any) {
	if l.level > core.LevelError {
		return
	}
	l.printf(core.LevelError, "❌", format, v...)
}

func (l *logger) WithName(name string) core.ILogger {
	if l.level == core.LevelNone {
		return l
	}

	newName := name
	if l.name != "" {
		newName = l.name + " > " + name
	}

	return &logger{
		w:     l.w,
		level: l.level,
		name:  newName,
	}
}

func (l *logger) formatHeader(b []byte, level core.LogLevel, emoji string) []byte {
	b = time.Now().AppendFormat(b, "2006/01/02 15:04:05 ")
	b = append(b, emoji...)
	b = append(b, ' ')

	if int(level) < len(levelBytes) {
		b = append(b, levelBytes[level]...)
	}

	if l.name != "" {
		b = append(b, '[')
		b = append(b, l.name...)
		b = append(b, ']', ' ')
	}
	return b
}

func (l *logger) log(level core.LogLevel, emoji, format string, v ...any) {
	bp := bytePool.Get().(*[]byte)
	b := (*bp)[:0]
	defer func() {
		*bp = b
		bytePool.Put(bp)
	}()

	b = l.formatHeader(b, level, emoji)
	if format == "" {
		b = fmt.Append(b, v...)
	} else {
		b = fmt.Appendf(b, format, v...)
	}
	b = append(b, '\n')

	l.w.Write(b)
}

func (l *logger) print(level core.LogLevel, emoji string, v ...any) {
	l.log(level, emoji, "", v...)
}

func (l *logger) printf(level core.LogLevel, emoji, format string, v ...any) {
	l.log(level, emoji, format, v...)
}
