package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

var logFile *os.File

func Init() error {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		return fmt.Errorf("error creating logs directory: %v", err)
	}

	// Create log file with timestamp
	filename := fmt.Sprintf("logs/api_%s.log", time.Now().Format("2006-01-02"))
	var err error
	logFile, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error creating log file: %v", err)
	}

	log.SetOutput(logFile)
	return nil
}

func Close() {
	if logFile != nil {
		logFile.Close()
	}
}

func LogAPIRequest(method, url string, body []byte) {
	log.Printf("[%s] REQUEST: %s %s\nBody: %s\n", time.Now().Format("2006-01-02 15:04:05"), method, url, string(body))
}

func LogAPIResponse(statusCode int, body []byte) {
	log.Printf("[%s] RESPONSE: Status %d\nBody: %s\n", time.Now().Format("2006-01-02 15:04:05"), statusCode, string(body))
}

func LogAPIError(err error) {
	log.Printf("[%s] ERROR: %v\n", time.Now().Format("2006-01-02 15:04:05"), err)
}
