package log

import (
	"log"
	"os"
)

/* TODO:
- Should possibly add Debug, Debugf type helper methods
*/

var logger *log.Logger

type LogLevel int8

var minLevel LogLevel

const (
	DEBUG LogLevel = iota
	TRACE
	INFO
	WARN
	ERROR
	FATAL
	PANIC
)

func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG "
	case TRACE:
		return "TRACE "
	case INFO:
		return "INFO "
	case WARN:
		return "WARN "
	case ERROR:
		return "ERROR "
	case FATAL:
		return "FATAL "
	case PANIC:
		return "PANIC "
	}

	return ""
}

func (l LogLevel) Interface() interface{} {
	return interface{}(l.String())
}

func init() {
	// Default the logger, implementors can override the Output if they'd like to change it
	logger = log.New(os.Stdout, "skynet: ", log.LstdFlags)
}

func Fatal(v ...interface{}) {
	if minLevel <= FATAL {
		Print(FATAL, v...)
	}
}

func Fatalf(format string, v ...interface{}) {
	if minLevel <= FATAL {
		Printf(FATAL, format, v...)
	}
}

func Fatalln(v ...interface{}) {
	if minLevel <= FATAL {
		Println(FATAL, v...)
	}
}

func Flags() int {
	return logger.Flags()
}

func Panic(v ...interface{}) {
	if minLevel <= PANIC {
		Print(PANIC, v...)
	}
}

func Panicf(format string, v ...interface{}) {
	if minLevel <= PANIC {
		Printf(PANIC, format, v...)
	}
}

func Panicln(v ...interface{}) {
	if minLevel <= PANIC {
		Println(PANIC, v...)
	}
}

func Prefix() string {
	return logger.Prefix()
}

func Print(level LogLevel, v ...interface{}) {
	if minLevel <= level {
		args := []interface{}{level.Interface()}
		args = append(args, v)

		switch level {
		case FATAL:
			logger.Fatal(args...)
		case PANIC:
			logger.Panic(args...)
		default:
			logger.Print(args...)
		}
	}
}

func Printf(level LogLevel, format string, v ...interface{}) {
	if minLevel <= level {
		logger.Printf(level.String()+format, v...)
	}
}

func Println(level LogLevel, v ...interface{}) {
	if minLevel <= level {
		l := []interface{}{level.Interface()}
		l = append(l, v)
		logger.Println(l...)
	}
}

func SetFlags(flag int) {
	logger.SetFlags(flag)
}

func SetPrefix(prefix string) {
	logger.SetPrefix(prefix)
}

func SetLogLevel(level LogLevel) {
	minLevel = level
}

func GetLogLevel() LogLevel {
	return minLevel
}

func LevelFromString(l string) (level LogLevel) {
	switch l {
	case "DEBUG":
		level = DEBUG
	case "TRACE":
		level = TRACE
	case "INFO":
		level = INFO
	case "WARN":
		level = WARN
	case "ERROR":
		level = ERROR
	case "FATAL":
		level = FATAL
	case "PANIC":
		level = PANIC
	}

	return
}
