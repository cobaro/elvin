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

package elvin

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
)

// The API uses KeyBlocks for securing subscriptions and notifications
//
// A KeyBlock is a map of KeySetLists each of which is of a particular
// KeyScheme and may contain only one KeySetList of a particular scheme
//
// A KeySetList is an ordered list of KeySets, represented as a slice
//
// A KeySet is an unordered set of Keys, represented as a slice
//
// A Key is either private (aka raw), or public (aka cooked) such that
// a private key is converted to a public key via a one-way function
// (hash) such as SHA-256.
//
// A scheme can be producer, consumer, or dual
//
// A producer scheme is when the producer has a secret and hashes it
// giving that to consumers. The producer may decide whether the
// notification is available to consumers without the key hash
// (DeliverInsecure) and the consumer may decide whether to accept a
// message without a matching Key (AcceptInsecure).
//
// A consumer scheme is the reverse where the consumer has a private
// key and gives the public key (hash) to producers.
//
// A Dual scheme is one where the producer and consumer schemes
// are combined. On this scheme the KeySetLists must be symmetric,
// producer first, then the consumer. There must be at least one of
// each.
//
// As a shortcutting mechanism and a means of saving bandwidth on
// individual messages it is possible to also have these mechanisms
// attached to a connection, rather than each message. Connection
// level security acts in the same manner but are then not required on
// each message.  Connection level and message level schemes may be
// used in tandem and augment each other.

// Supported key schemes
const (
	KeySchemeSha1Dual       = 1 // (deprecated) SHA-1 dual
	KeySchemeSha1Producer   = 2 // (deprecated) SHA-1 producer
	KeySchemeSha1Consumer   = 3 // (deprecated) SHA-1 consumer
	KeySchemeSha256Dual     = 7 // SHA-256 dual
	KeySchemeSha256Producer = 8 // SHA-256 producer
	KeySchemeSha256Consumer = 9 // SHA-256 consumer
)

// KeySetList indices
const (
	KeySetProducer = 0
	KeySetConsumer = 1
)

type Key []byte                  // A single key
type KeySet []Key                // An unordered set of keys
type KeySetList []KeySet         // indexed numerically (dual ordering matters)
type KeyBlock map[int]KeySetList // This is what notify/subscribe use

// Prime the key by running it the sha1()
func PrimeSha1(in Key) Key {
	// can't slice something that's not assigned to a variable
	// and we want to convert it from a fixed size
	tmp := sha1.Sum(in)
	return tmp[:]
}

// Prime the key by running it the sha1()
func PrimeSha256(in Key) Key {
	// can't slice something that's not assigned to a variable
	// and we want to convert it from a fixed size
	tmp := sha256.Sum256(in)
	return tmp[:]
}

// Add a key to a keyset.
// Duplicates are not added but this is not an error/
func KeySetAddKey(keyset *KeySet, new Key) {
	if keyset == nil {
		return // bail
	}

	for _, old := range *keyset {
		if bytes.Equal(old, new) {
			return
		}
	}
	*keyset = append(*keyset, new)
}

// Delete a key to a keyset.
// Duplicates are not added but this is not an error/
func KeySetDeleteKey(keyset *KeySet, delete Key) {
	if keyset == nil {
		return // bail
	}

	for i, old := range *keyset {
		if bytes.Equal(old, delete) {
			(*keyset)[i] = (*keyset)[len(*keyset)-1]
			*keyset = (*keyset)[:len(*keyset)-1]
			return
		}
	}
	return
}

// Add the keys in the second KeyBlock to the existing
// Does not prime keys.
// Duplicates are simple ignored.
// The dual schemes have two keysets where producer and consumer have only one
func KeyBlockAddKeys(existing KeyBlock, add KeyBlock) {
	if existing == nil {
		return // bail
	}

	for scheme, kslAdd := range add {
		switch scheme {
		case KeySchemeSha1Dual:
			fallthrough
		case KeySchemeSha256Dual:
			// If we don't have this scheme already then we can
			// just copy it in, otherwise check every key
			if kslExisting, ok := existing[scheme]; !ok {
				existing[scheme] = add[scheme]
			} else {
				for _, keyAdd := range kslAdd[KeySetConsumer] {
					KeySetAddKey(&kslExisting[KeySetConsumer], keyAdd)
					continue
				}
			}
			fallthrough
		case KeySchemeSha1Producer:
			fallthrough
		case KeySchemeSha1Consumer:
			fallthrough
		case KeySchemeSha256Producer:
			fallthrough
		case KeySchemeSha256Consumer:
			// If we don't have this scheme already then we can
			// just copy it in, otherwise check every key
			if kslExisting, ok := existing[scheme]; !ok {
				existing[scheme] = add[scheme]
			} else {
				for _, keyAdd := range kslAdd[KeySetProducer] {
					KeySetAddKey(&kslExisting[KeySetProducer], keyAdd)
				}
			}
		}
	}
}
