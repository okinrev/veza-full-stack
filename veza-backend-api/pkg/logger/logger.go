package logger

import (
    "log"
    "os"
)

type Logger struct {
    *log.Logger
}

// New creates a new logger instance
func New(environment string) *Logger {
    logger := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
    
    if environment == "development" {
        logger.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
    }
    
    return &Logger{Logger: logger}
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
    l.Printf("[INFO] "+msg, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...interface{}) {
    l.Printf("[ERROR] "+msg, args...)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...interface{}) {
    l.Printf("[DEBUG] "+msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
    l.Printf("[WARN] "+msg, args...)
}