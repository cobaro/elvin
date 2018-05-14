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
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

// Read an Elvin subscription from the command-line, stdin, or a file.
// Parse the subscription, and report any errors.  If no errors are
// found, print the subscription as a parse tree.

// Usage: esl [-f file-name | - | ...]
// Read from file (-f), stdin (-) or command line arguments.

func main() {

	// Process arguments.
	var fileName = flag.String("input", "", "Input file name")
	var help = flag.Bool("help", false, "Request help with options")
	var version = flag.Bool("version", false, "Print version info")

	flag.Parse()

	if *version {
		fmt.Println("0.0.1")
		os.Exit(0)
	}

	if *help {
		fmt.Println("[-f filename | - | ... ]")
		os.Exit(0)
	}

	var subExpr string

	if flag.Arg(0) == "-" {
		buffer, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Println("Error reading from stdin")
			os.Exit(1)
		}

		subExpr = string(buffer)

	} else if *fileName != "" {
		buffer, err := ioutil.ReadFile(*fileName)
		if err != nil {
			fmt.Println("Error reading file")
			os.Exit(1)
		}

		subExpr = string(buffer)

	} else {
		for _, arg := range flag.Args() {
			subExpr += " " + arg
		}
	}

	// Create parser.

	// Parse content, and run reductions.

	fmt.Println(subExpr)

	lexer(subExpr)

	os.Exit(0)
}
