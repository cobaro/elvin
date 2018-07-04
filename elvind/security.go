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
	"bytes"
	//"fmt"
	"github.com/cobaro/elvin/elvin"
)

// Producer scheme notification keys are sent in the clear
// Consumer scheme subscription keys are sent in the clear
//
// As a connection is assumed to be long lived, any connection
// keys are primed (run through one way function) on arrival
// and the original discarded
//
// As a subscription is assumed to be long lived, any subscription
// keys are primed (run through one way function) on arrival
// and the original discarded
//
// Connection keys were that way but are stored primed
// to avoid calculating them every time.
//
// FIXME: There is an optimization at some point where for producer
// and consumer connections keys we might mark (cache)a pair as
// matching updating as needed.

// Prime (run through one way func) a KeyBlock

// Prime a consumer
func PrimeConsumer(keys elvin.KeyBlock) {
	for scheme, ksl := range keys {
		switch scheme {
		case elvin.KeySchemeSha1Consumer:
			for i, raw := range ksl[0] {
				ksl[0][i] = elvin.PrimeSha1(raw)
			}
		case elvin.KeySchemeSha1Dual:
			for i, raw := range ksl[1] {
				ksl[1][i] = elvin.PrimeSha1(raw)
			}
		case elvin.KeySchemeSha256Consumer:
			for i, raw := range ksl[0] {
				ksl[0][i] = elvin.PrimeSha256(raw)
			}
		case elvin.KeySchemeSha256Dual:
			for i, raw := range ksl[1] {
				ksl[1][i] = elvin.PrimeSha256(raw)
			}
		}
	}
}

// Prime a producer
func PrimeProducer(keys elvin.KeyBlock) {
	for scheme, ksl := range keys {
		switch scheme {
		case elvin.KeySchemeSha1Producer:
			fallthrough
		case elvin.KeySchemeSha1Dual:
			for i, raw := range ksl[0] {
				ksl[0][i] = elvin.PrimeSha1(raw)
			}
		case elvin.KeySchemeSha256Producer:
			fallthrough
		case elvin.KeySchemeSha256Dual:
			for i, raw := range ksl[0] {
				ksl[0][i] = elvin.PrimeSha256(raw)
			}
		}
	}
}

// Do the keys match
// i.e, amongst the notifications and producer's keys does one of
// our schemes succeed agains the subscriptions and consumer's keys.
func SecurityMatches(nfn elvin.NotifyEmit, sub Subscription, pKeys, cKeys elvin.KeyBlock) bool {

	// Start with the simple (and hence common) cases
	//fmt.Printf("%v\n%v\n%v\n%v\n", nfn, sub, pKeys, cKeys)

	// No-one cares
	if nfn.DeliverInsecure && sub.AcceptInsecure {
		return true
	}

	// No producer keys, so determined by subscriber
	if len(nfn.Keys) == 0 && len(pKeys) == 0 {
		return sub.AcceptInsecure
	}

	// No consumer keys so determined by producer
	if len(sub.Keys) == 0 && len(cKeys) == 0 {
		return nfn.DeliverInsecure
	}

	// We got here cos we have to deal with keys

	// We could merge togther the connection and notification/subscription
	// keys but it's simpler just to run the combinations

	if len(nfn.Keys) > 0 {
		// A Match is always true
		if KeyBlocksMatches(nfn.Keys, sub.Keys) || KeyBlocksMatches(nfn.Keys, cKeys) {
			return true
		}
	}
	if len(pKeys) > 0 {
		// A Match is always true
		if KeyBlocksMatches(pKeys, sub.Keys) || KeyBlocksMatches(pKeys, cKeys) {
			return true
		}
	}

	// No matches so it shall not not pass
	return false
}

func KeyBlocksMatches(producer, consumer elvin.KeyBlock) bool {
	if len(producer) == 0 || len(consumer) == 0 {
		return false
	}
	for scheme, ksl := range producer {
		switch scheme {
		case elvin.KeySchemeSha1Dual:
			if KeySetMatches(ksl[0], consumer[elvin.KeySchemeSha1Dual][1]) && KeySetMatches(ksl[1], consumer[elvin.KeySchemeSha1Dual][0]) {
				return true
			}
		case elvin.KeySchemeSha1Producer:
			if KeySetMatches(ksl[0], consumer[elvin.KeySchemeSha1Producer][0]) {
				return true
			}
		case elvin.KeySchemeSha1Consumer:
			if KeySetMatches(ksl[0], consumer[elvin.KeySchemeSha1Consumer][0]) {
				return true
			}
		case elvin.KeySchemeSha256Dual:
		case elvin.KeySchemeSha256Producer:
		case elvin.KeySchemeSha256Consumer:
		}
	}

	return false
}

// A match occurs if there is a match across any of the two sets of keys
func KeySetMatches(first, second elvin.KeySet) bool {
	for _, f := range first {
		for _, s := range second {
			if bytes.Equal(f, s) {
				return true
			}
		}
	}
	return false
}
