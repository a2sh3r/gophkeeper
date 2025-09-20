// Package logger provides a global logger instance and initialization logic.
package logger

import "go.uber.org/zap"

// Log is the global logger instance.
var Log *zap.Logger = zap.NewNop()

// Initialize sets up the global logger with the specified log level.
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewDevelopmentConfig()

	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = zl

	return nil
}

// AsyncInfo logs an info message asynchronously.
func AsyncInfo(msg string, fields ...zap.Field) {
	go func() {
		Log.Info(msg, fields...)
	}()
}

// AsyncWarn logs a warning message asynchronously.
func AsyncWarn(msg string, fields ...zap.Field) {
	go func() {
		Log.Warn(msg, fields...)
	}()
}

// AsyncError logs an error message asynchronously.
func AsyncError(msg string, fields ...zap.Field) {
	go func() {
		Log.Error(msg, fields...)
	}()
}
