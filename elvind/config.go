// Copyright 2018 Cobaro Pty Ltd. All Rights Reserved.
// Copyright 2013-2018 https://stackoverflow.com/questions/16465705/how-to-handle-configuration-in-go

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
	"encoding/json"
	"os"
)

type Protocol struct {
	Network string
	Marshal string
	Address string
}

type Configuration struct {
	Protocols        []Protocol
	FailoverHosts    []Protocol
	DoFailover       bool
	MaxConnections   int
	TestConnInterval int64 // idle seconds to trigger, 0 to disable
	TestConnTimeout  int64 // Time to await a response
}

func LoadConfig(configFile string) (config *Configuration, err error) {
	file, err := os.Open(configFile)
	defer file.Close()
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err = decoder.Decode(&configuration)
	return &configuration, err
}
