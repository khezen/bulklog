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
	bulklogEndpoint string
}

// bulklogLogger contains data necessary to log with bulklog
type bulklogLogger struct {
	config Config
	msg    chan Message
	web    chan WebCall
}

// NewLogger creates new bulklog logger
func NewLogger(config Config) Logger {
	return &bulklogLogger{
		config,
		make(chan Message),
		make(chan WebCall),
	}
}

// Message logs a message
func (r *bulklogLogger) Message(lvl Level, rid, msg string) {
	r.msg <- Message{
		Level:     lvl,
		Source:    r.config.Source,
		RequestID: rid,
		Message:   msg,
	}
}

// WebCall logs a web request/response
func (r *bulklogLogger) WebCall(rid, cip, host, path, method, reqStr string, statCode int, resStr string, sec float64) {
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
func (r *bulklogLogger) ListenAndServe() {
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

func (r *bulklogLogger) renderBody(document interface{}) []byte {
	body, err := json.Marshal(document)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
	return body
}

func (r *bulklogLogger) postMsg(msg Message) {
	url := fmt.Sprintf("%s/bulklog/v1/logs/log", r.config.bulklogEndpoint)
	body := r.renderBody(msg)
	res, err := http.Post(url, "application/json", bytes.NewReader(body))
	r.handleResponse(res, err)
}

func (r *bulklogLogger) postTrace(trace WebCall) {
	url := fmt.Sprintf("%s/bulklog/v1/web/trace", r.config.bulklogEndpoint)
	body := r.renderBody(trace)
	res, err := http.Post(url, "application/json", bytes.NewReader(body))
	r.handleResponse(res, err)
}

func (r *bulklogLogger) handleResponse(res *http.Response, err error) {
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
