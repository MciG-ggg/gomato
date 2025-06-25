package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	logger  *log.Logger
	logFile *os.File
)

// Init sets up the logger to write to a file.
func Init() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	logDir := filepath.Join(homeDir, ".gomato")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	logFilePath := filepath.Join(logDir, "gomato.log")
	logFile, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	logger = log.New(logFile, "", 0) // No prefix or flags, we'll format it ourselves
	return nil
}

// Log writes a message to the log file with a timestamp.
func Log(message string) {
	if logger == nil {
		fmt.Println("Logger not initialized. Please call logging.Init() first.")
		return
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logger.Printf("[%s] %s", timestamp, message)
}

// Close closes the log file.
func Close() {
	if logFile != nil {
		logFile.Close()
	}
}
