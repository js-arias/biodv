// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package info implements the set.info command,
// i.e. print dataset information.
package info

import (
	"fmt"
	"strings"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/cmdapp"

	"github.com/pkg/errors"
)

var cmd = &cmdapp.Command{
	UsageLine: "set.info [-db <database>] <value>",
	Short:     "print dataset information",
	Long: `
Command set.info prints the information data available for a dataset,
in a given database.

Options are:

    -db <database>
    --db <database>
      If set, the indicated database will be used to extract the
      dataset information.
      To see the available databases use the command ‘db.drivers’.

    <value>
      The ID of the dataset.
	`,
	Run:           run,
	RegisterFlags: register,
}

func init() {
	cmdapp.Add(cmd)
}

var dbName string

func register(c *cmdapp.Command) {
	c.Flag.StringVar(&dbName, "db", "", "")
}

func run(c *cmdapp.Command, args []string) error {
	if dbName == "" {
		return nil
	}
	var param string
	dbName, param = biodv.ParseDriverString(dbName)

	if len(args) != 1 {
		return errors.Errorf("%s: a dataset ID should be given", c.Name())
	}
	id := args[0]

	dsets, err := biodv.OpenSet(dbName, param)
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	set, err := dsets.SetID(id)
	if err != nil {
		return errors.Wrap(err, c.Name())
	}
	print(set)
	return nil
}

func print(set biodv.Dataset) {
	fmt.Printf("Title: %s\n", set.Title())
	fmt.Printf("\tID:\t%s\n", set.ID())
	if about := set.Value(biodv.SetAboutKey); about != "" {
		fmt.Printf("Description:\n%s\n", strings.TrimSpace(about))
	}
	for _, k := range set.Keys() {
		if k == biodv.SetAboutKey {
			continue
		}
		v := set.Value(k)
		if v == "" {
			continue
		}
		fmt.Printf("%s:\t%s\n", k, v)
	}
}
