package skylib

import (
	"log"
)

const (
	ERROR = iota
	WARN
	INFO
	DEBUG	
)



func LogError(logLevel int, v ...interface{}){
	
	if logLevel <= *LogLevel {
		log.Println(v)
	}
	
}