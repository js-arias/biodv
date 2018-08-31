// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package del implements the rec.del command,
// i.e. eliminate an specimen record from the database.
package del

import (
	"strings"

	"github.com/js-arias/biodv/cmdapp"
	"github.com/js-arias/biodv/records"

	"github.com/pkg/errors"
)

var cmd = &cmdapp.Command{
	UsageLine: "rec.del <record>",
	Short:     "eliminate an specimen record from the database",
	Long: `
Command rec.del removes the indicated specimen record from the database.

Options are:

    <name>
      The specimen record to be deleted.
	`,
	Run: run,
}

func init() {
	cmdapp.Add(cmd)
}

func run(c *cmdapp.Command, args []string) error {
	id := strings.Join(args, " ")
	if id == "" {
		return errors.Errorf("%s: a record should be defined", c.Name())
	}

	recs, err := records.Open("")
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	recs.Delete(id)
	if err := recs.Commit(); err != nil {
		return errors.Wrap(err, c.Name())
	}
	return nil
}
