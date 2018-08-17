// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package drivers implements the db.drivers command,
// i.e. list the database drivers.
package drivers

import (
	"fmt"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/cmdapp"
)

var cmd = &cmdapp.Command{
	UsageLine: "db.drivers [-d|--database <database>]",
	Short:     "list the database drivers",
	Long: `
Command db.drivers prints a list of available drivers, sorted by the
kind of the database used. If the -d or --atabase option is given, only
the drivers for that database will be printed.

Options are:

    -d <database>
    --database <database>
      If set, only the drivers of the given database kind will be
      printed.
      Valid database kinds are:
        taxonomy  taxonomic names databases
	`,
	Run:           run,
	RegisterFlags: register,
}

func init() {
	cmdapp.Add(cmd)
}

var dbKind string

func register(c *cmdapp.Command) {
	c.Flag.StringVar(&dbKind, "database", "", "")
	c.Flag.StringVar(&dbKind, "d", "", "")
}

func run(c *cmdapp.Command, args []string) error {
	ls := biodv.TaxDrivers()
	fmt.Printf("Taxonomy drivers:\n")
	for _, dv := range ls {
		fmt.Printf("    %-16s %s\n", dv, biodv.TaxAbout(dv))
	}
	return nil
}
