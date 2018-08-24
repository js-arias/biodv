// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package move implements the tax.move command,
// i.e. change a taxon parent.
package move

import (
	"strings"

	"github.com/js-arias/biodv/cmdapp"
	"github.com/js-arias/biodv/taxonomy"

	"github.com/pkg/errors"
)

var cmd = &cmdapp.Command{
	UsageLine: "tax.move [--to <name>] [-s|--status <value>] <name>",
	Short:     "change a taxon parent",
	Long: `
Command tax.move changes moves the taxon to a new parent. If the -s,
or --status option is not defined, it will move the taxon with their
current status. The movement must be compatible with the current
taxonomy, or an error will be produced.

If a taxon is set as a synonym (either with the –status option or
because is already a synonym), it should have a defined parent (with
the --to option), as a synonym can not be attached to the root of
the taxonomy. If the moved taxon have descendants, all of the
descendants will be moved as children of the new parent.

Only the correct/valid taxons can be attached to the root.

Only a correct/valid taxon can be a parent.

Options are:

    -to <name>
    --to <name>
      Sets the new parent of the taxon. If no parent is set, the
      taxons will be moved to the root of the taxonomy.

    -s <value>
    --status <value>
      Sets the new status of the taxon. If no status is given, the
      current taxon status will be used.
      Valid values are:
        correct   the taxon name is correct
        synonym   the taxon name is a synonym
        accepted  equivalent to correct
        valid     equivalent to correct
        true      equivalent to correct
        false     equivalent to synonym

    <name>
      The taxon to be moved. This parameter is required.
	`,
	Run:           run,
	RegisterFlags: register,
}

func init() {
	cmdapp.Add(cmd)
}

var to string
var status string

func register(c *cmdapp.Command) {
	c.Flag.StringVar(&to, "to", "", "")
	c.Flag.StringVar(&status, "status", "", "")
	c.Flag.StringVar(&status, "s", "", "")
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
	sv := tax.IsCorrect()

	switch strings.ToLower(status) {
	case "":
	case "correct":
		sv = true
	case "accepted":
		sv = true
	case "valid":
		sv = true
	case "true":
		sv = true
	case "synonym":
		sv = false
	case "false":
		sv = false
	default:
		return errors.Errorf("%s: invalid status value %q", c.Name(), status)
	}

	pID, err := getParentID(db)
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	if err := tax.Move(pID, sv); err != nil {
		return errors.Wrap(err, c.Name())
	}

	if err := db.Commit(); err != nil {
		return errors.Wrap(err, c.Name())
	}
	return nil
}

func getParentID(db *taxonomy.DB) (string, error) {
	var pID string

	if to != "" {
		p := db.TaxEd(to)
		if p == nil {
			return "", errors.Errorf("parent %q not in database", to)
		}
		pID = p.ID()
	}
	return pID, nil
}
