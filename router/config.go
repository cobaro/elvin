// Copyright (c) Cobaro Pty Ltd. All Rights Reserved. GPL-V3
// Copyright (c) https://stackoverflow.com/questions/16465705/how-to-handle-configuration-in-go

package main

import (
	"encoding/json"
	"os"
)

type Configuration struct {
	Protocols      []string
	MaxConnections int
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
