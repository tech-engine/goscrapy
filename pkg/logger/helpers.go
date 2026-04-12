package logger

import (
	"strings"

	"github.com/tech-engine/goscrapy/pkg/core"
)

func ParseLevel(s string) core.LogLevel {
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
