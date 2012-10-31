package skynet

import (
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
	payload.UUID = ml.uuid

	// Log regardless of the log level
	err := ml.session.DB(ml.dbName).C(ml.collectionName).Insert(payload)
	if err != nil {
		errStr := "Error logging with MongoSemanticLogger %s: %v"
		log.Printf(errStr, ml.uuid, err)
	}
}

// Trace logs the given payload to MongoDB
func (ml *MongoSemanticLogger) Trace(msg string) {
	ml.Log(NewLogPayload(TRACE, msg))
}

// Debug logs the given payload to MongoDB
func (ml *MongoSemanticLogger) Debug(msg string) {
	ml.Log(NewLogPayload(DEBUG, msg))
}

// Info logs the given payload to MongoDB
func (ml *MongoSemanticLogger) Info(msg string) {
	ml.Log(NewLogPayload(INFO, msg))
}

// Warn logs the given payload to MongoDB
func (ml *MongoSemanticLogger) Warn(msg string) {
	ml.Log(NewLogPayload(WARN, msg))
}

// Error logs the given payload to MongoDB
func (ml *MongoSemanticLogger) Error(msg string) {
	ml.Log(NewLogPayload(ERROR, msg))
}

// Fatal logs the given payload to MongoDB (after adding stacktrace
// data), then panics.
func (ml *MongoSemanticLogger) Fatal(msg string) {
	payload := NewLogPayload(FATAL, msg)
	payload.SetException()
	ml.Log(payload)
	panic(payload)
}

// BenchmarkInfo currently does nothing but should measure the time
// it takes to execute `f` based on the log level
func (ml *MongoSemanticLogger) BenchmarkInfo(level LogLevel, msg string,
	f func(logger SemanticLogger)) {
	// TODO: Implement
}
