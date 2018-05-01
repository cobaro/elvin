// Copyright (c) Cobaro Pty Ltd. All Rights Reserved. GPL-V3
// Copyright (c) https://stackoverflow.com/questions/16465705/how-to-handle-configuration-in-go

// This file is part of elvind
//
// elvind is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// elvind is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with elvind. If not, see <http://www.gnu.org/licenses/>.
// Copyright 2018 Cobaro Pty Ltd. All Rights Reserved.

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
