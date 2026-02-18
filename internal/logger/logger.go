// Package logger provides structured logging functionality.
package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	instance *logrus.Logger
	once     sync.Once
)

// Fields is a type alias for logrus.Fields.
type Fields map[string]interface{}

// Config holds logger configuration.
type Config struct {
	Level      string
	ToFile     bool
	FilePath   string
	JSONFormat bool
	LogDir     string
}

// Initialize initializes the global logger instance.
func Initialize(cfg Config) error {
	var initErr error
	once.Do(func() {
		instance = logrus.New()

		// Set log level
		level, err := logrus.ParseLevel(cfg.Level)
		if err != nil {
			level = logrus.InfoLevel
		}
		instance.SetLevel(level)

		// Set formatter
		if cfg.JSONFormat {
			instance.SetFormatter(&logrus.JSONFormatter{
				TimestampFormat: "2006-01-02 15:04:05",
			})
		} else {
			instance.SetFormatter(&logrus.TextFormatter{
				FullTimestamp:   true,
				TimestampFormat: "2006-01-02 15:04:05",
				CallerPrettyfier: func(f *runtime.Frame) (string, string) {
					filename := filepath.Base(f.File)
					return "", fmt.Sprintf("[%s:%d]", filename, f.Line)
				},
			})
		}

		// Set output
		if cfg.ToFile {
			logDir := cfg.LogDir
			if logDir == "" && cfg.FilePath != "" {
				logDir = filepath.Dir(cfg.FilePath)
			}
			if logDir == "" {
				logDir = "logs"
			}
			if err := setupFileOutput(logDir); err != nil {
				initErr = fmt.Errorf("failed to setup file logging: %w", err)
				return
			}
		} else {
			instance.SetOutput(os.Stdout)
		}

		instance.SetReportCaller(true)
	})

	return initErr
}

// setupFileOutput configures file logging with timestamp-based filename.
func setupFileOutput(logDir string) error {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Generate filename with timestamp: bot_20060102_150405.log
	timestamp := time.Now().Format("20060102_150405")
	filePath := filepath.Join(logDir, fmt.Sprintf("%s.log", timestamp))

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	instance.SetOutput(io.MultiWriter(os.Stdout, file))
	return nil
}

// Get returns the logger instance.
func Get() *logrus.Logger {
	if instance == nil {
		_ = Initialize(Config{Level: "info"})
	}
	return instance
}

// WithFields creates a log entry with fields.
func WithFields(fields map[string]interface{}) *logrus.Entry {
	return Get().WithFields(logrus.Fields(fields))
}

// WithError creates a log entry with an error.
func WithError(err error) *logrus.Entry {
	return Get().WithError(err)
}

// WithUser creates a log entry with user context.
func WithUser(userID int64, username string) *logrus.Entry {
	return Get().WithFields(logrus.Fields{
		"user_id":  userID,
		"username": username,
	})
}

// Debug logs a debug message.
func Debug(msg string, args ...interface{}) {
	Get().Debugf(msg, args...)
}

// Info logs an info message.
func Info(msg string, args ...interface{}) {
	Get().Infof(msg, args...)
}

// Warn logs a warning message.
func Warn(msg string, args ...interface{}) {
	Get().Warnf(msg, args...)
}

// Error logs an error message.
func Error(msg string, args ...interface{}) {
	Get().Errorf(msg, args...)
}

// Fatal logs a fatal message and exits.
func Fatal(msg string, args ...interface{}) {
	Get().Fatalf(msg, args...)
}

// LogError logs an error and returns a formatted error.
func LogError(err error, msg string, args ...interface{}) error {
	formattedMsg := fmt.Sprintf(msg, args...)
	Get().WithError(err).Error(formattedMsg)
	return fmt.Errorf("%s: %w", formattedMsg, err)
}

// LogUserAction logs a user action.
func LogUserAction(userID int64, username, action string, details map[string]interface{}) {
	entry := Get().WithFields(logrus.Fields{
		"user_id":  userID,
		"username": username,
		"action":   action,
	})

	if details != nil {
		entry = entry.WithFields(logrus.Fields(details))
	}

	entry.Info("user action")
}

// LogUpload logs an upload attempt.
func LogUpload(userID int64, filename string, fileSize int64, success bool, err error) {
	entry := Get().WithFields(logrus.Fields{
		"user_id":   userID,
		"filename":  filename,
		"file_size": fileSize,
		"success":   success,
	})

	if err != nil {
		entry.WithError(err).Error("upload failed")
	} else {
		entry.Info("upload successful")
	}
}

// LogAPICall logs an API call.
func LogAPICall(api, method, url string, statusCode int, durationMs int64, err error) {
	entry := Get().WithFields(logrus.Fields{
		"api":         api,
		"method":      method,
		"url":         sanitizeURL(url),
		"status_code": statusCode,
		"duration_ms": durationMs,
	})

	if err != nil {
		entry.WithError(err).Error("api call failed")
	} else {
		entry.Debug("api call completed")
	}
}

// sanitizeURL removes sensitive query parameters from URLs.
func sanitizeURL(url string) string {
	if idx := strings.Index(url, "?"); idx != -1 {
		return url[:idx] + "?[REDACTED]"
	}
	return url
}
