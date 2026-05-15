// File logging.go provides level-aware logger construction and helpers.
package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// Level is the configured log verbosity threshold.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// Logger provides leveled logging for runtime components.
type Logger struct {
	base  *log.Logger
	level Level
}

// New returns a project logger configured with the given level.
func New(level string) *Logger {
	parsed := parseLevel(level)
	return &Logger{
		base:  log.New(io.MultiWriter(os.Stdout), "socks2proxy ", log.LstdFlags|log.Lmicroseconds),
		level: parsed,
	}
}

// Debugf logs diagnostic details useful for troubleshooting.
func (l *Logger) Debugf(format string, args ...any) {
	l.logf(LevelDebug, "DEBUG", format, args...)
}

// Infof logs normal operational messages.
func (l *Logger) Infof(format string, args ...any) {
	l.logf(LevelInfo, "INFO", format, args...)
}

// Warnf logs non-fatal anomalous conditions.
func (l *Logger) Warnf(format string, args ...any) {
	l.logf(LevelWarn, "WARN", format, args...)
}

// Errorf logs errors that prevented handling a request or operation.
func (l *Logger) Errorf(format string, args ...any) {
	l.logf(LevelError, "ERROR", format, args...)
}

// Fatalf logs an error and exits the process.
func (l *Logger) Fatalf(format string, args ...any) {
	l.base.Fatalf("[FATAL] %s", fmt.Sprintf(format, args...))
}

// Printf aliases Infof for compatibility with call sites that use Printf.
func (l *Logger) Printf(format string, args ...any) {
	l.Infof(format, args...)
}

func (l *Logger) logf(at Level, label string, format string, args ...any) {
	if at < l.level {
		return
	}
	l.base.Printf("[%s] %s", label, fmt.Sprintf(format, args...))
}

func parseLevel(raw string) Level {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return LevelDebug
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	case "info":
		fallthrough
	default:
		return LevelInfo
	}
}
