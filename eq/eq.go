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
	"flag"
	"github.com/cobaro/elvin/elog"
	"github.com/cobaro/elvin/elvin"
	"os"
	"os/signal"
	"time"
)

// Handle Disconnects stub
func disconnector(client *elvin.Client) {
	for {
		select {
		case event := <-client.Events:
			switch event.(type) {
			case *elvin.Disconn:
				disconn := event.(*elvin.Disconn)
				// client.Logf(elog.LogLevelInfo1, "Received Disconn:\n%v", disconn)
				switch disconn.Reason {

				case elvin.DisconnReasonRouterShuttingDown:
					client.Logf(elog.LogLevelInfo1, "router shutting down, exiting")
					os.Exit(1)

				case elvin.DisconnReasonRouterProtocolErrors:
					client.Logf(elog.LogLevelInfo1, "router thinks we violated the protocol")
					os.Exit(1)

				case elvin.DisconnReasonRouterRedirect:
					if len(disconn.Args) > 0 {
						client.Logf(elog.LogLevelInfo1, "redirected to %s", disconn.Args)
						// FIXME: tidy this
						client.Endpoint = disconn.Args
						client.Close()
						// client.Logf(elog.LogLevelInfo1, "disconnector State(%d)", client.State())
						if err := client.Connect(); err != nil {
							client.Logf(elog.LogLevelInfo1, "%v", err)
							os.Exit(1)
						}
						client.Logf(elog.LogLevelInfo1, "connected to %s", client.Endpoint)
					} else {
						client.Logf(elog.LogLevelInfo1, "redirected to %s", disconn.Args)
					}
					break

				case elvin.DisconnReasonClientConnectionLost:
					client.Logf(elog.LogLevelInfo1, "Lost connection. Reconnecting")
					if err := client.DefaultReconnect(0, time.Duration(0), time.Duration(time.Minute)); err != nil {
						client.Logf(elog.LogLevelInfo1, "Reconnect failed")
					} else {
						client.Logf(elog.LogLevelInfo1, "Reconnected")
					}

				case elvin.DisconnReasonClientProtocolErrors:
					client.Logf(elog.LogLevelInfo1, "client library detected protocol errors")
					os.Exit(1)
				}
			case *elvin.DropWarn:
				client.Logf(elog.LogLevelInfo1, "DropWarn (lost packets)")

			default:
				client.Logf(elog.LogLevelInfo1, "FIXME: bad connection notification")
				os.Exit(1)

			}
		}
	}
}

func main() {
	// Argument parsing
	flag.Parse()

	endpoint := "localhost:2917"
	eq := elvin.NewClient(endpoint, nil, nil, nil)
	eq.SetLogDateFormat(elog.LogDateLocaltime)
	eq.SetLogLevel(elog.LogLevelInfo1)
	go disconnector(eq)

	eq.Options = make(map[string]interface{})
	// FIXME: At some point let's formalize these as test cases
	// eq.Options["TestNack"] = 1
	// eq.Options["TestDisconn"] = 1
	// eq.Logf(elog.LogLevelInfo1, "Options:%v\n", eq.Options)

	if err := eq.Connect(); err != nil {
		eq.Logf(elog.LogLevelInfo1, "%v", err)
		os.Exit(1)
	}
	eq.Logf(elog.LogLevelInfo1, "connected to %s", endpoint)

	// FIXME: do a NewSubscription()
	quench := new(elvin.Quench)
	quench.Names = map[string]bool{"int32": true, "float64": true}
	quench.DeliverInsecure = true
	quench.Keys = nil
	quench.Notifications = make(chan elvin.QuenchNotification)

	if err := eq.Quench(quench); err != nil {
		eq.Logf(elog.LogLevelInfo1, "Quench failed %v", err)
	} else {
		eq.Logf(elog.LogLevelInfo1, "Quench succeeded %v", *quench)
	}

	addNames := map[string]bool{"int64": true}
	delNames := map[string]bool{"float64": true}
	if err := eq.QuenchModify(quench, addNames, delNames, true, nil, nil); err != nil {
		eq.Logf(elog.LogLevelInfo1, "Quench mod failed %v", err)
	} else {
		eq.Logf(elog.LogLevelInfo1, "Quench mod succeeded %v", *quench)
	}

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)

Loop:
	for {
		select {
		case sig := <-ch:
			eq.Logf(elog.LogLevelInfo1, "Exiting on %v", sig)
			break Loop
		case pkt := <-quench.Notifications:
			eq.Logf(elog.LogLevelInfo1, "Received quench:\n%v", pkt)
		}
	}

	if err := eq.QuenchDelete(quench); err != nil {
		eq.Logf(elog.LogLevelInfo1, "QuenchDel failed %v", err)
	} else {
		eq.Logf(elog.LogLevelInfo1, "QuenchDel succeeded %v", quench)
	}

	// Exit a little gracefully
	eq.Logf(elog.LogLevelInfo1, "Disconnecting")
	if err := eq.Disconnect(); err != nil {
		eq.Logf(elog.LogLevelInfo1, "%v", err)
		os.Exit(1)
	}

	eq.Logf(elog.LogLevelInfo1, "Disconnected")
	os.Exit(0)
}
