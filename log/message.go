package log

const (
	// WebCallMessage is the default message for WebCallLogs
	WebCallMessage = "http-call-trace"
)

// Message is a basic representation of a log
type Message struct {
	Level     Level  `json:"level"`
	Service   string `json:"service"`
	RequestID string `json:"request_id"`
	Message   string `json:"message"`
}

// WebCall is http call trace ready to be logged
type WebCall struct {
	Level             Level   `json:"level"`
	Service           string  `json:"service"`
	Message           string  `json:"message"`
	RequestID         string  `json:"request_id"`
	ClientIP          string  `json:"client_ip"`
	Host              string  `json:"host"`
	Path              string  `json:"path"`
	Method            string  `json:"method"`
	Request           string  `json:"request"`
	StatusCode        int     `json:"status_code"`
	Response          string  `json:"response"`
	ResponseInSeconds float64 `json:"response_in_seconds"`
}
