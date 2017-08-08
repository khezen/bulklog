package util

import (
	"bytes"
	"time"
)

// RenderIndex - logs: logs-2017.05.26
func RenderIndex(templateName string, t time.Time) string {
	indexBuf := bytes.NewBufferString(string(templateName))
	indexBuf.WriteString("-")
	indexBuf.WriteString(t.Format("2006.01.02"))
	return indexBuf.String()
}
