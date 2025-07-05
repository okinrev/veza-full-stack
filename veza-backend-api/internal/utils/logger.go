package utils

import (
	"fmt"
	"log"
	"os"
	"time"
)

var (
	infoLogger  = log.New(os.Stdout, "", log.LstdFlags)
	errorLogger = log.New(os.Stderr, "", log.LstdFlags)
	debugMode   = os.Getenv("DEBUG") == "true"
)

func init() {
	// Configuration des loggers
	infoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// LogInfo enregistre un message d'information
func LogInfo(message string) {
	if err := infoLogger.Output(2, fmt.Sprintf("[%s] %s", time.Now().Format("2006-01-02 15:04:05"), message)); err != nil {
		// Fallback to stderr if logging fails
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to log info: %v\n", err)
	}
}

// LogError enregistre un message d'erreur
func LogError(message string) {
	if err := errorLogger.Output(2, fmt.Sprintf("[%s] %s", time.Now().Format("2006-01-02 15:04:05"), message)); err != nil {
		// Fallback to stderr if logging fails
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to log error: %v\n", err)
	}
}

// LogFatal enregistre un message fatal et termine le programme
func LogFatal(message string) {
	if err := errorLogger.Output(2, fmt.Sprintf("[%s] FATAL: %s", time.Now().Format("2006-01-02 15:04:05"), message)); err != nil {
		// Fallback to stderr if logging fails
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to log fatal: %v\n", err)
	}
	os.Exit(1)
}

// LogDebug enregistre un message de débogage (si le mode debug est activé)
func LogDebug(message string) {
	if debugMode {
		if err := infoLogger.Output(2, fmt.Sprintf("[%s] DEBUG: %s", time.Now().Format("2006-01-02 15:04:05"), message)); err != nil {
			// Fallback to stderr if logging fails
			fmt.Fprintf(os.Stderr, "[ERROR] Failed to log debug: %v\n", err)
		}
	}
}
