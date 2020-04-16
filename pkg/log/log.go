package log

import (
	"log"
	"os"
)

var (
	stderr = log.New(os.Stderr, "", log.Lshortfile)
	stdout = log.New(os.Stdout, "", log.Lshortfile)
)

// Err returns a logger over stderr
func Err() *log.Logger {
	return stderr
}

// Out returns a logger over stdout
func Out() *log.Logger {
	return stdout
}
