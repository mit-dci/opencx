// Package logging provides utilities to easily set a universal log level, as well as intuitively set a log file.
package logging

// Log Levels:
// 3: DebugLevel prints Panics, Fatals, Errors, Warnings, Infos and Debugs
// 2: InfoLevel  prints Panics, Fatals, Errors, Warnings and Info
// 1: WarnLevel  prints Panics, Fatals, Errors and Warnings
// 0: ErrorLevel prints Panics, Fatals and Errors
// Default is level 0
// Code for tagging logs:
// Debug -> Useful debugging information
// Info  -> Something noteworthy happened
// Warn  -> You should probably take a look at this
// Error -> Something failed but I'm not quitting
// Fatal -> Bye

import (
	"fmt"
	"io"
	"log"
	"os"
)

// LogLevel indicates what to log based on the method called
type LogLevel int

const (
	// LogLevelError is the log level that prints Panics, Fatals, and Errors.
	LogLevelError LogLevel = 0
	// LogLevelWarning is the log level that prints Panics, Fatals, Errors, and Warnings.
	LogLevelWarning LogLevel = 1
	// LogLevelInfo is the log level that prints Panics, Fatals, Errors, Warnings, and Infos.
	LogLevelInfo LogLevel = 2
	// LogLevelDebug is the log level that prints Panics, Fatals, Errors, Warnings, Infos, and Debugs.
	LogLevelDebug LogLevel = 3
)

var logLevel = LogLevelError // the default

// SetLogLevel sets the global log level
func SetLogLevel(newLevel int) {
	logLevel = LogLevel(newLevel)
}

// SetLogFile sets a file to write to in addition to standard output.
func SetLogFile(logFile io.Writer) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	logOutput := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(logOutput)
}

func getPrefix(level string) string {
	return fmt.Sprintf("[%s]", level)
}

// Fatalln prints a message and a new line, then calls os.Exit(1).
func Fatalln(args ...interface{}) {
	log.Fatalln(args...)
}

// Fatalf prints a message with a formatting directive, then calls os.Exit(1).
func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

// Fatal prints a message, then calls os.Exit(1).
func Fatal(args ...interface{}) {
	log.Fatal(args...)
}

// Debugf prints debug logs with a formatting directive.
func Debugf(format string, args ...interface{}) {
	if logLevel >= LogLevelDebug {
		log.Printf(fmt.Sprintf("%s %s", getPrefix("DEBUG"), format), args...)
	}
}

// Infof prints info logs with a formatting directive.
func Infof(format string, args ...interface{}) {
	if logLevel >= LogLevelInfo {
		log.Printf(fmt.Sprintf("%s %s", getPrefix("INFO"), format), args...)
	}
}

// Warnf prints warning logs with a formatting directive.
func Warnf(format string, args ...interface{}) {
	if logLevel >= LogLevelWarning {
		log.Printf(fmt.Sprintf("%s %s", getPrefix("WARN"), format), args...)
	}
}

// Errorf prints error logs with a formatting directive.
func Errorf(format string, args ...interface{}) {
	if logLevel >= LogLevelError {
		log.Printf(fmt.Sprintf("%s %s", getPrefix("ERROR"), format), args...)
	}
}

// Debugln prints debug logs, followed by a new line.
func Debugln(args ...interface{}) {
	if logLevel >= LogLevelDebug {
		args = append([]interface{}{getPrefix("DEBUG")}, args...)
		log.Println(args...)
	}
}

// Infoln prints info logs, followed by a new line.
func Infoln(args ...interface{}) {
	if logLevel >= LogLevelInfo {
		args = append([]interface{}{getPrefix("INFO")}, args...)
		log.Println(args...)
	}
}

// Warnln prints warning logs, followed by a new line.
func Warnln(args ...interface{}) {
	if logLevel >= LogLevelWarning {
		args = append([]interface{}{getPrefix("WARN")}, args...)
		log.Println(args...)
	}
}

// Errorln prints error logs, followed by a new line.
func Errorln(args ...interface{}) {
	if logLevel >= LogLevelError {
		args = append([]interface{}{getPrefix("ERROR")}, args...)
		log.Println(args...)
	}
}

// Debug prints debug logs.
func Debug(args ...interface{}) {
	if logLevel >= LogLevelDebug {
		args = append([]interface{}{getPrefix("DEBUG")}, args...)
		log.Print(args...)
	}
}

// Info prints info logs.
func Info(args ...interface{}) {
	if logLevel >= LogLevelInfo {
		args = append([]interface{}{getPrefix("INFO")}, args...)
		log.Print(args...)
	}
}

// Warn prints warning logs.
func Warn(args ...interface{}) {
	if logLevel >= LogLevelWarning {
		args = append([]interface{}{getPrefix("WARN")}, args...)
		log.Print(args...)
	}
}

// Error prints error logs.
func Error(args ...interface{}) {
	if logLevel >= LogLevelError {
		args = append([]interface{}{getPrefix("ERROR")}, args...)
		log.Print(args...)
	}
}
