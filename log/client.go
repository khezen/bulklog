package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Config contains end point for a service
type Config struct {
	Service       string
	LogEndpoint   string
	TraceEndpoint string
}

// SetConfig retuns Config Object from given parameters
func SetConfig(service, logEndpoint, traceEndpoint string) *Config {
	return &Config{
		service,
		logEndpoint,
		traceEndpoint,
	}
}

// espipeLogger contains data necessary to log with espipe
type espipeLogger struct {
	config Config
	msg    chan Message
	web    chan WebCall
}

// NewLogger creates new espipe logger
func NewLogger(config Config) Logger {
	return &espipeLogger{
		config,
		make(chan Message),
		make(chan WebCall),
	}
}

// Message logs a message
func (r *espipeLogger) Message(lvl Level, rid, msg string) {
	r.msg <- Message{
		Level:     lvl,
		Service:   r.config.Service,
		RequestID: rid,
		Message:   msg,
	}
}

// WebCall logs a web request/response
func (r *espipeLogger) WebCall(rid, cip, host, path, method, reqBody string, statCode int, resBody string, sec float64) {
	r.web <- WebCall{
		Level:             LevelVerbose,
		Service:           r.config.Service,
		Message:           WebCallMessage,
		RequestID:         rid,
		ClientIP:          cip,
		Host:              host,
		Path:              path,
		Method:            method,
		RequestBody:       reqBody,
		StatusCode:        statCode,
		ResponseBody:      resBody,
		ResponseInSeconds: sec,
	}
}

// ListenAndServe blocks the current goroutine and start sending all messages and web calls to logstash
func (r *espipeLogger) ListenAndServe() {
	for {
		select {
		case m := <-r.msg:
			go func() {
				fmt.Printf("%v %v: %v\n", time.Now().UTC().Format("2006-01-02T15:04:05.999999Z"), m.Level, m.Message)
				r.postMsg(m)
			}()
			break
		case c := <-r.web:
			go r.postTrace(c)
			break
		}
	}
}

func (r *espipeLogger) renderBody(document interface{}) []byte {
	body, err := json.Marshal(document)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
	return body
}

func (r *espipeLogger) postMsg(msg Message) {
	body := r.renderBody(msg)
	res, err := http.Post(r.config.LogEndpoint, "application/json", bytes.NewReader(body))
	r.handleResponse(res, err)
}

func (r *espipeLogger) postTrace(trace WebCall) {
	body := r.renderBody(trace)
	res, err := http.Post(r.config.TraceEndpoint, "application/json", bytes.NewReader(body))
	r.handleResponse(res, err)
}

func (r *espipeLogger) handleResponse(res *http.Response, err error) {
	if err != nil {
		fmt.Printf("ERROR in sending message: %v\n", err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Printf("ERROR in sending log message: Response status: %v\n", res.Status)
		} else {
			fmt.Printf("ERROR in sending log message: Response status: %v; Response body: %v;\n", res.Status, string(body))
		}
	}
}
