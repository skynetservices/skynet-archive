package log

import (
	"log"
	"os"
)

/* TODO:
- Should prefix lines with log level
- Should possibly add Debug, Debugf type helper methods
*/

var logger *log.Logger

type LogLevel int8

var level LogLevel

const (
	DEBUG LogLevel = iota
	TRACE
	INFO
	WARN
	ERROR
	FATAL
	PANIC
)

func init() {
	// Default the logger, implementors can override the Output if they'd like to change it
	logger = log.New(os.Stdout, "skynet: ", log.LstdFlags)
}

func Fatal(v ...interface{}) {
	if level <= FATAL {
		logger.Fatal(v...)
	}
}

func Fatalf(format string, v ...interface{}) {
	if level <= FATAL {
		logger.Fatalf(format, v...)
	}
}

func Fatalln(v ...interface{}) {
	if level <= FATAL {
		logger.Fatalln(v...)
	}
}

func Flags() int {
	return logger.Flags()
}

func Panic(v ...interface{}) {
	if level <= PANIC {
		logger.Panic(v...)
	}
}

func Panicf(format string, v ...interface{}) {
	if level <= PANIC {
		logger.Panicf(format, v...)
	}
}

func Panicln(v ...interface{}) {
	if level <= PANIC {
		logger.Panicln(v...)
	}
}

func Prefix() string {
	return logger.Prefix()
}

func Print(level LogLevel, v ...interface{}) {
	if level <= level {
		logger.Print(v...)
	}
}

func Printf(level LogLevel, format string, v ...interface{}) {
	if level <= level {
		logger.Printf(format, v...)
	}
}

func Println(level LogLevel, v ...interface{}) {
	if level <= level {
		logger.Println(v...)
	}
}

func SetFlags(flag int) {
	logger.SetFlags(flag)
}

func SetPrefix(prefix string) {
	logger.SetPrefix(prefix)
}

func SetLogLevel(level LogLevel) {
	level = level
}
