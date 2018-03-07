// Copyright 2018 yunfei.liu
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package accesslog

import (
	"net/http"
	"os"
	"log"
	"text/template"
	"time"
	"bytes"

	"github.com/urfave/negroni"
	"github.com/sirupsen/logrus"
	"github.com/meatballhat/negroni-logrus"

	"github.com/yflau/gomw/chi/realip"
)

func logBefore(entry *logrus.Entry, req *http.Request, remoteAddr string) *logrus.Entry {
	return entry.WithFields(logrus.Fields{
		"request": req.RequestURI,
		"method":  req.Method,
		"remote":  remoteAddr,
		"proto" :  req.Proto,
		"useragent": req.UserAgent(),
		"client" : realip.GetRealIP(req.Context()),
	})
}

func logAfter(entry *logrus.Entry, res negroni.ResponseWriter, latency time.Duration, name string) *logrus.Entry {
	return entry.WithFields(logrus.Fields{
		"status":      res.Status(),
		//"text_status": http.StatusText(res.Status()),
		"took":        latency.Seconds() * 1000,
		"size":        res.Size(),
	})
}

// NewLogursLogger return a new access logger from a logrus.Logger as negroni middleware
func NewLogursLogger(logger *logrus.Logger) *negronilogrus.Middleware {
	m := negronilogrus.NewMiddlewareFromLogger(logger, "negroni")
	m.SetLogStarting(false)
	m.Before = logBefore
	m.After = logAfter
	return m
}

// LoggerEntry stands for log entry
type LoggerEntry struct {
	StartTime  string
	Status     int
	Duration   time.Duration
	Hostname   string
	Method     string
	Path       string
	Request    *http.Request
	Response   *http.ResponseWriter
	Size       int
	RemoteAddr string
	ClientAddr string
}

// LoggerDefaultFormat is the format
// logged used by the default Logger instance.
var LoggerDefaultFormat = `{{.ClientAddr}} - - [{{.StartTime}}] "{{.Method}} {{.Request.RequestURI}} {{.Request.Proto}}" {{.Status}} {{.Size}}\n`

// LoggerDefaultDateFormat is the
// format used for date by the
// default Logger instance.
var LoggerDefaultDateFormat = time.RFC3339

// StandardLogger interface
type StandardLogger interface {
	Println(v ...interface{})
	Printf(format string, v ...interface{})
}

// Logger object
type Logger struct {
	// ALogger implements just enough log.Logger interface to be compatible with other implementations
	StandardLogger
	dateFormat string
	template   *template.Template
}

// NewLogger return a new access logger as negroni middleware
func NewLogger() *Logger {
	logger := &Logger{StandardLogger: log.New(os.Stdout, "[negroni] ", 0), dateFormat: LoggerDefaultDateFormat}
	logger.SetFormat(LoggerDefaultFormat)
	return logger
}

// SetFormat set the logging format
func (l *Logger) SetFormat(format string) {
	l.template = template.Must(template.New("negroni_parser").Parse(format))
}

// SetDateFormat set the date format
func (l *Logger) SetDateFormat(format string) {
	l.dateFormat = format
}

func (l *Logger) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()

	next(rw, r)

	res := rw.(negroni.ResponseWriter)
	log := LoggerEntry{
		StartTime:  start.Format(l.dateFormat),
		Status:     res.Status(),
		Duration:   time.Since(start),
		Hostname:   r.Host,
		Method:     r.Method,
		Path:       r.URL.Path,
		Request:    r,
		Response:   &rw,
		Size:       res.Size(),
		RemoteAddr: r.RemoteAddr,
		ClientAddr: realip.GetRealIP(r.Context()),
	}
	buff := &bytes.Buffer{}
	l.template.Execute(buff, log)
	l.Printf(buff.String())
}

