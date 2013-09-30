// Package log provides syslog logging to a local or remote
// Syslog logger.  To specify a remote syslog host, set the
// "log.sysloghost" key in the Skynet configuration.  Specify
// the port with "log.syslogport".  If "log.sysloghost" is not provided,
// skynet will log to local syslog.
package log

import (
	"fmt"
	"log/syslog"
	"strconv"
)

type LogLevel int8

var syslogHost string
var syslogPort int = 0

var minLevel LogLevel
var logger *syslog.Writer

const (
	TRACE LogLevel = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
	PANIC
)

// Call Initialize after setting (or not setting) SyslogHost and SyslogPort when
// they're read from configuration source.
func Initialize() {

	var e error

	if len(syslogHost) > 0 {

		logger, e = syslog.New(syslog.LOG_INFO|syslog.LOG_USER, "skynet")
		if e != nil {
			panic(e)
		}
	} else {
		logger, e = syslog.Dial("tcp4", syslogHost+":"+strconv.Itoa(syslogPort), syslog.LOG_INFO|syslog.LOG_USER, "skynet")
		if e != nil {
			panic(e)
		}
	}

}

func Panic(message interface{}) {
	logger.Emerg(message.(string))
}

func Panicf(format string, messages ...interface{}) {
	m := fmt.Sprintf(format, messages...)
	logger.Emerg(m)
}

func Fatal(message interface{}) {
	if minLevel <= FATAL {
		logger.Crit(message.(string))
	}
}

func Fatalf(format string, messages ...interface{}) {
	if minLevel <= FATAL {
		m := fmt.Sprintf(format, messages...)
		logger.Crit(m)
	}
}

func Error(message interface{}) {
	if minLevel <= ERROR {
		logger.Err(message.(string))
	}
}

func Errorf(format string, messages ...interface{}) {
	if minLevel <= ERROR {
		m := fmt.Sprintf(format, messages...)
		logger.Err(m)
	}
}

func Warn(message interface{}) {
	if minLevel <= WARN {
		logger.Warning(message.(string))
	}
}

func Warnf(format string, messages ...interface{}) {
	if minLevel <= WARN {
		m := fmt.Sprintf(format, messages...)
		logger.Warning(m)
	}
}

func Info(message interface{}) {
	if minLevel <= INFO {
		logger.Info(message.(string))
	}
}

func Infof(format string, messages ...interface{}) {
	if minLevel <= INFO {
		m := fmt.Sprintf(format, messages...)
		logger.Info(m)
	}
}

func Debug(message interface{}) {
	if minLevel <= DEBUG {
		logger.Debug(message.(string))
	}
}

func Debugf(format string, messages ...interface{}) {
	if minLevel <= DEBUG {
		m := fmt.Sprintf(format, messages...)
		logger.Debug(m)
	}
}

func Trace(message interface{}) {
	if minLevel <= TRACE {
		logger.Debug(message.(string))
	}
}

func Tracef(format string, messages ...interface{}) {
	if minLevel <= TRACE {
		m := fmt.Sprintf(format, messages...)
		logger.Debug(m)
	}
}

func Println(level LogLevel, messages ...interface{}) {

	switch level {
	case DEBUG:
		Debugf("%v", messages)
	case TRACE:
		Tracef("%v", messages)
	case INFO:
		Infof("%v", messages)
	case WARN:
		Warnf("%v", messages)
	case ERROR:
		Errorf("%v", messages)
	case FATAL:
		Fatalf("%v", messages)
	case PANIC:
		Panicf("%v", messages)
	}

	return
}

func Printf(level LogLevel, format string, messages ...interface{}) {

	switch level {
	case DEBUG:
		Debugf(format, messages)
	case TRACE:
		Tracef(format, messages)
	case INFO:
		Infof(format, messages)
	case WARN:
		Warnf(format, messages)
	case ERROR:
		Errorf(format, messages)
	case FATAL:
		Fatalf(format, messages)
	case PANIC:
		Panicf(format, messages)
	}

	return
}

func SetSyslogHost(host string) {
	syslogHost = host
}

func SetSyslogPort(port int) {
	syslogPort = port
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
