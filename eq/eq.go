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
	number            int
	unotify           bool
	producerKeyString string
	producerKeyHex    string
	consumerKeyString string
	consumerKeyHex    string
	secureDelivery    bool
}

func main() {
	// Argument parsing
	args := flags()
	flag.Parse()

	eq := elvin.NewClient(args.url, nil, nil, nil)
	eq.SetLogDateFormat(elog.LogDateLocaltime)
	eq.SetLogLevel(args.verbosity)

	// Process security arguments
	// FIXME; How compatible to be here? for now not very
	// !multiple args
	// !smart security
	// !load from file
	// stick with sha1 instead of 256?
	// producerKeyBlock := make(map[int]elvin.KeySetList)
	// consumerKeyBlock := make(map[int]elvin.KeySetList)
	eq.KeysNfn = make(map[int]elvin.KeySetList)
	var producerKeySet elvin.KeySet
	var consumerKeySet elvin.KeySet

	if len(args.producerKeyString) > 0 {
		elvin.KeySetAddKey(&producerKeySet, []byte(args.producerKeyString))
	}
	if len(args.producerKeyHex) > 0 {
		if key, err := hex.DecodeString(args.producerKeyHex); err != nil {
			eq.Logf(elog.LogLevelError, "Failed to interpret keyhex: %v", err)
			os.Exit(1)
		} else {
			elvin.KeySetAddKey(&producerKeySet, key)
		}
	}
	eq.KeysNfn[elvin.KeySchemeSha1Producer] = elvin.KeySetList{producerKeySet}

	if len(args.consumerKeyString) > 0 {
		// This is totally bogus as it can't be primed and we're not handling files
		eq.Logf(elog.LogLevelError, "Opening keyfiles not supported yet")
		os.Exit(1)
	}
	if len(args.consumerKeyHex) > 0 {
		if key, err := hex.DecodeString(args.consumerKeyHex); err != nil {
			eq.Logf(elog.LogLevelError, "Failed to interpret keyhex: %v", err)
			os.Exit(1)
		} else {
			elvin.KeySetAddKey(&consumerKeySet, key)
		}
	}
	eq.KeysNfn[elvin.KeySchemeSha1Consumer] = elvin.KeySetList{consumerKeySet}
	eq.Logf(elog.LogLevelDebug2, "secureDelivery is: %v, keys: %v", args.secureDelivery, eq.KeysNfn)

	eq.Options = make(map[string]interface{})

	// FIXME: At some point let's formalize these as test cases
	// eq.Options["TestNack"] = 1
	// eq.Options["TestDisconn"] = 1
	// eq.Logf(elog.LogLevelInfo1, "Options:%v\n", eq.Options)

	if err := eq.Connect(); err != nil {
		eq.Logf(elog.LogLevelInfo1, "%v", err)
		os.Exit(1)
	}
	eq.Logf(elog.LogLevelInfo1, "connected to %s", args.url)

	// FIXME: do a NewSubscription()
	quench := new(elvin.Quench)
	quench.Names = map[string]bool{"int32": true, "float64": true}
	quench.DeliverInsecure = !args.secureDelivery
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

// Argument parsing
func flags() (args arguments) {
	flag.BoolVar(&args.help, "h", false, "Print this help")
	flag.StringVar(&args.url, "e", "elvin://", "elvin url e.g., elvin://host")
	flag.IntVar(&args.verbosity, "v", 3, "verbosity (default 3)")
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
