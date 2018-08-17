// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package dbadd implements the tax.db.add command,
// i.e. add taxons validated on an external DB.
package dbadd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"unicode"
	"unicode/utf8"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/cmdapp"
	"github.com/js-arias/biodv/taxonomy"

	"github.com/pkg/errors"
)

var cmd = &cmdapp.Command{
	UsageLine: `tax.db.add -e|--extern <database> [-u|--uprank <rank>]
		[<file>...]`,
	Short: "add taxons validated on an external DB",
	Long: `
Command tax.db.add adds one or more taxons from the indicated files,
or the standard input (if no file is defined) to the taxonomy database.
Each name is validated against an external database, and only taxons
found in that DB will be added (with the additional information given
by the database.

If a taxon name can not be added, it will be printed in the standard
output.

In the input file, it is assumed that each line contains a taxon name
(empty lines, or lines starting with the sharp symbol ( # ) or
semicolon character ( ; ) will be ignored).

If the option -u or --uprank is given, it will add additional parents up
to the given rank.

Options are:

    -e <database>
    --extern <database>
      A required parameter. It will set the external database.
      To see the available databases use the command ‘db.drivers’.

    -u <rank>
    --uprank <rank>
      If set, parent taxons, up to the given rank, will be added to
      the database.
      Valid ranks are:
        kingdom
        class
        order
        family
        genus     [default value]
        species
	`,
	Run:           run,
	RegisterFlags: register,
}

func init() {
	cmdapp.Add(cmd)
}

var extName string
var upRank string

func register(c *cmdapp.Command) {
	c.Flag.StringVar(&extName, "extern", "", "")
	c.Flag.StringVar(&extName, "e", "", "")
	c.Flag.StringVar(&upRank, "uprank", biodv.Genus.String(), "")
	c.Flag.StringVar(&upRank, "u", biodv.Genus.String(), "")
}

func run(c *cmdapp.Command, args []string) (err error) {
	if extName == "" {
		return errors.Errorf("%s: an external database should be defined", c.Name())
	}
	var param string
	extName, param = biodv.ParseDriverString(extName)
	ext, err := biodv.OpenTax(extName, param)
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	var rank biodv.Rank
	if upRank != "" {
		rank = biodv.GetRank(upRank)
		if rank == biodv.Unranked {
			return errors.Errorf("%s: unknown rank: %s", c.Name(), upRank)
		}
	}
	db, err := taxonomy.Open("")
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	if len(args) == 0 {
		args = append(args, "-")
	}
	for _, a := range args {
		if a == "-" {
			if err := read(db, ext, os.Stdin, rank); err != nil {
				return errors.Wrapf(err, "%s: while reading from stdin", c.Name())
			}
			continue
		}
		f, err := os.Open(a)
		if err != nil {
			return errors.Wrapf(err, "%s: unable to open %s", c.Name(), a)
		}
		err = read(db, ext, f, rank)
		f.Close()
		if err != nil {
			return errors.Wrapf(err, "%s: while reading from %s", c.Name(), a)
		}
	}
	if err := db.Commit(); err != nil {
		return errors.Wrap(err, c.Name())
	}
	return nil
}

// Read reads the data from a reader.
func read(db *taxonomy.DB, ext biodv.Taxonomy, r io.Reader, rk biodv.Rank) error {
	s := bufio.NewScanner(r)
	for s.Scan() {
		name := biodv.TaxCanon(s.Text())
		if name == "" {
			continue
		}
		if nm, _ := utf8.DecodeRuneInString(name); nm == '#' || nm == ';' || !unicode.IsLetter(nm) {
			continue
		}

		// skip taxons already in database
		if tax, _ := db.TaxID(name); tax != nil {
			continue
		}
		ls, err := biodv.TaxList(ext.Taxon(name))
		if err != nil {
			fmt.Printf("%s\n", name)
			fmt.Fprintf(os.Stderr, "warning: when searching %s: %v\n", name, err)
			continue
		}
		tx := matchFromParent(db, ls)
		if tx == nil {
			fmt.Printf("%s\n", name)
			if len(ls) == 0 {
				fmt.Fprintf(os.Stderr, "warning: when searching %s: not in database\n", name)
				continue
			}
			fmt.Fprintf(os.Stderr, "warning: ambiguous name:\n")
			for _, tx := range ls {
				fmt.Fprintf(os.Stderr, "\t%s:%s\t%s %s\t", extName, tx.ID(), tx.Name(), tx.Value(biodv.TaxAuthor))
				if tx.IsCorrect() {
					fmt.Fprintf(os.Stderr, "correct name\n")
				} else {
					fmt.Fprintf(os.Stderr, "synonym\n")
				}
			}
			continue
		}
		tax := addTaxon(db, ext, tx, rk)
		if tax == nil {
			fmt.Printf("%s\n", name)
			continue
		}
		if tax.Name() != name {
			fmt.Fprintf(os.Stderr, "warning: taxon %q added as %q [%s:%s]\n", name, tax.Name(), extName, tx.ID())
		}
	}
	return s.Err()
}

func addTaxon(db *taxonomy.DB, ext biodv.Taxonomy, tx biodv.Taxon, r biodv.Rank) *taxonomy.Taxon {
	var pID string
	if r != biodv.Unranked && (tx.Rank() > r || tx.Rank() == biodv.Unranked) {
		p := db.TaxEd(extName + ":" + tx.Parent())
		if p == nil {
			ptx, err := ext.TaxID(tx.Parent())
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: %v\n", err)
				return nil
			}
			p = addTaxon(db, ext, ptx, r)
			if p == nil {
				return nil
			}
		}
		pID = p.ID()
	} else {
		p := db.TaxEd(extName + ":" + tx.Parent())
		if p != nil {
			pID = p.ID()
		}
	}

	tax, err := db.Add(tx.Name(), pID, tx.Rank(), tx.IsCorrect())
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: %v\n", err)
		return nil
	}
	if err := tax.Set(biodv.TaxExtern, extName+":"+tx.ID()); err != nil {
		fmt.Fprintf(os.Stderr, "warning: when matching %s: %v\n", tax.Name(), err)
		return nil
	}
	update(tax, tx)
	return tax
}

func update(tax *taxonomy.Taxon, tx biodv.Taxon) {
	for _, k := range tx.Keys() {
		v := tx.Value(k)
		if k == biodv.TaxSource {
			v = extName + ":" + v
		}
		if err := tax.Set(k, v); err != nil {
			fmt.Fprintf(os.Stderr, "warning: when updating %s: %v\n", tax.Name(), err)
		}
	}
}

// MatchFromParent search for an extern taxon
// that has a parent already in the database.
func matchFromParent(db *taxonomy.DB, ls []biodv.Taxon) biodv.Taxon {
	if len(ls) == 0 {
		return nil
	}
	if len(ls) == 1 {
		return ls[0]
	}
	for _, tx := range ls {
		if tax := db.TaxEd(extName + ":" + tx.Parent()); tax != nil {
			return tx
		}
	}
	return nil
}
