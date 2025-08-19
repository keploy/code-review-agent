package utils

import (
	"fmt"
	"log"
	"os"

	"go.uber.org/zap"
)

type Logger struct {
	*zap.SugaredLogger
}

func NewLogger() *Logger {
	// Using NewDevelopment for human-readable logs, which is similar to the previous logger.
	// For a production environment, you might switch to zap.NewProduction() which logs in JSON.
	logger, err := zap.NewDevelopment()
	if err != nil {
		// Fallback to standard logger if zap fails to initialize
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	return &Logger{
		SugaredLogger: logger.Sugar(),
	}
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.Infof(format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.Errorf(format, v...)
}

func (l *Logger) GitHubOutput(name, value string) {
	// Use the modern, recommended way to set outputs
	if githubOutput := os.Getenv("GITHUB_OUTPUT"); githubOutput != "" {
		f, err := os.OpenFile(githubOutput, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			defer f.Close()
			if _, err := f.WriteString(fmt.Sprintf("%s=%s\n", name, value)); err == nil {
				return
			}
		}
	}
	// Fallback for local testing or older environments
	fmt.Printf("::set-output name=%s::%s\n", name, value)
}
