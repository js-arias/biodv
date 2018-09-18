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
	"strings"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/cmdapp"
)

var cmd = &cmdapp.Command{
	UsageLine: "db.drivers [<database>]",
	Short:     "list the database drivers",
	Long: `
Command db.drivers prints a list of available drivers, sorted by the
kind of the database used. If a <database> is given, only the drivers
for that database will be printed.

Options are:

    <database>
      If set, only the drivers of the given database kind will be
      printed.
      Valid database kinds are:
      	dataset   dataset databases
        records   specimen record databases
        taxonomy  taxonomic names databases
	`,
	Run: run,
}

func init() {
	cmdapp.Add(cmd)
}

func run(c *cmdapp.Command, args []string) error {
	dbKind := strings.ToLower(strings.Join(args, ""))
	switch dbKind {
	case "records":
		recDrivers()
	case "taxonomy":
		taxDrivers()
	case "dataset":
		setDrivers()
	default:
		setDrivers()
		recDrivers()
		taxDrivers()
	}
	return nil
}

func setDrivers() {
	ls := biodv.SetDrivers()
	fmt.Printf("Dataset-DB drivers:\n")
	for _, dv := range ls {
		fmt.Printf("    %-16s %s\n", dv, biodv.SetAbout(dv))
	}
}

func recDrivers() {
	ls := biodv.RecDrivers()
	fmt.Printf("Record-DB drivers:\n")
	for _, dv := range ls {
		fmt.Printf("    %-16s %s\n", dv, biodv.RecAbout(dv))
	}
}

func taxDrivers() {
	ls := biodv.TaxDrivers()
	fmt.Printf("Taxonomy drivers:\n")
	for _, dv := range ls {
		fmt.Printf("    %-16s %s\n", dv, biodv.TaxAbout(dv))
	}
}
