// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package ed implements the rec.ed command,
// i.e. edit records interactively.
package ed

import (
	"fmt"
	"os"

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

   h [<command>]
   help [<command>]
     Print command help.

    q
    quit
      Quit the program, without making any change.
	`,
	Run: run,
}

func init() {
	cmdapp.Add(cmd)
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

	i := cmdapp.NewInter(os.Stdin)
	i.Add(&cmdapp.Cmd{"q", "quit", "quit the program", quitHelp, func([]string) bool { return true }})
	i.Prompt = prompt(tax, rec)

	i.Loop()
	return nil
}

func prompt(tax biodv.Taxon, rec *records.Record) string {
	var p = "root:"
	if tax != nil {
		p = fmt.Sprintf("%s:", tax.Name())
	}
	if rec != nil {
		p += fmt.Sprintf("%s:", rec.ID())
	}
	return p
}

var quitHelp = `
Usage:
    q
    quit
Ends the program without saving any change.
`
