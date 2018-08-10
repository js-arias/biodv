// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package rank implement the tax.rank command,
// i.e. change a taxon rank.
package rank

import (
	"strings"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/cmdapp"
	"github.com/js-arias/biodv/taxonomy"

	"github.com/pkg/errors"
)

var cmd = &cmdapp.Command{
	UsageLine: "tax.rank [-r|--rank <rank>] <name>",
	Short:     "change a taxon rank",
	Long: `
Command tax.rank sets a new rank to a given taxon. If no rank is
defined, it will set the taxon as unranked. The new rank should be
compatible with the current taxonomy.

Options are:

    -r <rank>
    --rank <rank>
      Sets the new rank of the taxon.
      Valid ranks are:
        unranked  (default)
        kingdom
        class
        order
        family
        genus
        species

    <name>
      The taxon to be reranked. This parameter is required.
	`,
	Run:           run,
	RegisterFlags: register,
}

func init() {
	cmdapp.Add(cmd)
}

var rankStr string

func register(c *cmdapp.Command) {
	c.Flag.StringVar(&rankStr, "rank", "unranked", "")
	c.Flag.StringVar(&rankStr, "r", "unranked", "")
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

	r := biodv.GetRank(rankStr)
	if r.String() != strings.ToLower(rankStr) {
		return errors.Errorf("%s: unknown rank %s", c.Name(), rankStr)
	}
	if err := tax.SetRank(r); err != nil {
		return errors.Wrap(err, c.Name())
	}

	if err := db.Commit(); err != nil {
		return errors.Wrap(err, c.Name())
	}
	return nil
}
