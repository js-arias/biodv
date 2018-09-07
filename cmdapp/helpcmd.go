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
	"sort"
	"strings"

	"github.com/pkg/errors"
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
		return errors.Errorf("%s: too many arguments", c.Name())
	}

	arg := args[0]

	// 'help documentation' generates doc.go
	if arg == "documentation" {
		f, err := os.Create("doc.go")
		if err != nil {
			return errors.Wrap(err, c.Name())
		}
		fmt.Fprintf(f, "%s\n", strings.TrimSpace(goHead))
		printUsage(f)
		mutex.Lock()
		defer mutex.Unlock()
		var cmds []string
		for cn := range commands {
			cmds = append(cmds, cn)
		}
		sort.Strings(cmds)

		for _, c := range cmds {
			commands[c].documentation(f)
		}
		fmt.Fprintf(f, "\n%s", strings.TrimSpace(goFoot))
		if err := f.Close(); err != nil {
			return errors.Wrap(err, c.Name())
		}
		return nil
	}

	cn := getCommand(arg)
	if cn == nil {
		return errors.Errorf("%s: unknown help topic '%s'. Run '%s help'.\n", c.Name(), arg, Name)
	}
	cn.documentation(os.Stdout)
	return nil
}

var goHead = `// Authomatically generated doc.go file for use with godoc.

/*`

var goFoot = `*/
package main`
