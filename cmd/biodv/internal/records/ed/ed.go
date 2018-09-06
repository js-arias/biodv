// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package ed implements the rec.ed command,
// i.e. edit records interactively.
package ed

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/cmdapp"
	"github.com/js-arias/biodv/records"

	"github.com/pkg/errors"
)

var cmd = &cmdapp.Command{
	UsageLine: "rec.ed",
	Short:     "edit records interactively",
	Long: `
Command rec.ed implements a simple interactive specimen record editor.

The commands understood by rec.ed are:

    quit
    q
      Quit the program, without making any change.
	`,
	Run: run,
}

func init() {
	cmdapp.Add(cmd)
}

type command struct {
	short string
	long  string
	about string
	run   func() bool
}

var commands = []command{
	{"q", "quit", "quit the program", func() bool { return true }},
}

func run(c *cmdapp.Command, args []string) error {
	_, err := biodv.OpenTax("biodv", "")
	if err != nil {
		return errors.Wrap(err, c.Name())
	}
	var tax biodv.Taxon

	_, err = records.Open("")
	if err != nil {
		return errors.Wrap(err, c.Name())
	}
	var rec *records.Record

	// Command loop
	r := bufio.NewReader(os.Stdin)
	for {
		prompt(tax, rec)
		cmdLine, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			fmt.Printf("error: %v\n", err)
		}
		cmds := strings.Fields(cmdLine)
		if len(cmds) == 0 {
			continue
		}
		cmd := strings.ToLower(cmds[0])
		var fn func() bool
		for _, c := range commands {
			if c.short == cmd || c.long == cmd {
				fn = c.run
				break
			}
		}
		if fn == nil {
			if cmd == "h" || cmd == "help" {
				help()
				continue
			}
			if x, _ := utf8.DecodeRuneInString(cmds[0]); x == '#' {
				continue
			}
			fmt.Printf("error: unknown command '%s'\n", cmds[0])
			continue
		}
		if fn() {
			break
		}
	}
	return nil
}

func prompt(tax biodv.Taxon, rec *records.Record) {
	if tax == nil {
		fmt.Printf("root:")
	} else {
		fmt.Printf("%s:", tax.Name())
	}
	if rec == nil {
		fmt.Printf(" ")
	} else {
		fmt.Printf("%s: ", rec.ID())
	}
}

func help() {
	for _, c := range commands {
		fmt.Printf("    %s, %-16s %s\n", c.short, c.long, c.about)
	}
}
