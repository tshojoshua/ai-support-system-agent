package agent

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// LogLevel represents log severity
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelFatal LogLevel = "fatal"
)

// Logger provides structured JSON logging
type Logger struct {
	writer   io.Writer
	agentID  string
	minLevel LogLevel
}

// NewLogger creates a new logger
func NewLogger(agentID string, minLevel LogLevel) *Logger {
	return &Logger{
		writer:   os.Stdout,
		agentID:  agentID,
		minLevel: minLevel,
	}
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Component string                 `json:"component"`
	AgentID   string                 `json:"agent_id,omitempty"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

func (l *Logger) log(level LogLevel, component string, fields map[string]interface{}) {
	if !l.shouldLog(level) {
		return
	}

	message := ""
	if msg, ok := fields["message"]; ok {
		message = fmt.Sprint(msg)
		delete(fields, "message")
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     string(level),
		Component: component,
		AgentID:   l.agentID,
		Message:   message,
	}

	if len(fields) > 0 {
		entry.Fields = fields
	}

	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal log entry: %v\n", err)
		return
	}

	fmt.Fprintln(l.writer, string(data))
}

func (l *Logger) shouldLog(level LogLevel) bool {
	levels := []LogLevel{LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError, LogLevelFatal}
	
	minIdx := 0
	levelIdx := 0
	
	for i, lvl := range levels {
		if lvl == l.minLevel {
			minIdx = i
		}
		if lvl == level {
			levelIdx = i
		}
	}
	
	return levelIdx >= minIdx
}

// Debug logs a debug message
func (l *Logger) Debug(component string, fields map[string]interface{}) {
	l.log(LogLevelDebug, component, fields)
}

// Info logs an info message
func (l *Logger) Info(component string, fields map[string]interface{}) {
	l.log(LogLevelInfo, component, fields)
}

// Warn logs a warning message
func (l *Logger) Warn(component string, fields map[string]interface{}) {
	l.log(LogLevelWarn, component, fields)
}

// Error logs an error message
func (l *Logger) Error(component string, fields map[string]interface{}) {
	l.log(LogLevelError, component, fields)
}

// Fatal logs a fatal message
func (l *Logger) Fatal(component string, fields map[string]interface{}) {
	l.log(LogLevelFatal, component, fields)
}
