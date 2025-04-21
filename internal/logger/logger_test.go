package logger

import (
	"testing"
)

func TestSetup(t *testing.T) {
	logger := Setup()

	// Проверяем, что возвращается не nil
	if logger == nil {
		t.Error("Expected logger to be not nil, got nil")
	}

	// Проверяем, что handler установлен (косвенная проверка)
	handler := logger.Handler()
	if handler == nil {
		t.Error("Expected handler to be not nil, got nil")
	}
}
