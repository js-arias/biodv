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
	"strings"

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

var txm biodv.Taxonomy
var tax biodv.Taxon

func run(c *cmdapp.Command, args []string) error {
	var err error
	txm, err = biodv.OpenTax("biodv", "")
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	_, err = records.Open("")
	if err != nil {
		return errors.Wrap(err, c.Name())
	}
	var rec *records.Record

	i := cmdapp.NewInter(os.Stdin)
	addCommands(i)
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

func addCommands(i *cmdapp.Inter) {
	i.Add(&cmdapp.Cmd{"l", "list", "list descendant taxons", listHelp, listCmd})
	i.Add(&cmdapp.Cmd{"q", "quit", "quit the program", quitHelp, func([]string) bool { return true }})
}

var quitHelp = `
Usage:
    q
    quit
Ends the program without saving any change.
`

var listHelp = `
Usage:
    l [<taxon>]
    list [<taxon>]
Without parameters shows the list of descendants of the current taxon.
If a taxon is given, it will show the descendants of the indicated
taxon.
`

func listCmd(args []string) bool {
	nm := strings.Join(args, " ")
	if nm == "" {
		if tax != nil {
			nm = tax.ID()
		}
	} else if nm == "/" {
		nm = ""
	} else if nm == "." {
		if tax == nil {
			return false
		}
		nm = tax.Parent()
	} else {
		nt, _ := txm.TaxID(nm)
		if nt == nil {
			return false
		}
		nm = nt.ID()
	}
	ls, _ := biodv.TaxList(txm.Children(nm))
	if nm != "" {
		syns, _ := biodv.TaxList(txm.Synonyms(nm))
		ls = append(ls, syns...)
	}
	for _, c := range ls {
		fmt.Printf("%s\n", c.Name())
	}
	return false
}
