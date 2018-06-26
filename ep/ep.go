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
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"bufio"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/cobaro/elvin/elog"
	"github.com/cobaro/elvin/elvin"
	"os"
	"os/signal"
	"strconv"
	"strings"
)

type arguments struct {
	help     bool
	endpoint string
	number   int
	unotify  bool
}

func main() {
	// Parse command line options
	args := flags()
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)

	var notify func(map[string]interface{}, bool, elvin.KeyBlock) error

	ep := elvin.NewClient(args.endpoint, nil, nil, nil)
	ep.SetLogDateFormat(elog.LogDateLocaltime)
	ep.SetLogLevel(elog.LogLevelInfo1)

	if !args.unotify {
		if err := ep.Connect(); err != nil {
			ep.Logf(elog.LogLevelInfo1, "%v", err)
			os.Exit(1)
		}
		ep.Logf(elog.LogLevelInfo1, "connected to %s", args.endpoint)
		notify = ep.Notify
	} else {
		notify = ep.UNotify
	}

	// Set up our notitfication reader
	notifications := make(chan map[string]interface{})
	go parseNotifications(ep, notifications)

Loop:
	for {
		select {
		case notification, more := <-notifications:
			if more {
				// ep.Logf(elog.LogLevelInfo1, "read %+v", notification)

				for i := 0; i < args.number; i++ {
					if err := notify(notification, true, nil); err != nil {
						ep.Logf(elog.LogLevelInfo1, "Notify failed")
					}
				}
			} else {
				ep.Logf(elog.LogLevelInfo1, "Exiting")
				break Loop
			}
		case s := <-sig:
			ep.Logf(elog.LogLevelInfo1, "Exiting on %v", s)
			break Loop
		}
	}

	if !args.unotify {
		if err := ep.Disconnect(); err != nil {
			ep.Logf(elog.LogLevelInfo1, "%v", err)
			os.Exit(1)
		}
		ep.Logf(elog.LogLevelInfo1, "Disconnected")
	}

	os.Exit(0)
}

// Argument parsing
func flags() (args arguments) {
	flag.BoolVar(&args.help, "h", false, "prints this help")
	flag.StringVar(&args.endpoint, "e", "localhost:2917", "host:port of router")
	flag.IntVar(&args.number, "n", 1, "number of notifications to send")
	flag.BoolVar(&args.unotify, "unotify", false, "send using UNotify")
	flag.Parse()

	if args.help {
		flag.Usage()
		os.Exit(0)
	}

	return args
}

// Read's notifications from stdin into a channel, exits on EOF
func parseNotifications(ep *elvin.Client, sendto chan map[string]interface{}) {
	ep.Logf(elog.LogLevelInfo1, "parser starting")
	scanner := bufio.NewScanner(os.Stdin)

	nfn := make(map[string]interface{})

	for scanner.Scan() {
		// ep.Logf(elog.LogLevelInfo1, ("parser processing:", scanner.Text())
		// Look for end of message marker '^---.*$'
		if scanner.Text()[:3] == "---" {
			if len(nfn) > 0 {
				sendto <- nfn
				nfn = make(map[string]interface{})
			}
		} else {
			// look for name : value (with or without space around :)
			namevalue := strings.SplitN(scanner.Text(), ":", 2)
			if len(namevalue) != 2 {
				ep.Logf(elog.LogLevelInfo1, "Failed to parse '%s' as attribute: value", scanner.Text())
			} else {
				// Try to convert the value
				name := strings.TrimSpace(namevalue[0])
				value := strings.TrimSpace(namevalue[1])
				// ep.Logf(elog.LogLevelInfo1, "%s:%s", name, value)

				if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
					// string "delimited"
					nfn[name] = value[1 : len(value)-1]

				} else if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
					// opaque [delimited]
					size := (len(value) - 2) / 2 // half what's between the []
					opaque := make([]byte, size)
					len, err := hex.Decode(opaque, []byte(value[1:len(value)-1]))
					// ep.Logf(elog.LogLevelInfo1, "opaque %v len:%d, in:%d, out:%d", opaque, len, 7-2, 7)
					if err != nil {
						ep.Logf(elog.LogLevelInfo1, "ParseError: %s", err.Error())

					} else if size != len {
						ep.Logf(elog.LogLevelInfo1, "ParseError: Couldn't convert entirety of %s", value)
					} else {
						nfn[name] = opaque
					}
				} else if strings.HasSuffix(value, "L") || strings.HasSuffix(value, "l") {
					// int64 e.g., 123L
					i64, err := strconv.ParseInt(value[:len(value)-1], 10, 64)
					if err != nil {
						ep.Logf(elog.LogLevelInfo1, "ParseError: converting '%s' to int64: %v", value, err.Error())
					} else {
						nfn[name] = i64
					}
				} else if strings.Contains(value, ".") {
					// float64 e.g. 3.14
					f64, err := strconv.ParseFloat(value, 64)
					if err != nil {
						ep.Logf(elog.LogLevelInfo1, "ParseError: converting '%s' to float64: %v", value, err.Error())
					} else {
						nfn[name] = f64
					}
				} else {
					// int32
					i64, err := strconv.ParseInt(value, 10, 32)
					if err != nil {
						ep.Logf(elog.LogLevelInfo1, "ParseError: converting '%s' to int32: %v", value, err.Error())
					} else {
						nfn[name] = int32(i64)
					}
				}
			}
			// ep.Logf(elog.LogLevelInfo1, "%+v", nfn)
		}
	}
	// EOF which is normal if a file is redirected in for example
	// so send it and return, exiting the parser
	if len(nfn) > 0 {
		fmt.Println("Send out", nfn)
		sendto <- nfn
	}
	close(sendto)
}
