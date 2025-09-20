package logger

import (
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestInitialize(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		wantError bool
	}{
		{
			name:      "valid debug level",
			level:     "debug",
			wantError: false,
		},
		{
			name:      "valid info level",
			level:     "info",
			wantError: false,
		},
		{
			name:      "valid warn level",
			level:     "warn",
			wantError: false,
		},
		{
			name:      "valid error level",
			level:     "error",
			wantError: false,
		},
		{
			name:      "invalid level",
			level:     "invalid",
			wantError: true,
		},
		{
			name:      "empty level",
			level:     "",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Initialize(tt.level)
			if (err != nil) != tt.wantError {
				t.Errorf("Initialize() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestAsyncInfo(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	Log = zap.New(core)

	AsyncInfo("test message", zap.String("key", "value"))

	time.Sleep(10 * time.Millisecond)

	logs := recorded.All()
	if len(logs) == 0 {
		t.Error("Expected log message to be recorded")
		return
	}

	log := logs[0]
	if log.Message != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", log.Message)
	}

	if log.Level != zapcore.InfoLevel {
		t.Errorf("Expected level Info, got %v", log.Level)
	}
}

func TestAsyncWarn(t *testing.T) {
	core, recorded := observer.New(zapcore.WarnLevel)
	Log = zap.New(core)

	AsyncWarn("warning message", zap.String("key", "value"))

	time.Sleep(10 * time.Millisecond)

	logs := recorded.All()
	if len(logs) == 0 {
		t.Error("Expected log message to be recorded")
		return
	}

	log := logs[0]
	if log.Message != "warning message" {
		t.Errorf("Expected message 'warning message', got '%s'", log.Message)
	}

	if log.Level != zapcore.WarnLevel {
		t.Errorf("Expected level Warn, got %v", log.Level)
	}
}

func TestAsyncError(t *testing.T) {
	core, recorded := observer.New(zapcore.ErrorLevel)
	Log = zap.New(core)

	AsyncError("error message", zap.String("key", "value"))

	time.Sleep(10 * time.Millisecond)

	logs := recorded.All()
	if len(logs) == 0 {
		t.Error("Expected log message to be recorded")
		return
	}

	log := logs[0]
	if log.Message != "error message" {
		t.Errorf("Expected message 'error message', got '%s'", log.Message)
	}

	if log.Level != zapcore.ErrorLevel {
		t.Errorf("Expected level Error, got %v", log.Level)
	}
}

func TestLogGlobalInstance(t *testing.T) {
	if Log == nil {
		t.Error("Expected Log to be initialized")
	}

	if Log.Core().Enabled(zapcore.DebugLevel) {
		t.Error("Expected no-op logger to not be enabled")
	}
}
