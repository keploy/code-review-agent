package utils

import (
	"fmt"
	"log"
	"os"
)

type Logger struct {
    *log.Logger
}

func NewLogger() *Logger {
    return &Logger{
        Logger: log.New(os.Stdout, "", log.LstdFlags),
    }
}

func (l *Logger) Info(format string, v ...interface{}) {
    l.Printf("INFO: "+format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
    l.Printf("ERROR: "+format, v...)
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
