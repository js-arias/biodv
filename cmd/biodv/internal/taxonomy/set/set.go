// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package set implements the tax.set command,
// i.e. set a taxon data value.
package set

import (
	"strings"

	"github.com/js-arias/biodv/cmdapp"
	"github.com/js-arias/biodv/taxonomy"

	"github.com/pkg/errors"
)

var cmd = &cmdapp.Command{
	UsageLine: "tax.set -k|--key <key> [-v|--value <value>] <name>",
	Short:     "set a taxon data value",
	Long: `
Command tax.set sets the value of a given key for the indicated taxon,
overwriting any previous value. If value is empty, the content of the
key will be eliminated.

Command tax.set can be used to set almost any key value, except the
taxon name (that can not be changed), the rank (use tax.rank), and the
parent and status (correct/valid or synonym) of the taxon
(use tax.move).

Except for some standard keys, no content of the values will be
validated by the program, or the database.

Options are:

    -k <key>
    --key <key>
      A key, a required parameter. Keys must be in lower case and
      without spaces (it will be reformatted to lower case, and spaces
      between words replaced by the dash ‘-’ character). Any key can
      be stored, but the recognized keys (and their expected values)
      are:
        author     to set the taxon’s author.
        extern     to set the ID on an external database, it will be on
                   the form "<service>:<id>", if only “<service>:” is
                   given, the indicated external ID will be eliminated.
        reference  to set a bibliographic reference.
        source     to set the ID of the source of the taxonomic data.
      For a set of available keys of a given taxon, use tax.value.

    -v <value>
    --value <value>
      The value to set. If no value is defined, or an empty string is
      given, the value on that key will be deleted.

    <name>
      The taxon to be set.
	`,
	Run:           run,
	RegisterFlags: register,
}

func init() {
	cmdapp.Add(cmd)
}

var key string
var value string

func register(c *cmdapp.Command) {
	c.Flag.StringVar(&key, "key", "", "")
	c.Flag.StringVar(&key, "k", "", "")
	c.Flag.StringVar(&value, "value", "", "")
	c.Flag.StringVar(&value, "v", "", "")
}

func run(c *cmdapp.Command, args []string) error {
	if key == "" {
		return errors.Errorf("%s: a key should be defined", c.Name())
	}
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
	if err := tax.Set(key, value); err != nil {
		return errors.Wrap(err, c.Name())
	}
	if err := db.Commit(); err != nil {
		return errors.Wrap(err, c.Name())
	}
	return nil
}
