package services

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type LogWriter struct {
	stdout *os.File
	file   *os.File
}

func (t *LogWriter) Write(p []byte) (n int, err error) {
	n, err = t.stdout.Write(p)
	if err != nil {
		return n, err
	}
	n, err = t.file.Write(p)
	return n, err
}

func NewLogger() (*slog.Logger, error) {
	logDir := os.Getenv("LOG_DIR")

	if err := os.MkdirAll(logDir, 0750); err != nil {
		return nil, err
	}

	currentTime := time.Now().Format("2006-01-02")
	logFileName := fmt.Sprintf("%s/logs_%s.log", logDir, currentTime)

	if !strings.HasPrefix(logFileName, logDir+"/") {
		return nil, fmt.Errorf("invalid log file path")
	}

	file, err := os.OpenFile(filepath.Clean(logFileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}

	writer := &LogWriter{
		stdout: os.Stdout,
		file:   file,
	}

	h := slog.NewJSONHandler(writer, nil)
	logger := slog.New(h)

	return logger, nil
}
