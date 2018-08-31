// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package assign implements the rec.assign command,
// i.e. change taxon assignment of an specimen record.
package assign

import (
	"strings"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/cmdapp"
	"github.com/js-arias/biodv/records"

	"github.com/pkg/errors"
)

var cmd = &cmdapp.Command{
	UsageLine: "rec.assign --to <name> [-c|--check] <record>",
	Short:     "change taxon assignment of an specimen record",
	Long: `
Command rec.assign changes the taxon assignation of an specimen record.
If the -c or --check option is defined, it will check if the taxon
assignation is on a taxon that exist on the taxonomy database.

Options are:

    -to <name>
    --to <name>
      Sets the new assignation of the specimen. It is a required
      parameter.

    -c
    --check
      If set, the taxon name will be validated on the taxonomy
      database.

    <record>
      The record to be re-assigned.
	`,
	Run:           run,
	RegisterFlags: register,
}

func init() {
	cmdapp.Add(cmd)
}

var to string
var check bool

func register(c *cmdapp.Command) {
	c.Flag.StringVar(&to, "to", "", "")
	c.Flag.BoolVar(&check, "check", false, "")
	c.Flag.BoolVar(&check, "c", false, "")
}

func run(c *cmdapp.Command, args []string) error {
	id := strings.Join(args, " ")
	if id == "" {
		return errors.Errorf("%s: a record should be defined", c.Name())
	}

	if check {
		if err := checkTaxon(); err != nil {
			return errors.Wrap(err, c.Name())
		}
	}

	recs, err := records.Open("")
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	if err := recs.Move(id, to); err != nil {
		return errors.Wrap(err, c.Name())
	}
	if err := recs.Commit(); err != nil {
		return errors.Wrap(err, c.Name())
	}
	return nil
}

func checkTaxon() error {
	txm, err := biodv.OpenTax("biodv", "")
	if err != nil {
		return err
	}

	if tax, _ := txm.TaxID(to); tax == nil {
		return errors.Errorf("taxon %q not in taxon database", to)
	}
	return nil
}
