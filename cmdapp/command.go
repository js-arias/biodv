// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// This work is derived from the go tool source code
// Copyright 2011 The Go Authors.  All rights reserved.

package cmdapp

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
)

// A Command is an implementation of a hosted command.
type Command struct {
	// Run runs the command.
	// It returns an error if the command finish on error.
	// The argument list is the set of arguments
	// unparsed by flag package.
	Run func(c *Command, args []string) error

	// UsageLine is the usage message.
	// The first word in the line is taken to be the command name.
	UsageLine string

	// Short is a short description of the command.
	Short string

	// Long is the long message shown in 'help <this-command>' output.
	Long string

	// Flag is the set of flags specific to this command.
	Flag *flag.FlagSet

	// RegisterFlags sets command-specific flags.
	RegisterFlags func(c *Command)
}

// Name returns the command's name:
// the first word in the is the usage line.
func (c *Command) Name() string {
	name := c.UsageLine
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

// Usage prints the usage message and exits the program.
func (c *Command) Usage() {
	fmt.Fprintf(os.Stderr, "usage: %s %s\n\n", Name, c.UsageLine)
	fmt.Fprintf(os.Stderr, "Type '%s help %s' for more information\n", Name, c.Name())
	os.Exit(1)
}

// Documentation prints command documentation.
func (c *Command) documentation(w io.Writer) {
	fmt.Fprintf(w, "%s\n\n", capitalize(c.Short))
	if c.Run != nil {
		fmt.Fprintf(w, "Usage:\n\n\t%s %s\n\n", Name, c.UsageLine)
	}
	fmt.Fprintf(w, "%s\n\n", strings.TrimSpace(c.Long))
}

// Runnable returns true is the command is a runnable command.
func (c *Command) runnable() bool {
	return c.Run != nil
}

// Capitalize sets the first rune of a string as upper case.
func capitalize(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToTitle(r)) + s[n:]
}
