package tui

import (
	"strings"
	"sync"
)

// represents a single line in the TUI log view
type LogEntry struct {
	Level string
	Msg   string
}

// maintains a sliding window of recent logs
type LogBuffer struct {
	mu      sync.RWMutex
	logs    []string
	maxSize int
}

func NewLogBuffer(maxSize int) *LogBuffer {
	return &LogBuffer{
		logs:    make([]string, 0, maxSize),
		maxSize: maxSize,
	}
}

func (lb *LogBuffer) Write(p []byte) (n int, err error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	// we split by newline and add each as a separate entry
	lines := strings.Split(string(p), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if len(lb.logs) >= lb.maxSize {
			lb.logs = lb.logs[1:]
		}
		lb.logs = append(lb.logs, trimmed)
	}

	return len(p), nil
}

func (lb *LogBuffer) GetLogs() []string {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	copyLogs := make([]string, len(lb.logs))
	copy(copyLogs, lb.logs)
	return copyLogs
}

// handles log capturing
type LogWriter struct {
	buffer *LogBuffer
}

func NewLogWriter(buffer *LogBuffer) *LogWriter {
	return &LogWriter{buffer: buffer}
}

func (cw *LogWriter) Write(p []byte) (n int, err error) {
	return cw.buffer.Write(p)
}
