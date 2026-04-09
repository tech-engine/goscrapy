package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync/atomic"

	"github.com/tech-engine/goscrapy/pkg/core"
)

var (
	globalLevel  atomic.Uint32
	activeLogger atomic.Pointer[core.ILogger]
)

func init() {
	// defaults to Info
	globalLevel.Store(uint32(core.LevelInfo))
	
	dl := &defaultLogger{
		logger: log.New(os.Stderr, "", 0), // No timestamps by default for a clean cli look
	}
	var logger core.ILogger = dl
	activeLogger.Store(&logger)

	// pickup level from environment
	if lv := os.Getenv("GOSCRAPY_LOG_LEVEL"); lv != "" {
		SetLevel(parseLevel(lv))
	}
}

// Allows users to plug in a custom logger
func SetLogger(l core.ILogger) {
	activeLogger.Store(&l)
}

// Sets the global log level
func SetLevel(l core.LogLevel) {
	globalLevel.Store(uint32(l))
}

func GetLevel() core.LogLevel {
	return core.LogLevel(globalLevel.Load())
}

// Returns the current active global logger
func GetLogger() core.ILogger {
	return *activeLogger.Load()
}

func Info(v ...any) {
	if GetLevel() > core.LevelInfo {
		return
	}
	(*activeLogger.Load()).Info(v...)
}

func Debug(v ...any) {
	if GetLevel() > core.LevelDebug {
		return
	}
	(*activeLogger.Load()).Debug(v...)
}

func Warn(v ...any) {
	if GetLevel() > core.LevelWarn {
		return
	}
	(*activeLogger.Load()).Warn(v...)
}

func Error(v ...any) {
	if GetLevel() > core.LevelError {
		return
	}
	(*activeLogger.Load()).Error(v...)
}

func Infof(format string, v ...any) {
	if GetLevel() > core.LevelInfo {
		return
	}
	(*activeLogger.Load()).Infof(format, v...)
}

func Debugf(format string, v ...any) {
	if GetLevel() > core.LevelDebug {
		return
	}
	(*activeLogger.Load()).Debugf(format, v...)
}

func Warnf(format string, v ...any) {
	if GetLevel() > core.LevelWarn {
		return
	}
	(*activeLogger.Load()).Warnf(format, v...)
}

func Errorf(format string, v ...any) {
	if GetLevel() > core.LevelError {
		return
	}
	(*activeLogger.Load()).Errorf(format, v...)
}

type defaultLogger struct {
	logger *log.Logger
	name string
}

func (d *defaultLogger) Debug(v ...any)            { d.print("DEBUG", "🔍", v...) }
func (d *defaultLogger) Info(v ...any)             { d.print("INFO", "🕷️", v...) }
func (d *defaultLogger) Warn(v ...any)             { d.print("WARN", "⚠️", v...) }
func (d *defaultLogger) Error(v ...any)            { d.print("ERROR", "❌", v...) }
func (d *defaultLogger) Debugf(f string, v ...any) { d.printf("DEBUG", "🔍", f, v...) }
func (d *defaultLogger) Infof(f string, v ...any)  { d.printf("INFO", "🕷️", f, v...) }
func (d *defaultLogger) Warnf(f string, v ...any)  { d.printf("WARN", "⚠️", f, v...) }
func (d *defaultLogger) Errorf(f string, v ...any) { d.printf("ERROR", "❌", f, v...) }

func (d *defaultLogger) WithName(name string) core.ILogger {
	return &defaultLogger{
		logger:    d.logger,
		name: name,
	}
}

func (d *defaultLogger) print(level, emoji string, v ...any) {
	msg := fmt.Sprint(v...)
	if d.name != "" {
		d.logger.Printf("%s %-5s [%s] %s", emoji, level, d.name, msg)
	} else {
		d.logger.Printf("%s %-5s %s", emoji, level, msg)
	}
}

func (d *defaultLogger) printf(level, emoji, f string, v ...any) {
	msg := fmt.Sprintf(f, v...)
	if d.name != "" {
		d.logger.Printf("%s %-5s [%s] %s", emoji, level, d.name, msg)
	} else {
		d.logger.Printf("%s %-5s %s", emoji, level, msg)
	}
}

func parseLevel(s string) core.LogLevel {
	switch strings.ToUpper(s) {
	case "DEBUG":
		return core.LevelDebug
	case "INFO":
		return core.LevelInfo
	case "WARN":
		return core.LevelWarn
	case "ERROR":
		return core.LevelError
	case "NONE":
		return core.LevelNone
	default:
		return core.LevelInfo
	}
}
