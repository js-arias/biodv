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

    c [<taxon>]
    count [<taxon>]
      Print the number of specimen records of a given taxon.

    h [<command>]
    help [<command>]
      Print command help.

    l [<taxon>]
    list [<taxon>]
      List descendants of a taxon.

    r [<record>]
    record [<record>]
      Move to the indicated record.

    q
    quit
      Quit the program, without making any change.

    t <taxon>
    taxon <taxon>
      Move to the indicated taxon.
	`,
	Run: run,
}

func init() {
	cmdapp.Add(cmd)
}

var txm biodv.Taxonomy
var tax biodv.Taxon
var recs *records.DB
var recLs []*records.Record
var curRec int

func run(c *cmdapp.Command, args []string) error {
	var err error
	txm, err = biodv.OpenTax("biodv", "")
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	recs, err = records.Open("")
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	i := cmdapp.NewInter(os.Stdin)
	addCommands(i)
	i.Prompt = prompt()

	i.Loop()
	return nil
}

func prompt() string {
	var p = "root:"
	if tax != nil {
		p = fmt.Sprintf("%s:", tax.Name())
	}
	if len(recLs) > 0 {
		p += fmt.Sprintf("%s:", recLs[curRec].ID())
	}
	return p
}

func addCommands(i *cmdapp.Inter) {
	i.Add(&cmdapp.Cmd{"c", "count", "number of specimen records", countHelp, countCmd})
	i.Add(&cmdapp.Cmd{"l", "list", "list descendant taxons", listHelp, listCmd})
	i.Add(&cmdapp.Cmd{"q", "quit", "quit the program", quitHelp, func([]string) bool { return true }})
	i.Add(&cmdapp.Cmd{"r", "record", "move to specimen record", recordHelp, recordCmd(i)})
	i.Add(&cmdapp.Cmd{"t", "taxon", "move to taxon", taxonHelp, taxonCmd(i)})
}

var countHelp = `
Usage:
    c [<taxon>]
    count [<taxon>]
Indicates the number of specimen records attached to the indiated
taxon (not including descendants). If no taxon is given, it will use
the current taxon.
`

func countCmd(args []string) bool {
	nm := strings.Join(args, " ")
	switch nm {
	case "", ".":
		if tax == nil {
			return false
		}
		nm = tax.ID()
	case "/":
		return false
	case "..":
		if tax == nil {
			return false
		}
		nm = tax.Parent()
		if nm == "" {
			return false
		}
	default:
		nt, _ := txm.TaxID(nm)
		if nt == nil {
			return false
		}
		nm = nt.ID()
	}

	ls := recs.RecList(nm)
	fmt.Printf("%d\n", len(ls))
	return false
}

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
	switch nm {
	case "", ".":
		if tax != nil {
			nm = tax.ID()
		}
	case "/":
		nm = ""
	case "..":
		if tax == nil {
			return false
		}
		nm = tax.Parent()
	default:
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

var quitHelp = `
Usage:
    q
    quit
Ends the program without saving any change.
`

var recordHelp = `
Usage:
    r [<record>]
    record [<record>]
Change the current record to the indicated record. If no record ID
is given, it will use the first record of the current
taxon.
`

func recordCmd(i *cmdapp.Inter) func(args []string) bool {
	return func(args []string) bool {
		id := strings.Join(args, " ")
		if id == "" {
			recLs = recs.RecList(tax.ID())
			curRec = 0
			if len(recLs) == 0 {
				return false
			}
		} else {
			rec := recs.Record(id)
			if rec == nil {
				return false
			}
			if rec.Taxon() != tax.ID() {
				nt, _ := txm.TaxID(rec.Taxon())
				if nt == nil {
					return false
				}
				tax = nt
			}
			recLs = recs.RecList(tax.ID())
			for i, r := range recLs {
				if r.ID() == rec.ID() {
					curRec = i
					break
				}
			}
		}
		i.Prompt = prompt()
		return false
	}
}

var taxonHelp = `
Usage:
    t <taxon>
    taxon <taxon>
Changes the current taxon to the indicated taxon. To move to a
parent use '..' to move to a parent, or use '/' to move to the
root of the taxonomy.
`

func taxonCmd(i *cmdapp.Inter) func(args []string) bool {
	return func(args []string) bool {
		nm := strings.Join(args, " ")
		switch nm {
		case "", ".":
			return false
		case "/":
			tax = nil
		case "..":
			if tax == nil {
				return false
			}
			tax, _ = txm.TaxID(tax.Parent())
		default:
			nt, _ := txm.TaxID(nm)
			if nt == nil {
				return false
			}
			tax = nt
		}
		recLs = nil
		curRec = 0
		i.Prompt = prompt()
		return false
	}
}
