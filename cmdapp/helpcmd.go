// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// This work is derived from the go tool source code
// Copyright 2011 The Go Authors.  All rights reserved.

package cmdapp

import (
	"fmt"
	"os"
)

// Help is the help command.
var help = &Command{
	UsageLine: "help [<command>]",
	Long: `
Command help displays help information for a command or a help topic.

With no arguments it prints the list of available commands and help topics to
the standard output.
	`,
	Run: runHelp,
}

func runHelp(c *Command, args []string) error {
	if len(args) == 0 {
		printUsage(os.Stdout)
		return nil
	}
	if len(args) != 1 {
		return fmt.Errorf("%s: too many arguments", c.Name())
	}

	arg := args[0]

	cn := getCmd(arg)
	if cn == nil {
		return fmt.Errorf("%s: unknown help topic '%s'. Run '%s help'.\n", c.Name(), arg, Name)
	}
	cn.documentation(os.Stdout)
	return nil
}
