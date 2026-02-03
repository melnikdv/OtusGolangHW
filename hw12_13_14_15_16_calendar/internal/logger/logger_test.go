package logger

import (
	"bytes"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNew_ValidLevel(t *testing.T) {
	log := New("debug")
	assert.Equal(t, logrus.DebugLevel, log.GetLevel())
}

func TestNew_InvalidLevel(t *testing.T) {
	t.Parallel()
	log := New("invalid-level")
	assert.Equal(t, logrus.InfoLevel, log.GetLevel()) // fallback to info
}

func TestNew_EmptyLevel(t *testing.T) {
	t.Parallel()
	log := New("")
	assert.Equal(t, logrus.InfoLevel, log.GetLevel())
}

func TestLogger_LogMethods(t *testing.T) {
	t.Parallel()
	log := New("debug")

	// Просто убедимся, что не падает
	log.Debug("debug message")
	log.Info("info message")
	log.Warn("warn message")
	log.Error("error message")
}

func TestLogger_Output(t *testing.T) {
	var buf bytes.Buffer
	log := logrus.New()
	log.SetOutput(&buf)
	log.SetLevel(logrus.InfoLevel)
	log.Formatter = &logrus.TextFormatter{
		DisableColors:    true,
		DisableTimestamp: true,
	}

	log.Info("test message")

	output := buf.String()
	assert.Contains(t, output, "level=info")
	assert.Contains(t, output, "msg=\"test message\"")
}
