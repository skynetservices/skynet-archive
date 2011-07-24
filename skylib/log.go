package skylib

import (
	"log"
)


const (
	error = iota
	warn
	info
	debug
)


func lg(logLevel int, v ...interface{}) {

	if logLevel <= *LogLevel {
		log.Println(v)
	}

}

func LogError(v ...interface{}) {
	lg(error, v)
}

func LogWarn(v ...interface{}) {
	lg(warn, v)
}

func LogInfo(v ...interface{}) {
	lg(info, v)
}

func LogDebug(v ...interface{}) {
	lg(debug, v)
}
