package skynet

type MultiSemanticLogger []SemanticLogger

func NewMultiSemanticLogger(loggers ...SemanticLogger) (ml MultiSemanticLogger) {
	ml = loggers
	return
}

//
// This section defines methods necessary for MultiSemanticLogger to
// implement SemanticLogger
//

// Log calls .Log(payload) for each logger in the
// MultiSemanticLogger. For each logger, logging behavior may vary
// depending upon the LogLevel.
func (ml MultiSemanticLogger) Log(payload *LogPayload) {
	switch payload.Level {
	default:
		// Log payloads with custom log levels just like those with
		// the known/defult log levels
		fallthrough
	case TRACE, DEBUG, INFO, WARN, ERROR, FATAL:
		for _, lgr := range ml {
			lgr.Log(payload)
		}
	}
}

func (ml MultiSemanticLogger) Trace(msg string) {
	for _, lgr := range ml {
		lgr.Log(NewLogPayload(TRACE, msg))
	}
}

func (ml MultiSemanticLogger) Debug(msg string) {
	for _, lgr := range ml {
		lgr.Log(NewLogPayload(DEBUG, msg))
	}
}

func (ml MultiSemanticLogger) Info(msg string) {
	for _, lgr := range ml {
		lgr.Log(NewLogPayload(INFO, msg))
	}
}

func (ml MultiSemanticLogger) Warn(msg string) {
	for _, lgr := range ml {
		lgr.Log(NewLogPayload(WARN, msg))
	}
}

func (ml MultiSemanticLogger) Error(msg string) {
	for _, lgr := range ml {
		lgr.Log(NewLogPayload(ERROR, msg))
	}
}

// Fatal creates a *LogPayload, adds stacktrace data to it, calls
// .Log(payload) for each logger in the MultiSemanticLogger, then
// panics.
func (ml MultiSemanticLogger) Fatal(msg string) {
	payload := NewLogPayload(FATAL, msg)
	payload.SetException()
	for _, lgr := range ml {
		// Calling .Fatal for each would result in panicking on
		// the first logger, so we log them all, then panic.
		lgr.Log(payload)
	}
	panic(payload)
}

// BenchmarkInfo runs .BenchmarkInfo(level, msg, f) on every logger in
// the MultiSemanticLogger.
func (ml MultiSemanticLogger) BenchmarkInfo(level LogLevel, msg string,
	f func(logger SemanticLogger)) {
	for _, lgr := range ml {
		lgr.BenchmarkInfo(level, msg, f)
	}
}
