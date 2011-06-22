package skylib

import (
	"log"
)


func LogError(logLevel int, v ...interface{}) {

	if logLevel <= *LogLevel {
		log.Println(v)
	}

}
