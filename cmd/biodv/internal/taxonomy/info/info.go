// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package info implements the tax.info command,
// i.e. prints taxon information.
package info

import (
	"fmt"
	"os"
	"strings"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/cmdapp"

	"github.com/pkg/errors"
)

var cmd = &cmdapp.Command{
	UsageLine: "tax.info [--db <database>] [--id <value>] [<name>]",
	Short:     "prints taxon information",
	Long: `
Command tax.info prints the information data available for a taxon name, in
a given database.

Either a taxon name, of a database id, should be used.

Options are:

    -db <database>
    --db <database>
      If set, the indicated database will be used to extract the taxon
      information.
      Available databases are:
        gbif

    -id <value>
    --id <value>
      If set, the information of the indicated taxon will be printed.

    <name>
      If set, the information taxon with the given name will be printed,
      if the name is ambiguous, the ID of the ambigous taxa will be
      printed.
	`,
	Run:           run,
	RegisterFlags: register,
}

func init() {
	cmdapp.Add(cmd)
}

var dbName string
var id string

func register(c *cmdapp.Command) {
	c.Flag.StringVar(&dbName, "db", "", "")
	c.Flag.StringVar(&id, "id", "", "")
}

func run(c *cmdapp.Command, args []string) error {
	if dbName == "" {
		return errors.Errorf("%s: a database must be defined", c.Name())
	}
	nm := strings.Join(args, " ")
	if id == "" && nm == "" {
		return errors.Errorf("%s: either a --id or a taxon name, should be given", c.Name())
	}

	db, err := biodv.OpenTax(dbName, "")
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	var tax biodv.Taxon
	if id != "" {
		tax, err = db.TaxID(id)
		if err != nil {
			return errors.Wrap(err, c.Name())
		}
	} else {
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
		tax = ls[0]
	}
	var p biodv.Taxon
	if pID := tax.Parent(); pID != "" {
		p, err = db.TaxID(pID)
		if err != nil {
			return errors.Wrap(err, c.Name())
		}
	}
	fmt.Printf("%s %s\n", tax.Name(), tax.Value(biodv.TaxAuthor))
	fmt.Printf("%s-ID: %s\n", dbName, tax.ID())
	pLs, err := biodv.TaxParents(db, tax.Parent())
	if err != nil {
		return errors.Wrap(err, c.Name())
	}
	for i, pt := range pLs {
		if i > 0 {
			fmt.Printf(" > ")
		}
		if pt.Rank() != biodv.Unranked {
			fmt.Printf("%s: %s", pt.Rank(), pt.Name())
		}
	}
	if len(pLs) > 0 {
		fmt.Printf("\n")
	}
	fmt.Printf("\tRank: %s\n", tax.Rank())
	if tax.IsCorrect() {
		fmt.Printf("\tCorrect-Valid name\n")
		ls, err := biodv.TaxList(db.Synonyms(tax.ID()))
		if err != nil {
			return errors.Wrap(err, c.Name())
		}
		for _, syn := range ls {
			fmt.Fprintf(os.Stderr, "\t\t%s %s [%s:%s]\n", syn.Name(), syn.Value(biodv.TaxAuthor), dbName, syn.ID())
		}
	} else {
		fmt.Printf("\tSynonym of %s %s [%s:%s]\n", p.Name(), p.Value(biodv.TaxAuthor), dbName, p.ID())
	}
	if p != nil && tax.IsCorrect() {
		fmt.Printf("\tParent: %s %s [%s:%s]\n", p.Name(), p.Value(biodv.TaxAuthor), dbName, p.ID())

		ls, err := biodv.TaxList(db.Children(tax.ID()))
		if err != nil {
			return errors.Wrap(err, c.Name())
		}
		if len(ls) > 0 {
			fmt.Fprintf(os.Stderr, "\tContained taxa (%d):\n", len(ls))
		}
		for _, child := range ls {
			fmt.Fprintf(os.Stderr, "\t\t%s %s [%s:%s]\n", child.Name(), child.Value(biodv.TaxAuthor), dbName, child.ID())
		}
	}
	return nil
}
