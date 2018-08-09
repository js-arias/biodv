// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package value implements the tax.value command,
// i.e. get a taxon data value.
package value

import (
	"fmt"
	"os"
	"strings"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/cmdapp"

	"github.com/pkg/errors"
)

var cmd = &cmdapp.Command{
	UsageLine: `tax.value [--db <database>] [--id <value>]
		[-k|--key <key>] <name>`,
	Short: "get a taxon data value",
	Long: `
Command tax.value prints the value of a given key for the indicated
taxon. If no key is given, a list of available keys for the indicated
taxon will be given.

Options are:

    -db <database>
    --db <database>
      If set, the indicated database will be used to search the
      indicated taxon.
      Available databases are:
        biodv	default database (on current directory)
        gbif	GBIF webservice (requires internet connection)

    -id <value>
    --id <value>
      If set, then it will search the indicated taxon.

    -k <key>
    --key <key>
      If set, the value of the indicated key will be printed.

    <name>
      If set, the indicated taxon will be searched. If the name is
      ambiguous, the ID of the ambiguous taxa will be printed.
	`,
	Run:           run,
	RegisterFlags: register,
}

func init() {
	cmdapp.Add(cmd)
}

var dbName string
var id string
var key string

func register(c *cmdapp.Command) {
	c.Flag.StringVar(&dbName, "db", "biodv", "")
	c.Flag.StringVar(&id, "id", "", "")
	c.Flag.StringVar(&key, "key", "", "")
	c.Flag.StringVar(&key, "k", "", "")
}

func run(c *cmdapp.Command, args []string) error {
	if dbName == "" {
		dbName = "biodv"
	}
	var param string
	dbName, param = biodv.ParseDriverString(dbName)

	nm := strings.Join(args, " ")
	if id == "" && nm == "" {
		return errors.Errorf("%s: either a --id or a taxon name, should be given", c.Name())
	}

	db, err := biodv.OpenTax(dbName, param)
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	var tax biodv.Taxon
	if id != "" {
		tax, err = db.TaxID(id)
		if err != nil {
			return errors.Wrap(err, c.Name())
		}
	} else {
		ls, err := biodv.TaxList(db.Taxon(nm))
		if err != nil {
			return errors.Wrap(err, c.Name())
		}
		if len(ls) == 0 {
			return nil
		}
		if len(ls) > 1 {
			fmt.Fprintf(os.Stderr, "ambiguous name:\n")
			for _, tx := range ls {
				fmt.Fprintf(os.Stderr, "id:%s\t%s %s\t", tx.ID(), tx.Name(), tx.Value(biodv.TaxAuthor))
				if tx.IsCorrect() {
					fmt.Fprintf(os.Stderr, "correct name\n")
				} else {
					fmt.Fprintf(os.Stderr, "synonym\n")
				}
			}
			return nil
		}
		tax = ls[0]
	}

	if key == "" {
		ls := []string{"name", "id", "rank", "correct"}
		ls = append(ls, tax.Keys()...)
		for _, k := range ls {
			fmt.Printf("%s\n", k)
		}
		return nil
	}
	switch strings.ToLower(key) {
	case "name":
		fmt.Printf("%s\n", tax.Name())
	case "id":
		fmt.Printf("%s\n", tax.ID())
	case "rank":
		fmt.Printf("%s\n", tax.Rank())
	case "correct":
		fmt.Printf("%v\n", tax.IsCorrect())
	case biodv.TaxExtern:
		ext := tax.Value(key)
		for _, e := range strings.Fields(ext) {
			fmt.Printf("%s\n", e)
		}
	default:
		fmt.Printf("%s\n", tax.Value(key))
	}
	return nil
}
