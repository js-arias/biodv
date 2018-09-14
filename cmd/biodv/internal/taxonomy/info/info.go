// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package info implements the tax.info command,
// i.e. print taxon information.
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
	UsageLine: "tax.info [--db <database>] [--id] <taxon>",
	Short:     "print taxon information",
	Long: `
Command tax.info prints the information data available for a taxon in
a given database.

Either a taxon name, or, if the option --id is set, a taxon ID, should
be defined.

Options are:

    -db <database>
    --db <database>
      If set, the indicated database will be used to extract the taxon
      information.
      To see the available databases use the command ‘db.drivers’.
      The default biodv database on the current directory.

    -id
    --id
      If set, the search of the taxon will be based on the taxon ID,
      instead of the taxon name.

    <taxon>
      A required parameter. Indicates the taxon for which the information
      will be printed. If the name is ambiguous, the ID of the ambiguous
      taxa will be printed. If the option --id is set, it must be a taxon
      ID instead of a name.
	`,
	Run:           run,
	RegisterFlags: register,
}

func init() {
	cmdapp.Add(cmd)
}

var dbName string
var id bool

func register(c *cmdapp.Command) {
	c.Flag.StringVar(&dbName, "db", "biodv", "")
	c.Flag.BoolVar(&id, "id", false, "")
}

func run(c *cmdapp.Command, args []string) error {
	if dbName == "" {
		dbName = "biodv"
	}
	var param string
	dbName, param = biodv.ParseDriverString(dbName)

	nm := strings.Join(args, " ")
	if nm == "" {
		return errors.Errorf("%s: a taxon name of ID should be given", c.Name())
	}

	db, err := biodv.OpenTax(dbName, param)
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	tax, err := getTaxon(db, nm)
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	if err := print(db, tax); err != nil {
		return errors.Wrap(err, c.Name())
	}
	return nil
}

// GetTaxon returns a taxon from the options.
func getTaxon(db biodv.Taxonomy, nm string) (biodv.Taxon, error) {
	if id {
		return db.TaxID(nm)
	}
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

func print(db biodv.Taxonomy, tax biodv.Taxon) error {
	var p biodv.Taxon
	if pID := tax.Parent(); pID != "" {
		var err error
		p, err = db.TaxID(pID)
		if err != nil {
			return err
		}
	}
	fmt.Printf("%s %s\n", tax.Name(), tax.Value(biodv.TaxAuthor))
	ids(tax)
	if err := parents(db, tax); err != nil {
		return err
	}
	fmt.Printf("\tRank: %s\n", tax.Rank())
	if tax.IsCorrect() {
		fmt.Printf("\tCorrect-Valid name\n")
		ls, err := biodv.TaxList(db.Synonyms(tax.ID()))
		if err != nil {
			return err
		}
		for _, syn := range ls {
			fmt.Fprintf(os.Stderr, "\t\t%s %s [%s:%s]\n", syn.Name(), syn.Value(biodv.TaxAuthor), dbName, syn.ID())
		}
	} else {
		fmt.Printf("\tSynonym of %s %s [%s:%s]\n", p.Name(), p.Value(biodv.TaxAuthor), dbName, p.ID())
	}
	if p != nil && tax.IsCorrect() {
		fmt.Printf("\tParent: %s %s [%s:%s]\n", p.Name(), p.Value(biodv.TaxAuthor), dbName, p.ID())
	}
	return contained(db, tax)
}

func ids(tax biodv.Taxon) {
	fmt.Printf("\t%s-ID: %s\n", dbName, tax.ID())
	v := tax.Value(biodv.TaxExtern)
	for _, e := range strings.Fields(v) {
		fmt.Printf("\t\t%s\n", e)
	}
}

func parents(db biodv.Taxonomy, tax biodv.Taxon) error {
	pLs, err := biodv.TaxParents(db, tax.Parent())
	if err != nil {
		return err
	}
	for i, pt := range pLs {
		if i > 0 {
			fmt.Printf(" > ")
		}
		if pt.Rank() != biodv.Unranked {
			fmt.Printf("%s: %s", pt.Rank(), pt.Name())
		} else {
			fmt.Printf("%s", pt.Name())
		}
	}
	if len(pLs) > 0 {
		fmt.Printf("\n")
	}
	return nil
}

func contained(db biodv.Taxonomy, tax biodv.Taxon) error {
	if tax.IsCorrect() {
		ls, err := biodv.TaxList(db.Children(tax.ID()))
		if err != nil {
			return err
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
