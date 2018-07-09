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
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/cobaro/elvin/elog"
	"github.com/cobaro/elvin/elvin"
	"os"
	"os/signal"
	"time"
)

type arguments struct {
	help              bool
	verbosity         int
	url               string
	number            int
	acceptInsecure    bool
	producerKeyString string
	producerKeyHex    string
	consumerKeyString string
	consumerKeyHex    string
	secureAcceptance  bool
}

func main() {
	// Argument parsing
	args := flags()
	flag.Parse()

	ec := elvin.NewClient(args.url, nil, nil, nil)
	ec.SetLogDateFormat(elog.LogDateLocaltime)
	ec.SetLogLevel(args.verbosity)

	// Process security arguments
	// FIXME; How compatible to be here? for now not very
	// !multiple args
	// !smart security
	// !load from file
	// stick with sha1 instead of 256?
	// producerKeyBlock := make(map[int]elvin.KeySetList)
	// consumerKeyBlock := make(map[int]elvin.KeySetList)
	ec.KeysSub = make(map[int]elvin.KeySetList)
	var producerKeySet elvin.KeySet
	var consumerKeySet elvin.KeySet

	if len(args.producerKeyString) > 0 {
		// This is totally bogus as it can't be primed and we're not handling files
		ec.Logf(elog.LogLevelError, "Opening keyfiles not supported yet")
		os.Exit(1)
	}
	if len(args.producerKeyHex) > 0 {
		if key, err := hex.DecodeString(args.producerKeyHex); err != nil {
			ec.Logf(elog.LogLevelError, "Failed to interpret keyhex: %v", err)
			os.Exit(1)
		} else {
			elvin.KeySetAddKey(&producerKeySet, key)
		}
	}
	ec.KeysSub[elvin.KeySchemeSha1Producer] = elvin.KeySetList{producerKeySet}

	if len(args.consumerKeyString) > 0 {
		elvin.KeySetAddKey(&consumerKeySet, []byte(args.consumerKeyString))
	}
	if len(args.consumerKeyHex) > 0 {
		if key, err := hex.DecodeString(args.consumerKeyHex); err != nil {
			ec.Logf(elog.LogLevelError, "Failed to interpret keyhex: %v", err)
			os.Exit(1)
		} else {
			elvin.KeySetAddKey(&consumerKeySet, key)
		}
	}
	ec.KeysSub[elvin.KeySchemeSha1Consumer] = elvin.KeySetList{consumerKeySet}
	ec.Logf(elog.LogLevelDebug2, "secureAcceptance is: %v, keys: %v", args.secureAcceptance, ec.KeysSub)

	// Using default disconnector for now
	// go disconnector(ec)

	ec.Options = make(map[string]interface{})
	// FIXME: At some point let's formalize these as test cases
	// ec.Options["TestNack"] = 1
	// ec.Options["TestDisconn"] = 1
	// ec.Logf(elog.LogLevelInfo1, "Options:%v\n", ec.Options)

	if err := ec.Connect(); err != nil {
		ec.Logf(elog.LogLevelInfo1, "%v", err)
		os.Exit(1)
	}
	ec.Logf(elog.LogLevelInfo1, "connected to %s", args.url)

	// FIXME: do a NewSubscription()
	sub := new(elvin.Subscription)
	sub.Expression = "require(int32)"
	// sub.Expression = "bogus"
	sub.AcceptInsecure = true
	sub.Keys = nil
	sub.Notifications = make(chan map[string]interface{})

	if err := ec.Subscribe(sub); err != nil {
		ec.Logf(elog.LogLevelError, "Subscribe failed %v", err)
	} else {
		ec.Logf(elog.LogLevelInfo2, "Subscribe succeeded %v", sub)
	}

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)

	received := 0
	timeStart := time.Now()
Loop:
	for {
		select {
		case sig := <-ch:
			ec.Logf(elog.LogLevelInfo1, "Exiting on %v", sig)
			break Loop
		case nfn := <-sub.Notifications:
			if args.number == 1 {
				if s, err := elvin.NameValueToString(nfn, true); err != nil {
					ec.Logf(elog.LogLevelError, err.Error())
				} else {
					fmt.Println(s)
				}
			} else {
				received++
				if received == args.number {
					timeNow := time.Now()
					received = 0
					fmt.Println(timeNow.Sub(timeStart))
					timeStart = timeNow
				}
			}
		}
	}

	if err := ec.SubscriptionDelete(sub); err != nil {
		ec.Logf(elog.LogLevelInfo1, "SubDel failed %v", err)
	} else {
		ec.Logf(elog.LogLevelInfo1, "SubDel succeeded %v", sub)
	}

	// Exit a little gracefully
	ec.Logf(elog.LogLevelInfo1, "Disconnecting")
	if err := ec.Disconnect(); err != nil {
		ec.Logf(elog.LogLevelInfo1, "%v", err)
		os.Exit(1)
	}

	ec.Logf(elog.LogLevelInfo1, "Disconnected")
	os.Exit(0)
}

// Argument parsing
func flags() (args arguments) {
	flag.BoolVar(&args.help, "h", false, "Print this help")
	flag.StringVar(&args.url, "e", "elvin://", "elvin url e.g., elvin://host")
	flag.IntVar(&args.number, "n", 1, "number of notifications to receive before reporting")
	flag.IntVar(&args.verbosity, "v", 3, "verbosity (default 3)")
	flag.StringVar(&args.producerKeyString, "p", "", "SHA1 producer public key (string) ")
	flag.StringVar(&args.producerKeyHex, "P", "", "SHA1 producer public key (hex)")
	flag.StringVar(&args.consumerKeyString, "c", "", "SHA1 consumer private key (string) ")
	flag.StringVar(&args.consumerKeyHex, "C", "", "SHA1 consumer private key (hex)")
	flag.BoolVar(&args.secureAcceptance, "x", false, "Don't allow insecure acceptance (default is to allow)")

	flag.Parse()

	if args.help {
		flag.Usage()
		os.Exit(0)
	}

	return args
}

func disconnector(client *elvin.Client) {
	for {
		select {
		case event := <-client.Events:
			switch event.(type) {
			case *elvin.Disconn:
				disconn := event.(*elvin.Disconn)
				client.Logf(elog.LogLevelInfo1, "Received Disconn:\n%v", disconn)
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
						client.URL = disconn.Args
						client.Disconnect()
						// client.Logf(elog.LogLevelInfo1, "disconnector State(%d)", client.State())
						if err := client.Connect(); err != nil {
							client.Logf(elog.LogLevelInfo1, "%v", err)
							os.Exit(1)
						}
						client.Logf(elog.LogLevelInfo1, "connected to %s", client.URL)
					} else {
						client.Logf(elog.LogLevelInfo1, "redirected to %s", disconn.Args)
					}
					break

				case elvin.DisconnReasonClientConnectionLost:
					client.Logf(elog.LogLevelInfo1, "FIXME: connection lost")
					os.Exit(1)

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
