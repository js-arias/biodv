// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package del implements the tax.del command,
// i.e. eliminate a taxon from the database.
package del

import (
	"strings"

	"github.com/js-arias/biodv/cmdapp"
	"github.com/js-arias/biodv/taxonomy"

	"github.com/pkg/errors"
)

var cmd = &cmdapp.Command{
	UsageLine: "tax.del [-r|--recurse] <name>",
	Short:     "eliminate a taxon from the database",
	Long: `
Command tax.del removes a taxon from the database. By default only
the indicated taxon will be deleted, and both children and synonyms
will be moved to the parent of the taxon.

If the option -r or --recurse is used, then the taxon, and all of
its descendants will be deleted.

If the taxon is attached to the root, its synonyms will be also
deleted.

Options are:

    -r
    --recurse
      If set, the indicated taxon, as well of all of its
      descendants, will be eliminated from the database.

    <name>
      The taxon to be deleted.
	`,
	Run:           run,
	RegisterFlags: register,
}

func init() {
	cmdapp.Add(cmd)
}

var recurse bool

func register(c *cmdapp.Command) {
	c.Flag.BoolVar(&recurse, "recurse", false, "")
	c.Flag.BoolVar(&recurse, "r", false, "")
}

func run(c *cmdapp.Command, args []string) error {
	nm := strings.Join(args, " ")
	if nm == "" {
		return errors.Errorf("%s: a taxon name should be given", c.Name())
	}

	db, err := taxonomy.Open("")
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	tax := db.TaxEd(nm)
	if tax == nil {
		return nil
	}
	tax.Delete(recurse)
	if err := db.Commit(); err != nil {
		return errors.Wrap(err, c.Name())
	}
	return nil
}
