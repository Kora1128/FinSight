package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Logger is the application's logger instance
var Logger *logrus.Logger

// Init initializes the logger
func Init() {
	Logger = logrus.New()

	// Set output to stdout
	Logger.SetOutput(os.Stdout)

	// Set log level based on environment
	if os.Getenv("APP_ENV") == "production" {
		Logger.SetLevel(logrus.InfoLevel)
	} else {
		Logger.SetLevel(logrus.DebugLevel)
	}

	// Set JSON formatter for production
	if os.Getenv("APP_ENV") == "production" {
		Logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		Logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}
}

// Debug logs a debug message
func Debug(args ...interface{}) {
	Logger.Debug(args...)
}

// Debugf logs a formatted debug message
func Debugf(format string, args ...interface{}) {
	Logger.Debugf(format, args...)
}

// Info logs an info message
func Info(args ...interface{}) {
	Logger.Info(args...)
}

// Infof logs a formatted info message
func Infof(format string, args ...interface{}) {
	Logger.Infof(format, args...)
}

// Warn logs a warning message
func Warn(args ...interface{}) {
	Logger.Warn(args...)
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...interface{}) {
	Logger.Warnf(format, args...)
}

// Error logs an error message
func Error(args ...interface{}) {
	Logger.Error(args...)
}

// Errorf logs a formatted error message
func Errorf(format string, args ...interface{}) {
	Logger.Errorf(format, args...)
}

// Fatal logs a fatal message and exits
func Fatal(args ...interface{}) {
	Logger.Fatal(args...)
}

// Fatalf logs a formatted fatal message and exits
func Fatalf(format string, args ...interface{}) {
	Logger.Fatalf(format, args...)
}
