// Copyright 2018 Cobaro Pty Ltd. All Rights Reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAx1MAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package elog

import (
	"fmt"
	"io"
	"os"
	"time"
)

const (
	LogLevelEmerg = iota
	LogLevelError
	LogLevelWarning
	LogLevelInfo1
	LogLevelInfo2
	LogLevelDebug1
	LogLevelDebug2
	LogLevelDebug3
)

const (
	LogDateLocaltime = iota
	LogDateUTC
	LogDateEpochSecond
	LogDateEpochMilli
	LogDateEpochMicro
	LogDateEpochNano
	LogDateNone
)

type Elog struct {
	writer     io.Writer
	level      int
	dateFormat int
	logger     func(io.Writer, string, ...interface{}) (int, error)
}

// Set the logfile to an open file
// By defult we use stderr, set to io.Discard to disable
func (log *Elog) SetLogFile(w io.Writer) {
	log.writer = w
}

// Get the current logfile
func (log *Elog) LogFile() (w io.Writer) {
	return log.writer
}

// Set the log level
func (log *Elog) SetLogLevel(level int) {
	log.level = level
}

// Get the log level
func (log *Elog) LogLevel() (level int) {
	return log.level
}

// Set the log format
func (log *Elog) SetLogDateFormat(format int) {
	log.dateFormat = format
}

// Get the log format
func (log *Elog) LogDateFormat() (format int) {
	return log.dateFormat
}

// Set the log function
// If set, then we will call this function instead of our internal
// function. This allows for unification of application and client library
// logging.
func (log *Elog) SetLogFunc(logger func(io.Writer, string, ...interface{}) (int, error)) {
	log.logger = logger
}

// Get the current log function (which may be used)
func (log *Elog) LogFunc() func(io.Writer, string, ...interface{}) (int, error) {
	return log.logger
}

// Actually Log
func (log *Elog) Logf(level int, format string, a ...interface{}) (int, error) {

	if level > log.level {
		return 0, nil
	}

	if log.writer == nil {
		log.writer = os.Stderr
	}

	if log.logger == nil {
		log.logger = fmt.Fprintf
	}

	var t = time.Now()
	var stime string = ""
	var msg string = fmt.Sprintf(format, a...)

	switch log.dateFormat {
	case LogDateLocaltime:
		stime = t.Format(time.UnixDate)
	case LogDateUTC:
		stime = t.UTC().Format(time.UnixDate)
	case LogDateEpochSecond:
		stime = fmt.Sprintf("%d", t.Unix())
	case LogDateEpochMilli:
		seconds := t.Unix()
		millis := (t.UnixNano() - seconds*1000*1000*1000) / (1000 * 1000)
		stime = fmt.Sprintf("%d.%d", seconds, millis)
	case LogDateEpochMicro:
		seconds := t.Unix()
		micros := (t.UnixNano() - seconds*1000*1000*1000) / 1000
		stime = fmt.Sprintf("%d.%d", seconds, micros)
	case LogDateEpochNano:
		seconds := t.Unix()
		nanos := t.UnixNano() - seconds*1000*1000*1000
		stime = fmt.Sprintf("%d.%d", seconds, nanos)
	case LogDateNone:
		return log.logger(log.writer, "%s\n", msg)
	}

	return log.logger(log.writer, "%s %s\n", stime, msg)
}
