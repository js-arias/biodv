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
	"strings"
	"sync"
)

// Name stores the application name,
// the default is based on the arguments of the program.
var Name = filepath.Base(os.Args[0])

// Short is a short description of the application.
var Short string

// Commands is the list of available commands
// and help topics.
var (
	mutex    sync.Mutex
	commands = make(map[string]*Command)
)

// Add adds a new command to the application.
// Command names should be unique,
// otherwise it will trigger a panic.
func Add(c *Command) {
	name := strings.ToLower(c.Name())
	if name == "" {
		msg := fmt.Sprintf("cmdapp: Empty command name: %s", c.Short)
		panic(msg)
	}
	if getCmd(name) != nil {
		msg := fmt.Sprintf("cmdapp: Repeated command name: %s %s", name, c.Short)
		panic(msg)
	}
	mutex.Lock()
	defer mutex.Unlock()
	commands[name] = c
}

// GetCmd returns a command with a given name.
func getCmd(name string) *Command {
	name = strings.ToLower(name)
	mutex.Lock()
	defer mutex.Unlock()
	return commands[name]
}

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
