// Package logger provides logging utilities for the application.
package logger

import (
	"log"
	"os"
)

var std = log.New(os.Stdout, "", log.LstdFlags|log.Lmsgprefix)

// SetPrefix sets a prefix used for logs
func SetPrefix(p string) {
	std.SetPrefix(p + " ")
}

// Info logs an informational message
func Info(msg string) {
	std.SetPrefix("[INFO] ")
	std.Println(msg)
}

// Error logs an error message
func Error(msg string) {
	std.SetPrefix("[ERROR] ")
	std.Println(msg)
}
