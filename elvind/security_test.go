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
	"encoding/hex"
	"github.com/cobaro/elvin/elvin"
	"testing"
)

var k1 = []byte("foo")
var k2 = []byte("bar")
var k3 = []byte("baz")

var k1SHA1, k2SHA1, k3SHA1 []byte

var k1SHA256, k2SHA256, k3SHA256 []byte

var namevalue map[string]interface{}

func init() {
	// Hash the keys
	k1SHA1 = elvin.PrimeSha1(k1)
	k2SHA1 = elvin.PrimeSha1(k2)
	k3SHA1 = elvin.PrimeSha1(k3)

	k1SHA256 = elvin.PrimeSha256(k1)
	k2SHA256 = elvin.PrimeSha256(k2)
	k3SHA256 = elvin.PrimeSha256(k3)

	// namevalue initilization
	namevalue = make(map[string]interface{})
	namevalue["foo"] = 1
}

// Test the Prime function used to convert keys in the router
func TestPrime(t *testing.T) {

	// SHA1 Producer
	var producerKeySet elvin.KeySet
	producerKeySet = append(producerKeySet, k1)
	producerKeySetList := elvin.KeySetList{producerKeySet}
	producerKeyBlock := make(map[int]elvin.KeySetList)
	producerKeyBlock[elvin.KeySchemeSha1Producer] = producerKeySetList

	// Should not change the producer key
	PrimeConsumer(producerKeyBlock)
	if !bytes.Equal(k1, producerKeyBlock[elvin.KeySchemeSha1Producer][0][0]) {
		t.Fatalf("Prime() wrongly changed a producer key (%s->%s)", hex.EncodeToString(k1), hex.EncodeToString(producerKeyBlock[elvin.KeySchemeSha1Producer][0][0]))
	}

	// Should change the producer key
	PrimeProducer(producerKeyBlock)
	if bytes.Equal(k1, producerKeyBlock[elvin.KeySchemeSha1Producer][0][0]) {
		t.Fatalf("Prime() didn't change the producer key (%s->%s)", hex.EncodeToString(k1), hex.EncodeToString(producerKeyBlock[elvin.KeySchemeSha1Producer][0][0]))
	}
	if !bytes.Equal(producerKeyBlock[elvin.KeySchemeSha1Producer][0][0], k1SHA1) {
		t.Fatalf("Prime() changes the producer key to a bad value:\n%s\n%s", hex.EncodeToString(k1SHA1), hex.EncodeToString(producerKeyBlock[elvin.KeySchemeSha1Producer][0][0]))
	}

	// SHA1 Consumer
	var consumerKeySet elvin.KeySet
	consumerKeySet = append(consumerKeySet, k1)
	consumerKeySetList := elvin.KeySetList{consumerKeySet}
	consumerKeyBlock := make(map[int]elvin.KeySetList)
	consumerKeyBlock[elvin.KeySchemeSha1Consumer] = consumerKeySetList

	// Should not change the consumer key
	PrimeProducer(consumerKeyBlock)
	if !bytes.Equal(k1, consumerKeyBlock[elvin.KeySchemeSha1Consumer][0][0]) {
		t.Fatalf("Prime() wrongly changed a consumer key (%s->%s)", hex.EncodeToString(k1), hex.EncodeToString(consumerKeyBlock[elvin.KeySchemeSha1Consumer][0][0]))
	}

	// Should change the consumer key
	PrimeConsumer(consumerKeyBlock)
	if bytes.Equal(k1, consumerKeyBlock[elvin.KeySchemeSha1Consumer][0][0]) {
		t.Fatalf("Prime() didn't change the consumer key (%s->%s)", hex.EncodeToString(k1), hex.EncodeToString(consumerKeyBlock[elvin.KeySchemeSha1Consumer][0][0]))
	}
	if !bytes.Equal(consumerKeyBlock[elvin.KeySchemeSha1Consumer][0][0], k1SHA1) {
		t.Fatalf("Prime() changes the consumer key to a bad value:\n%s\n%s", hex.EncodeToString(k1SHA1), hex.EncodeToString(consumerKeyBlock[elvin.KeySchemeSha1Consumer][0][0]))
	}

	// SHA1 Dual
	producerKeySet = nil // reset
	consumerKeySet = nil
	producerKeySet = append(producerKeySet, k1)
	consumerKeySet = append(consumerKeySet, k1)
	keySetList := elvin.KeySetList{producerKeySet, consumerKeySet}

	keyBlock := make(map[int]elvin.KeySetList)

	keyBlock[elvin.KeySchemeSha1Dual] = keySetList

	// Should change the consumer but not the producer
	PrimeConsumer(keyBlock)
	if bytes.Equal(keyBlock[elvin.KeySchemeSha1Dual][0][0], keyBlock[elvin.KeySchemeSha1Dual][1][0]) {
		t.Fatalf("Prime() wrongly changed a producer key (%s->%s)", hex.EncodeToString(k1), hex.EncodeToString(keyBlock[elvin.KeySchemeSha1Dual][0][0]))
	}

	// Should change the producer not the consumer so now equal
	PrimeProducer(keyBlock)
	if !bytes.Equal(keyBlock[elvin.KeySchemeSha1Dual][0][0], keyBlock[elvin.KeySchemeSha1Dual][1][0]) {
		t.Fatalf("Producer and consumer show now match %s,%s", hex.EncodeToString(keyBlock[elvin.KeySchemeSha1Dual][0][0]), hex.EncodeToString(keyBlock[elvin.KeySchemeSha1Dual][0][1]))
	}

	// SHA256 Dual
	producerKeySet = nil // reset
	consumerKeySet = nil
	producerKeySet = append(producerKeySet, k1)
	consumerKeySet = append(consumerKeySet, k1)
	keySetList = elvin.KeySetList{producerKeySet, consumerKeySet}

	keyBlock = make(map[int]elvin.KeySetList)

	keyBlock[elvin.KeySchemeSha256Dual] = keySetList

	// Should change the consumer but not the producer
	PrimeConsumer(keyBlock)
	if bytes.Equal(keyBlock[elvin.KeySchemeSha256Dual][0][0], keyBlock[elvin.KeySchemeSha256Dual][1][0]) {
		t.Fatalf("Prime() wrongly changed a producer key (%s->%s)", hex.EncodeToString(k1), hex.EncodeToString(keyBlock[elvin.KeySchemeSha256Dual][0][0]))
	}

	// Should change the producer not the consumer so now equal
	PrimeProducer(keyBlock)
	if !bytes.Equal(keyBlock[elvin.KeySchemeSha256Dual][0][0], keyBlock[elvin.KeySchemeSha256Dual][1][0]) {
		t.Fatalf("Producer and consumer should now match %s,%s", hex.EncodeToString(keyBlock[elvin.KeySchemeSha256Dual][0][0]), hex.EncodeToString(keyBlock[elvin.KeySchemeSha256Dual][0][1]))
	}
}

// Test some matching
func TestMatching(t *testing.T) {

	// Producer keyblock
	var producerKeySet elvin.KeySet
	producerKeySet = append(producerKeySet, k1)
	producerKeySetList := elvin.KeySetList{producerKeySet}
	producerKeyBlock := make(map[int]elvin.KeySetList)
	producerKeyBlock[elvin.KeySchemeSha1Producer] = producerKeySetList

	// Make a notification with that key block that must match
	nfn := elvin.NotifyEmit{namevalue, false, producerKeyBlock}

	// Consumer keyblock
	var consumerKeySet elvin.KeySet
	consumerKeySet = append(consumerKeySet, elvin.PrimeSha1(k1))
	consumerKeySetList := elvin.KeySetList{consumerKeySet}
	consumerKeyBlock := make(map[int]elvin.KeySetList)
	consumerKeyBlock[elvin.KeySchemeSha1Producer] = consumerKeySetList

	// Make s subscription with that keyBlock that must match
	sub := Subscription{1, false, consumerKeyBlock, nil}

	// Because the producer key is not yet primed, these should not match
	if SecurityMatches(nfn, sub, nil, nil) {
		t.Fatalf("unprimed producer should not match")
	}

	nfn.Keys = nil
	if SecurityMatches(nfn, sub, producerKeyBlock, nil) {
		t.Fatalf("unprimed producer should not match")
	}

	sub.Keys = nil
	if SecurityMatches(nfn, sub, nil, consumerKeyBlock) {
		t.Fatalf("unprimed producer should not match")
	}

	// Prime the producer key and those tests should all work
	producerKeySet[0] = elvin.PrimeSha1(producerKeySet[0])
	if !SecurityMatches(nfn, sub, producerKeyBlock, consumerKeyBlock) {
		t.Fatalf("primed producer should match")
	}

	sub.Keys = consumerKeyBlock
	if !SecurityMatches(nfn, sub, producerKeyBlock, nil) {
		t.Fatalf("primed producer should match")
	}

	nfn.Keys = producerKeyBlock
	if !SecurityMatches(nfn, sub, nil, nil) {
		t.Fatalf("primed producer should match")
	}

	// Now test our matching across some KeyBlock actions
	sub.Keys = make(map[int]elvin.KeySetList)
	if SecurityMatches(nfn, sub, nil, nil) {
		t.Fatalf("empty subscriber keys should not match")
	}
	elvin.KeyBlockAddKeys(sub.Keys, nfn.Keys)
	if !SecurityMatches(nfn, sub, nil, nil) {
		t.Fatalf("same keys should not match")
	}

	elvin.KeyBlockDeleteKeys(sub.Keys, nfn.Keys)
	if SecurityMatches(nfn, sub, nil, nil) {
		t.Fatalf("empty subscriber keys should not match")
	}
}
