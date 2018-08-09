// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package add implements the tax.add command,
// i.e. add taxon names.
package add

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/cmdapp"
	"github.com/js-arias/biodv/taxonomy"

	"github.com/pkg/errors"
)

var cmd = &cmdapp.Command{
	UsageLine: `tax.add [-p|--parent <name>] [-r|--rank <rank>]
		[-s|--synonym] [-v|--verbose] [<file>...]`,
	Short: "add taxon names",
	Long: `
Command tax.add adds one or more taxons from the indicated files, or the
standard input (if no file is defined) to the taxonomy database. It
assumes that each line contains a taxon name (empty lines, or lines
starting with the sharp symbol ( # ) or semicolon character ( ; ) will
be ignored).

By default, taxons will be added to the root of the taxonomy (without
parent), as correct names, and rankless.

Options are:

    -p <name>
    --parent <name>
      Sets the parents of the added taxons. This must be a correct name
      (i.e. not a synonym), and should be present in the database.

    -r <rank>
    --rank <rank>
      Sets the rank of the added taxons. If a parent is defined, the rank
      should be concordant with the parent's rank.
      Valid ranks are:
        unranked
        kingdom
        class
        order
        family
        genus
        species
      If the rank is set to species, it will automatically interpret the
      first part of the name as a genus, and adds the species as a child
      of that genus.

    -s
    --synonym
      If set, the added taxons will be set as synonyms (non correct names)
      of its parent. It requires that a parent will be defined.

    -v
    --verbose
      If set, the name of each added taxon will be printed in the standard
      output.

    <file>
      One or more files to be processed by tax.add. If no file is given,
      the taxon names will be read from the standard input.
	`,
	Run:           run,
	RegisterFlags: register,
}

func init() {
	cmdapp.Add(cmd)
}

var parent string
var rank string
var synonym bool
var verbose bool

func register(c *cmdapp.Command) {
	c.Flag.StringVar(&parent, "parent", "", "")
	c.Flag.StringVar(&parent, "p", "", "")
	c.Flag.StringVar(&rank, "rank", "", "")
	c.Flag.StringVar(&rank, "r", "", "")
	c.Flag.BoolVar(&synonym, "synonym", false, "")
	c.Flag.BoolVar(&synonym, "s", false, "")
	c.Flag.BoolVar(&verbose, "verbose", false, "")
	c.Flag.BoolVar(&verbose, "v", false, "")
}

func run(c *cmdapp.Command, args []string) error {
	db, err := taxonomy.Open("")
	if err != nil {
		return errors.Wrap(err, c.Name())
	}
	r := biodv.GetRank(rank)
	var p biodv.Taxon
	if parent != "" {
		p, _ = db.TaxID(parent)
		if p == nil {
			return errors.Errorf("%s: parent %q not in database", c.Name(), parent)
		}
		if !p.IsCorrect() {
			return errors.Errorf("%s: taxon %q can not be a parent", c.Name(), p.Name())
		}
		if !isRankValid(db, p, r) {
			return errors.Errorf("%s: rank %s incompatible with taxonomy", c.Name(), r)
		}
		parent = p.Name()
	}
	if synonym && p == nil {
		return errors.Errorf("%s: synonym taxons require a parent")
	}
	if len(args) == 0 {
		args = append(args, "-")
	}
	for _, a := range args {
		if a == "-" {
			if err := read(db, os.Stdin, r); err != nil {
				return errors.Wrapf(err, "%s: while reading from stdin", c.Name())
			}
			continue
		}
		f, err := os.Open(a)
		if err != nil {
			return errors.Wrapf(err, "%s: unable to open %s", c.Name(), a)
		}
		err = read(db, f, r)
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
func read(db *taxonomy.DB, r io.Reader, rk biodv.Rank) error {
	s := bufio.NewScanner(r)
	for s.Scan() {
		name := strings.Join(strings.Fields(s.Text()), " ")
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

		pname := parent

		// we known that the first part of a sepcies name is the genus
		if rk == biodv.Species && !synonym {
			g := strings.Fields(name)[0]
			if p, _ := db.TaxID(g); p != nil {
				pname = p.Name()
			} else {
				p, err := db.Add(g, pname, biodv.Genus, true)
				if err != nil {
					return err
				}
				pname = p.Name()
				if verbose {
					fmt.Printf("%s\n", p.Name())
				}
			}
		}

		tax, err := db.Add(name, pname, rk, !synonym)
		if err != nil {
			return err
		}
		if verbose {
			fmt.Printf("%s\n", tax.Name())
		}
	}
	return s.Err()
}

// IsRankValid returns true if the rank r
// is compatible with the taxonomy.
func isRankValid(db *taxonomy.DB, p biodv.Taxon, rk biodv.Rank) bool {
	if rk == biodv.Unranked {
		return true
	}
	for ; p != nil; p, _ = db.TaxID(p.Parent()) {
		r := p.Rank()
		if r == biodv.Unranked {
			continue
		}
		if rk > r {
			return true
		}
		if rk == r && synonym {
			return true
		}
		return false
	}
	return true
}
