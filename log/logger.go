package log

// Logger defines the Peaks services generic logging interface
type Logger interface {
	Message(lvl Level, rid, msg string)
	WebCall(rid, cip, host, path, method, reqBody string, statCode int, resBody string, sec float64)
}
