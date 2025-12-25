package logging

import (
	"log"
	"os"
)

func InitLogging(logPath string) {
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Failed to open log file, using stderr:", err)
		return
	}
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Logging initialized")
}
