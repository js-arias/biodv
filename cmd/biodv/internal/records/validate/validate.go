// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package validate implements the rec.validate command,
// i.e. validate an specimen records database.
package validate

import (
	"github.com/js-arias/biodv/cmdapp"
	"github.com/js-arias/biodv/records"

	"github.com/pkg/errors"
)

var cmd = &cmdapp.Command{
	UsageLine: "rec.validate",
	Short:     "validate an specimen records database",
	Long: `
Command rec.validate validates a records database. It is useful to test
if a biodv database from a third party is correct. If there are no
errors, it will finish silently.
	`,
	Run: run,
}

func init() {
	cmdapp.Add(cmd)
}

func run(c *cmdapp.Command, args []string) error {
	if _, err := records.Open(""); err != nil {
		return errors.Wrap(err, c.Name())
	}
	return nil
}
