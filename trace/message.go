package log

// Message is a basic representation of a log
type Message struct {
	Level     Level  `json:"level"`
	Source    string `json:"source"`
	RequestID string `json:"request_id"`
	Message   string `json:"message"`
}

// WebCall is http call trace ready to be logged
type WebCall struct {
	Level             Level   `json:"level"`
	Source            string  `json:"source"`
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
