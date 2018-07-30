// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// This work is derived from the go tool source code
// Copyright 2011 The Go Authors.  All rights reserved.

// Package cmdapp
// implements a command line application
// that host a set of commands
// as in the go tool and git.
package cmdapp

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Name stores the application name,
// the default is based on the arguments of the program.
var Name = filepath.Base(os.Args[0])

// Short is a short description of the application.
var Short string

// Main runs the application.
func Main() {
	flag.Usage = usage
	flag.Parse()

	usage()
}

// Usage prints application help and exists.
func usage() {
	printUsage(os.Stderr)
	os.Exit(1)
}

// PrintUsage prints the application usage help.
func printUsage(w io.Writer) {
	fmt.Fprintf(w, "%s\n\n", Short)
}
