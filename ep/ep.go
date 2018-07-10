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
	"github.com/cobaro/elvin/elog"
	"github.com/cobaro/elvin/elvin"
	"os"
	"os/signal"
)

type arguments struct {
	help              bool
	verbosity         int
	url               string
	unotify           bool
	number            int
	multiplier        int
	producerKeyString string
	producerKeyHex    string
	consumerKeyString string
	consumerKeyHex    string
	secureDelivery    bool
}

func main() {
	// Parse command line options
	args := flags()

	ep := elvin.NewClient(args.url, nil, nil, nil)
	ep.SetLogDateFormat(elog.LogDateLocaltime)
	ep.SetLogLevel(args.verbosity)

	// Process security arguments
	// FIXME; How compatible to be here? for now not very
	// !multiple args
	// !smart security
	// !load from file
	// stick with sha1 instead of 256?
	// producerKeyBlock := make(map[int]elvin.KeySetList)
	// consumerKeyBlock := make(map[int]elvin.KeySetList)
	ep.KeysNfn = make(map[int]elvin.KeySetList)
	var producerKeySet elvin.KeySet
	var consumerKeySet elvin.KeySet

	if len(args.producerKeyString) > 0 {
		elvin.KeySetAddKey(&producerKeySet, []byte(args.producerKeyString))
	}
	if len(args.producerKeyHex) > 0 {
		if key, err := hex.DecodeString(args.producerKeyHex); err != nil {
			ep.Logf(elog.LogLevelError, "Failed to interpret keyhex: %v", err)
			os.Exit(1)
		} else {
			elvin.KeySetAddKey(&producerKeySet, key)
		}
	}
	ep.KeysNfn[elvin.KeySchemeSha1Producer] = elvin.KeySetList{producerKeySet}

	if len(args.consumerKeyString) > 0 {
		// This is totally bogus as it can't be primed and we're not handling files
		ep.Logf(elog.LogLevelError, "Opening keyfiles not supported yet")
		os.Exit(1)
	}
	if len(args.consumerKeyHex) > 0 {
		if key, err := hex.DecodeString(args.consumerKeyHex); err != nil {
			ep.Logf(elog.LogLevelError, "Failed to interpret keyhex: %v", err)
			os.Exit(1)
		} else {
			elvin.KeySetAddKey(&consumerKeySet, key)
		}
	}
	ep.KeysNfn[elvin.KeySchemeSha1Consumer] = elvin.KeySetList{consumerKeySet}
	ep.Logf(elog.LogLevelDebug2, "secureDelivery is: %v, keys: %v", args.secureDelivery, ep.KeysNfn)

	// Set up sigint
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)

	// Connect if we not using unotify
	if !args.unotify {
		if err := ep.Connect(); err != nil {
			ep.Logf(elog.LogLevelInfo1, "%v", err)
			os.Exit(1)
		}
		ep.Logf(elog.LogLevelInfo1, "connected to %s", args.url)
	}

	// Grab a channel of notifications from our Parser
	notifications := elvin.ParseNotifications(os.Stdin, os.Stderr, args.multiplier, ep.LogFunc())

Loop:
	for {
		select {
		case notification, more := <-notifications:
			if more {
				// ep.Logf(elog.LogLevelInfo1, "read %+v", notification)

				for i := 0; i < args.number; i++ {
					if args.unotify {
						if err := ep.UNotify(notification, !args.secureDelivery, ep.KeysNfn); err != nil {
							ep.Logf(elog.LogLevelInfo1, "UNotify failed: %v", err)
						}
					} else {
						if err := ep.Notify(notification, !args.secureDelivery, nil); err != nil {
							ep.Logf(elog.LogLevelInfo1, "Notify failed: %v", err)
						}
					}
				}
			} else {
				ep.Logf(elog.LogLevelInfo2, "Exiting")
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
	flag.StringVar(&args.url, "e", "elvin://", "elvin url e.g. elvin://host")
	flag.IntVar(&args.verbosity, "v", 3, "verbosity (default 3)")
	flag.IntVar(&args.number, "n", 1, "number of notifications to send")
	flag.IntVar(&args.multiplier, "m", 0, "speed increase when replaying ec log")
	flag.BoolVar(&args.unotify, "unotify", false, "send using UNotify")
	flag.StringVar(&args.producerKeyString, "p", "", "SHA1 producer private key (string) ")
	flag.StringVar(&args.producerKeyHex, "P", "", "SHA1 producer private key (hex)")
	flag.StringVar(&args.consumerKeyString, "c", "", "SHA1 consumer public key (string) ")
	flag.StringVar(&args.consumerKeyHex, "C", "", "SHA1 consumer public key (hex)")
	flag.BoolVar(&args.secureDelivery, "x", false, "Don't allow insecure delivery (default is to allow)")
	flag.Parse()

	if args.help {
		flag.Usage()
		os.Exit(0)
	}

	return args
}
