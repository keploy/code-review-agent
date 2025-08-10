package utils

import (
	"fmt"
	"log"
	"os"
	//"strings"
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

func (l *Logger) Debug(format string, v ...interface{}) {
    l.Printf("DEBUG: "+format, v...)
}

func (l *Logger) GitHubOutput(name, value string) {
    fmt.Printf("::set-output name=%s::%s\n", name, value)
}
