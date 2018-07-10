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
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

const DefaultNameValueTimeFormat = "2006-01-02T15:04:05.999999999-0700"

// Pretty print a NameValue in a standardized format
// separator is appended to the output
// If timsta
func NameValueToString(nv map[string]interface{}, separator string, timeFormat string) (string, error) {
	var sb strings.Builder

	if len(timeFormat) > 0 {
		now := time.Now()
		fmt.Fprintf(&sb, "$time %s\n", now.Format(timeFormat))
	}

	for name, value := range nv {
		fmt.Fprintf(&sb, "%s: ", name)

		switch value.(type) {
		case int32:
			fmt.Fprintf(&sb, "%d\n", value)
		case int64:
			fmt.Fprintf(&sb, "%d\n", value)
		case float64:
			fmt.Fprintf(&sb, "%e\n", value)
		case string:
			fmt.Fprintf(&sb, "\"%s\"\n", value)
		case []byte:
			fmt.Fprintf(&sb, "[%s]\n", hex.EncodeToString(value.([]uint8)))
		default:
			return "", fmt.Errorf("Bad *type* of %v in %v", value, nv)
		}
	}

	// Add any separator
	fmt.Fprintf(&sb, separator)

	return sb.String(), nil
}

// Read's notifications from input stream into an output channel.
// Exits on EOF closing the output channel.
// Set out to io.discard for quiet (on failure) operation.
// If replayMultiple is 0, then it will ignore timestamps. Otherwise
// it will look at the difference in timestamps ($time) and replay up
// to the desired multiple.
func ParseNotifications(in io.Reader, out io.Writer, replayMultiple int, logf func(io.Writer, string, ...interface{}) (int, error)) chan map[string]interface{} {

	var prevTime, thisTime time.Time
	var err error
	scanner := bufio.NewScanner(in)
	channel := make(chan map[string]interface{})

	go func() {
		nfn := make(map[string]interface{})

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if len(line) < 3 {
				// There is no valid input this small but
				// forgive empties
				if len(line) > 0 {
					logf(out, "Unrecognised input to parse '%s' as attribute: value\n", line)
				}
			} else if line[:3] == "---" {
				// Look for end of message marker '^---.*$'
				if len(nfn) > 0 {
					channel <- nfn
					nfn = make(map[string]interface{}) // reset
				}
			} else if len(line) >= 6 && line[:6] == "$time " {
				// If we're replaying then perhaps we should delay

				// Extract the timestamp
				prevTime = thisTime
				thisTime, err = time.Parse(DefaultNameValueTimeFormat, line[6:])
				if err != nil {
					logf(out, "Failed to parse '%s' as $time\n", line)
				}

				// And if we should delay then do so
				if !prevTime.IsZero() && replayMultiple > 0 {
					time.Sleep(thisTime.Sub(prevTime) / time.Duration(replayMultiple))
				}

			} else {
				// look for name : value (with or without space around :)
				namevalue := strings.SplitN(line, ":", 2)
				if len(namevalue) != 2 {
					logf(out, "Failed to parse '%s' as attribute: value\n", line)
				} else {
					// Try to convert the value
					name := strings.TrimSpace(namevalue[0])
					value := strings.TrimSpace(namevalue[1])

					if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
						// string "delimited"
						nfn[name] = value[1 : len(value)-1]

					} else if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
						// opaque [delimited]

						// It's optional to allow white space between hex digits
						s := strings.Join(strings.Split(value[1:len(value)-1], " "), "")
						size := len(s) / 2
						opaque := make([]byte, size)
						len, err := hex.Decode(opaque, []byte(s[:]))
						if err != nil {
							logf(out, "ParseError: %s\n", err.Error())

						} else if size != len {
							logf(out, "ParseError: Couldn't convert entirety of %s", value)
						} else {
							nfn[name] = opaque
						}
					} else if strings.HasSuffix(value, "L") || strings.HasSuffix(value, "l") {
						// int64 e.g., 123L
						if i64, err := strconv.ParseInt(value[:len(value)-1], 10, 64); err != nil {
							logf(out, "ParseError: converting '%s' to int64: %v\n", value, err.Error())
						} else {
							nfn[name] = i64
						}
					} else if strings.Contains(value, ".") {
						// float64 e.g. 3.14
						if f64, err := strconv.ParseFloat(value, 64); err != nil {
							logf(out, "ParseError: converting '%s' to float64: %v\n", value, err.Error())
						} else {
							nfn[name] = f64
						}
					} else if i64, err := strconv.ParseInt(value, 10, 32); err == nil {
						nfn[name] = int32(i64)
					} else if i64, err := strconv.ParseInt(value, 10, 64); err == nil {
						nfn[name] = i64
					} else {
						logf(out, "Failed to parse %s\n", value)
					}
				}
			}
		}
		// EOF which is normal if a file is redirected in for example
		// so send it and return, exiting the parser
		if len(nfn) > 0 {
			channel <- nfn
		}
		close(channel)
	}()
	return channel
}
