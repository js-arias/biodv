// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package list implements the tax.list command,
// i.e. print a list of taxons.
package list

import (
	"fmt"
	"os"
	"strings"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/cmdapp"

	"github.com/pkg/errors"
)

var cmd = &cmdapp.Command{
	UsageLine: `tax.list [--db <database>] [--id] [-m|--machine]
		[-p|--parents] [-s|--synonym] [-v|--verbose] [<taxon>]`,
	Short: "print a list of taxons",
	Long: `
Command tax.list prints a list of the contained taxa of a given taxon
in a given database.

If no taxon is defined, the list of taxons attached to the root of
the database will be given.

If the option synonym is defined, instead of the contained taxa, a
list of synonyms of the name will be given.

Only names will be printed, if the option machine is defined, only IDs
will be printed, and verbose option will print ID, taxon name and
taxon author (if available).

Options are:

    -db <database>
    --db <database>
      If set, the indicated database will be used to produce the taxon
      list.
      To see the available databases use the command ‘db.drivers’.
      The default biodv database on the current directory.

    -id
    --id
      If set, the search of the taxon will be based on the taxon ID,
      instead of the taxon name.

    -m
    --machine
      If set, only the IDs of the taxons will be printed.

    -p
    --parents
      If set, a list of parents of the taxon will be produced.

    -s
    --synonyms
      If set, a list of synonyms of the taxon, instead of contained
      taxa, will be produced.

    -v
    --verbose
      If set, the list will produced indicating the ID, the taxon
      name, and the author of the taxon.

    <taxon>
      A required parameter. Indicates the taxon for which the list will
      be printed. If the name is ambiguous, the ID of the ambiguous taxa
      will be printed. If the option --id is set, it must be a taxon ID
      instead of a taxon name.
	`,
	Run:           run,
	RegisterFlags: register,
}

func init() {
	cmdapp.Add(cmd)
}

var dbName string
var id bool
var machine bool
var parents bool
var synonyms bool
var verbose bool

func register(c *cmdapp.Command) {
	c.Flag.StringVar(&dbName, "db", "", "")
	c.Flag.BoolVar(&id, "id", false, "")
	c.Flag.BoolVar(&machine, "machine", false, "")
	c.Flag.BoolVar(&machine, "m", false, "")
	c.Flag.BoolVar(&parents, "parents", false, "")
	c.Flag.BoolVar(&parents, "p", false, "")
	c.Flag.BoolVar(&synonyms, "synonyms", false, "")
	c.Flag.BoolVar(&synonyms, "s", false, "")
	c.Flag.BoolVar(&verbose, "verbose", false, "")
	c.Flag.BoolVar(&verbose, "v", false, "")
}

func run(c *cmdapp.Command, args []string) error {
	if dbName == "" {
		dbName = "biodv"
	}
	var param string
	dbName, param = biodv.ParseDriverString(dbName)

	if machine && verbose {
		return errors.Errorf("%s: options --machine and --verbose are incompatible", c.Name())
	}
	if parents && synonyms {
		return errors.Errorf("%s: options --parents and --synonyms are incompatible", c.Name())
	}

	db, err := biodv.OpenTax(dbName, param)
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	idVal := strings.Join(args, "")
	if !id && len(args) >= 1 {
		nm := strings.Join(args, " ")
		tax, err := getTaxon(db, nm)
		if err != nil {
			return errors.Wrap(err, c.Name())
		}
		if tax == nil {
			return nil
		}
		idVal = tax.ID()
	}

	ls, err := getList(db, idVal)
	if err != nil {
		return errors.Wrap(err, c.Name())
	}
	printList(ls)

	return nil
}

// GetTaxon returns a taxon from the options.
func getTaxon(db biodv.Taxonomy, nm string) (biodv.Taxon, error) {
	ls, err := biodv.TaxList(db.Taxon(nm))
	if err != nil {
		return nil, err
	}
	if len(ls) == 0 {
		return nil, nil
	}
	if len(ls) > 1 {
		fmt.Fprintf(os.Stderr, "ambiguous name:\n")
		for _, tx := range ls {
			fmt.Fprintf(os.Stderr, "id:%s\t%s %s\t", tx.ID(), tx.Name(), tx.Value(biodv.TaxAuthor))
			if tx.IsCorrect() {
				fmt.Fprintf(os.Stderr, "correct name\n")
			} else {
				fmt.Fprintf(os.Stderr, "synonym\n")
			}
		}
		return nil, nil
	}
	return ls[0], nil
}

func getList(db biodv.Taxonomy, idVal string) ([]biodv.Taxon, error) {
	if synonyms {
		if idVal == "" {
			return nil, errors.New("a taxon must be defined for a synonyms list")
		}
		ls, err := biodv.TaxList(db.Synonyms(idVal))
		if err != nil {
			return nil, err
		}
		return ls, nil
	}

	if parents {
		if idVal == "" {
			return nil, errors.New("a taxon must be defined for a parent list")
		}
		ls, err := biodv.TaxParents(db, idVal)
		if err != nil {
			return nil, err
		}
		return ls, nil
	}

	return biodv.TaxList(db.Children(idVal))
}

func printList(ls []biodv.Taxon) {
	for _, tax := range ls {
		if machine {
			fmt.Printf("%s\n", tax.ID())
		} else if verbose {
			fmt.Printf("%s\t%s %s\n", tax.ID(), tax.Name(), tax.Value(biodv.TaxAuthor))
		} else {
			fmt.Printf("%s\n", tax.Name())
		}
	}
}
