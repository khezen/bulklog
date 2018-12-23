package log

// Level defines the severity of the log message
type Level string

const (
	// LevelVerbose defines the "Verbose" log level/severity
	LevelVerbose Level = "Verbose"
	// LevelInfo defines the "Information" log level/severity
	LevelInfo Level = "Information"
	// LevelWarning defines the "Warning" log level/severity
	LevelWarning Level = "Warning"
	// LevelError defines the "Error" log level/severity
	LevelError Level = "Error"
	// LevelFatal defines the "Fatal" log level/severity
	LevelFatal Level = "Fatal"
)
