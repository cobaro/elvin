// -*- mode: go -*-

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
