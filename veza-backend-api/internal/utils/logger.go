package utils

import (
	"fmt"
	"log"
	"os"
	"time"
)

var (
	infoLogger  *log.Logger
	errorLogger *log.Logger
)

func init() {
	// Configuration des loggers
	infoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// LogInfo enregistre un message d'information
func LogInfo(message string) {
	infoLogger.Output(2, fmt.Sprintf("[%s] %s", time.Now().Format("2006-01-02 15:04:05"), message))
}

// LogError enregistre un message d'erreur
func LogError(message string) {
	errorLogger.Output(2, fmt.Sprintf("[%s] %s", time.Now().Format("2006-01-02 15:04:05"), message))
}

// LogFatal enregistre un message fatal et termine le programme
func LogFatal(message string) {
	errorLogger.Output(2, fmt.Sprintf("[%s] FATAL: %s", time.Now().Format("2006-01-02 15:04:05"), message))
	os.Exit(1)
}

// LogDebug enregistre un message de débogage (si le mode debug est activé)
func LogDebug(message string) {
	if os.Getenv("DEBUG") == "true" {
		infoLogger.Output(2, fmt.Sprintf("[%s] DEBUG: %s", time.Now().Format("2006-01-02 15:04:05"), message))
	}
} 