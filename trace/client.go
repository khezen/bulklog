package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Config -
type Config struct {
	Source         string
	espipeEndpoint string
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
		Source:    r.config.Source,
		RequestID: rid,
		Message:   msg,
	}
}

// WebCall logs a web request/response
func (r *espipeLogger) WebCall(rid, cip, host, path, method, reqStr string, statCode int, resStr string, sec float64) {
	r.web <- WebCall{
		Level:             LevelVerbose,
		Source:            r.config.Source,
		RequestID:         rid,
		ClientIP:          cip,
		Host:              host,
		Path:              path,
		Method:            method,
		Request:           reqStr,
		StatusCode:        statCode,
		Response:          resStr,
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
	url := fmt.Sprintf("%s/espipe/v1/logs/log", r.config.espipeEndpoint)
	body := r.renderBody(msg)
	res, err := http.Post(url, "application/json", bytes.NewReader(body))
	r.handleResponse(res, err)
}

func (r *espipeLogger) postTrace(trace WebCall) {
	url := fmt.Sprintf("%s/espipe/v1/web/trace", r.config.espipeEndpoint)
	body := r.renderBody(trace)
	res, err := http.Post(url, "application/json", bytes.NewReader(body))
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
