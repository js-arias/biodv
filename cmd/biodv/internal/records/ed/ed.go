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

    d [<taxon>]
    desc [<taxon>]
      List descendants of a taxon.

    h [<command>]
    help [<command>]
      Print command help.

    l [<taxon>]
    list [<taxon>]
      List the IDs of the specimen records of a given taxon.

    n
    next
      Move to the next specimen record.

    p
    prev
      Move to the previous specimen record.

    rk [<taxon>]
    rank [<taxon>]
      Print the rank of a taxon.

    r [<record>]
    record [<record>]
      Move to the indicated specimen record.

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
	i.Add(&cmdapp.Cmd{"d", "desc", "list descendant taxons", descHelp, descCmd})
	i.Add(&cmdapp.Cmd{"l", "list", "list specimen records", listHelp, listCmd})
	i.Add(&cmdapp.Cmd{"n", "next", "move to next specimen record", nextHelp, nextCmd(i)})
	i.Add(&cmdapp.Cmd{"p", "prev", "move to previous specimen record", prevHelp, prevCmd(i)})
	i.Add(&cmdapp.Cmd{"q", "quit", "quit the program", quitHelp, func([]string) bool { return true }})
	i.Add(&cmdapp.Cmd{"rk", "rank", "print taxon rank", rankHelp, rankCmd})
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

var descHelp = `
Usage:
    d [<taxon>]
    desc [<taxon>]
Without parameters shows the list of descendants of the current taxon.
If a taxon is given, it will show the descendants of the indicated
taxon.
`

func descCmd(args []string) bool {
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

var listHelp = `
Usage:
    l [<taxon>]
    list [<taxon>]
List the IDs records of the records of the given taxon. If no taxon
is given, it will list the records of the current taxon.
`

func listCmd(args []string) bool {
	var ls []*records.Record
	nm := strings.Join(args, " ")
	switch nm {
	case "", ".":
		if tax == nil {
			return false
		}
		ls = recLs
		if len(recLs) == 0 {
			ls = recs.RecList(tax.ID())
		}
		if len(ls) == 0 {
			return false
		}
	case "/":
		return false
	case "..":
		if tax == nil {
			return false
		}
		p := tax.Parent()
		if p == "" {
			return false
		}
		ls = recs.RecList(p)
	default:
		ls = recs.RecList(nm)
	}
	for _, r := range ls {
		fmt.Printf("%s\n", r.ID())
	}
	return false
}

var quitHelp = `
Usage:
    q
    quit
Ends the program without saving any change.
`

var nextHelp = `
Usage:
    n
    next
Move the record to the next record of the list.
`

func nextCmd(i *cmdapp.Inter) func(args []string) bool {
	return func(args []string) bool {
		if tax == nil {
			return false
		}
		if len(recLs) == 0 {
			ls := recs.RecList(tax.ID())
			if len(ls) == 0 {
				return false
			}
			recLs = ls
			curRec = 0
		} else {
			curRec++
			if curRec >= len(recLs) {
				recLs = nil
				curRec = 0
			}
		}
		i.Prompt = prompt()
		return false
	}
}

var prevHelp = `
Usage:
    p
    prev
Move the record to the previous record of the list.
`

func prevCmd(i *cmdapp.Inter) func(args []string) bool {
	return func(args []string) bool {
		if tax == nil {
			return false
		}
		if len(recLs) == 0 {
			ls := recs.RecList(tax.ID())
			if len(ls) == 0 {
				return false
			}
			recLs = ls
			curRec = len(ls) - 1
		} else {
			curRec--
			if curRec < 0 {
				recLs = nil
				curRec = 0
			}
		}
		i.Prompt = prompt()
		return false
	}
}

var rankHelp = `
Usage:
    rk [<taxon>]
    rank [<taxon>]
Print the rank of a taxon. If no taxon is given it will print the
rank of the current taxon. If the taxon is unranked, the rank of
the most inmediate ranked parent will be printed in parenthesis.
`

func rankCmd(args []string) bool {
	tx := tax
	nm := strings.Join(args, " ")
	switch nm {
	case "", ".":
		if tax == nil {
			return false
		}
	case "/":
		return false
	case "..":
		if tax == nil {
			return false
		}
		tx, _ = txm.TaxID(tax.Parent())
	default:
		tx, _ = txm.TaxID(nm)
	}
	if tx == nil {
		return false
	}
	r := tx.Rank()
	if r == biodv.Unranked {
		r = getRank(tx)
		if r == biodv.Unranked {
			fmt.Printf("%s\n", r)
			return false
		}
		fmt.Printf("%s (%s)\n", biodv.Unranked, r)
		return false
	}
	fmt.Printf("%s\n", r)
	return false
}

func getRank(tx biodv.Taxon) biodv.Rank {
	for tx != nil {
		if tx.Rank() != biodv.Unranked {
			return tx.Rank()
		}
		tx, _ = txm.TaxID(tx.Parent())
	}
	return biodv.Unranked
}

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
