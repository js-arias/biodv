// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package validate implements the tax.validate command,
// i.e. validate a taxonomy database.
package validate

import (
	"github.com/js-arias/biodv/cmdapp"
	"github.com/js-arias/biodv/taxonomy"

	"github.com/pkg/errors"
)

var cmd = &cmdapp.Command{
	UsageLine: "tax.validate",
	Short:     "validate a taxonomy database",
	Long: `
Command tax.validate validates a taxonomy database. It is useful to test
if a biodov database from a third party is correct.  If there are no
errors, it will finish silently.
	`,
	Run: run,
}

func init() {
	cmdapp.Add(cmd)
}

func run(c *cmdapp.Command, args []string) error {
	if _, err := taxonomy.Open(""); err != nil {
		return errors.Wrap(err, c.Name())
	}
	return nil
}
