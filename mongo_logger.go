package skynet

import (
	"fmt"
	"labix.org/v2/mgo"
	"log"
)

// MongoSemanticLogger saves logging data to a MongoDB instance.
type MongoSemanticLogger struct {
	session                *mgo.Session
	dbName, collectionName string
	uuid                   string
}

// NewMongoSemanticLogger connects to a MongoDB instance at the given
// address (often "localhost").
func NewMongoSemanticLogger(addr, dbName, collectionName,
	uuid string) (ml *MongoSemanticLogger, err error) {
	ml = &MongoSemanticLogger{
		dbName:         dbName,
		collectionName: collectionName,
		uuid:           uuid,
	}
	ml.session, err = mgo.Dial(addr)
	return
}

// Log saves all fields of the given payload to MongoDB, setting
// unexported fields as necessary. May behave differently based upon
// payload.LogLevel.
func (ml *MongoSemanticLogger) Log(payload *LogPayload) {
	// Sanity checks
	if ml == nil {
		log.Printf("NOT LOGGING: Can't log to nil *MongoSemanticLogger\n")
		return
	}
	if payload == nil {
		log.Printf("NOT LOGGING: Can't log nil *LogPayload\n")
		return
	}

	// Set various Payload fields
	payload.setKnownFields()
	payload.Name = fmt.Sprintf("%T", ml)
	payload.UUID = ml.uuid
	payload.Table = ml.collectionName

	// Log regardless of the log level
	err := ml.session.DB(ml.dbName).C(ml.collectionName).Insert(payload)
	if err != nil {
		errStr := "Error logging with MongoSemanticLogger %s: %v"
		log.Printf(errStr, ml.uuid, err)
	}
}

// Fatal logs the given payload to MongoDB, then panics.
func (ml *MongoSemanticLogger) Fatal(payload *LogPayload) {
	payload.Backtrace = genStacktrace()
	ml.Log(payload)
	panic(payload)
}

// BenchmarkInfo currently does nothing but should measure the time
// it takes to execute `f` based on the log level
func (ml *MongoSemanticLogger) BenchmarkInfo(level LogLevel, msg string,
	f func(logger SemanticLogger)) {
	// TODO: Implement
}
