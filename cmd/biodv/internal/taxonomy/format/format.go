// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package format implements the tax.format command,
// i.e. synonymize rankless taxa.
package format

import (
	"fmt"
	"os"
	"strings"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/cmdapp"
	"github.com/js-arias/biodv/taxonomy"

	"github.com/pkg/errors"
)

var cmd = &cmdapp.Command{
	UsageLine: "tax.format [<name>]",
	Short:     "synonymize rankless taxa",
	Long: `
Command tax.format search for all unranked taxons (except from taxons
attached to the root), and make them synonyms, so the taxonomy will only
have as correct names, names with some rank. This is useful to, for
example, collapse subspecies as synonyms of a species.

Options are:

    <name>
      If defined, only the descendants of the indicated taxon will be
      searched and synonymized.
	`,
	Run: run,
}

func init() {
	cmdapp.Add(cmd)
}

func run(c *cmdapp.Command, args []string) error {
	db, err := taxonomy.Open("")
	if err != nil {
		return errors.Wrap(err, c.Name())
	}
	nm := strings.Join(args, " ")
	if nm == "" {
		ls := db.TaxList("")
		for _, c := range ls {
			collapse(db, c)
		}
	} else {
		tax := db.TaxEd(nm)
		if tax == nil {
			return err
		}
		collapse(db, tax)
	}

	if err := db.Commit(); err != nil {
		return errors.Wrap(err, c.Name())
	}
	return nil
}

func collapse(db *taxonomy.DB, tax *taxonomy.Taxon) bool {
	for {
		ok := false
		for _, c := range db.TaxList(tax.ID()) {
			if !c.IsCorrect() {
				continue
			}
			if collapse(db, c) {
				ok = true
			}
		}
		if !ok {
			break
		}
	}

	if tax.Rank() != biodv.Unranked {
		return false
	}
	if tax.Parent() == "" {
		return false
	}
	if err := tax.Move(tax.Parent(), false); err != nil {
		fmt.Fprintf(os.Stderr, "warning: when moving %q to %q: %v\n", tax.Name(), tax.Parent(), err)
		return false
	}
	return true
}
