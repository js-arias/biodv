// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package list implements the tax.list command,
// i.e. prints a list of taxons.
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
	UsageLine: `tax.list [--db <database>] [--id <value>] [-m|--machine]
		[-p|--parents] [-s|--synonym] [-v|--verbose] [<name>]`,
	Short: "prints a list of taxons",
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
      Available databases are:
        gbif

    -id <value>
    --id <value>
      If set, the list will be based on the indicated taxon.

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

    <name>
      If set, the list will be based on the indicated taxon, if the
      name is ambiguous, the ID of the ambigous taxa will be printed.
	`,
	Run:           run,
	RegisterFlags: register,
}

func init() {
	cmdapp.Add(cmd)
}

var dbName string
var id string
var machine bool
var parents bool
var synonyms bool
var verbose bool

func register(c *cmdapp.Command) {
	c.Flag.StringVar(&dbName, "db", "", "")
	c.Flag.StringVar(&id, "id", "", "")
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
		return errors.Errorf("%s: a database must be defined", c.Name())
	}

	if machine && verbose {
		return errors.Errorf("%s: options --machine and --verbose are incompatible", c.Name())
	}
	if parents && synonyms {
		return errors.Errorf("%s: options --parents and --synonyms are incompatible", c.Name())
	}

	db, err := biodv.OpenTax(dbName, "")
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	nm := strings.Join(args, " ")
	if nm != "" && id == "" {
		ls, err := biodv.TaxList(db.Taxon(nm))
		if err != nil {
			return errors.Wrap(err, c.Name())
		}
		if len(ls) == 0 {
			return nil
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
			return nil
		}
		id = ls[0].ID()
	}

	if synonyms {
		if id == "" {
			return errors.Errorf("%s: a taxon must be defined for a synonyms list", c.Name())
		}
		ls, err := biodv.TaxList(db.Synonyms(id))
		if err != nil {
			return errors.Wrap(err, c.Name())
		}
		printList(ls)
		return nil
	}

	if parents {
		if id == "" {
			return errors.Errorf("%s: a taxon must be defined for a parent list", c.Name())
		}
		ls, err := biodv.TaxParents(db, id)
		if err != nil {
			return errors.Wrap(err, c.Name())
		}
		printList(ls)
		return nil
	}

	ls, err := biodv.TaxList(db.Children(id))
	if err != nil {
		return errors.Wrap(err, c.Name())
	}
	printList(ls)
	return nil
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
