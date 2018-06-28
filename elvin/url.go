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
	"fmt"
	"strconv"
	"strings"
)

type Protocol struct {
	Network string
	Marshal string
	Address string
	Args    string
	Major   int
	Minor   int
}

func ProtocolToURL(protocol *Protocol) (url string) {
	var version, args string
	if protocol.Major == 0 && protocol.Minor == 0 {
		version = "4.1"
	} else {
		version = fmt.Sprintf("%d.%d", protocol.Major, protocol.Minor)
	}
	if len(protocol.Args) != 0 {
		args = "/" + protocol.Args
	}

	return fmt.Sprintf("elvin:%s/%s,%s/%s%s", version, protocol.Network, protocol.Marshal, protocol.Address, args)
}

// Convert a URL to a Protocol
func URLToProtocol(url string) (p *Protocol, err error) {
	protocol := new(Protocol)

	// Split the url into three or four pieces
	// 0. elvin:<version>    - version is optional
	// 1. /<network,marshal> - stack is optional, defaults to tcp,xdr
	// 2. /<host:port>       - address optional, defaults to localhost:2917
	// 3. </args>            - subscription/notification part

	splits := strings.Split(url, "/")
	if len(splits) < 3 {
		return nil, fmt.Errorf("Malformed url")
	}
	//fmt.Printf("%s -> (%d)%s\n", url, len(splits), splits)
	//fmt.Printf("(%d)%s\n", len(splits[0]), splits[0][:6])

	// 0. Check and strip leading 'elvin:' obtaining version if present
	length := len(splits[0])
	switch {
	case length < 6:
		return nil, fmt.Errorf("url must begin with 'elvin:' ")
	case length == 6:
		if splits[0][:6] != "elvin:" {
			return nil, fmt.Errorf("url must begin with 'elvin:' ")
		}
	default:
		versions := strings.Split(splits[0][6:], ".")
		if len(versions) != 2 {
			return nil, fmt.Errorf("Could not parse version number")
		}
		if protocol.Major, err = strconv.Atoi(versions[0]); err != nil {
			return nil, fmt.Errorf("Could not parse version number")
		}
		if protocol.Minor, err = strconv.Atoi(versions[1]); err != nil {
			return nil, fmt.Errorf("Could not parse version number")
		}
		// fmt.Printf("version is %d.%d\n", protocol.Major, protocol.Minor)
	}

	// 1. See if they specified a communications stack e.g., tcp,xdr
	if len(splits[1]) == 0 {
		protocol.Network = "tcp"
		protocol.Marshal = "xdr"
	} else {
		stacks := strings.Split(splits[1], ",")
		// There must be two or three: comms,<security,>marshal
		length := len(stacks)
		switch length {
		case 2:
			protocol.Network = stacks[0]
			protocol.Marshal = stacks[1]
		case 3:
			protocol.Network = stacks[0]
			// FIXME: protocol.Security = stacks[1]
			protocol.Marshal = stacks[2]
			// fmt.Printf("stack is %s,%s\n", protocol.Network, protocol.Marshal)
		default:
			return nil, fmt.Errorf("Could not parse protocol stack")
		}
	}

	// 2. <host:port> - address optional, defaults to localhost:2917
	// host can be an ipv4 or ipv6 address as well to add some complexity
	host := "localhost"
	port := 2917

	if len(splits[2]) != 0 { // otherwise default of localhost:2917
		if splits[2][0] == '[' { // ipv6
			ip6 := strings.Split(splits[2][1:], "]")
			if len(ip6) == 2 && ip6[1][0] == ':' {
				if port, err = strconv.Atoi(ip6[1][1:]); err != nil {
					return nil, fmt.Errorf("port is not a number")
				}
			}
			host = "[" + ip6[0] + "]"
		} else { // ipv4
			hostport := strings.Split(splits[2], ":")

			switch len(hostport) {
			case 1:
				host = hostport[0]
			case 2:
				if port, err = strconv.Atoi(hostport[1]); err != nil {
					return nil, fmt.Errorf("port is not a number")
				}
			default:
				return nil, fmt.Errorf("host:port parse failure")
			}
		}
	}
	protocol.Address = fmt.Sprintf("%s:%d", host, port)

	//3. args (optional)
	if len(splits) >= 3 {
		protocol.Args = strings.Join(splits[3:], "")
	}

	// fmt.Printf("return %v %v\n", protocol, err)
	return protocol, err
}
