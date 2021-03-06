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
	UsageLine: `tax.value [--db <database>] [--id] [-k|--key <key>]
		<taxon>`,
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
      To see the available databases use the command ‘db.drivers’.
      The default biodv database on the current directory.

    -id
    --id
      If set, the search of the taxon will be based on the taxon ID,
      instead of the taxon name.

    -k <key>
    --key <key>
      If set, the value of the indicated key will be printed.

    <taxon>
      A required parameter. Indicates the taxon for which the value
      will be printed. If the name is ambiguous, the ID of the ambiguous
      taxa will be printed. If the option --id is set, it must be a
      taxon ID instead of a taxon name.
	`,
	Run:           run,
	RegisterFlags: register,
}

func init() {
	cmdapp.Add(cmd)
}

var dbName string
var id bool
var key string

func register(c *cmdapp.Command) {
	c.Flag.StringVar(&dbName, "db", "biodv", "")
	c.Flag.BoolVar(&id, "id", false, "")
	c.Flag.StringVar(&key, "key", "", "")
	c.Flag.StringVar(&key, "k", "", "")
}

func run(c *cmdapp.Command, args []string) error {
	if dbName == "" {
		dbName = "biodv"
	}

	nm := strings.Join(args, " ")
	tax, err := getTaxon(nm)
	if err != nil {
		return errors.Wrap(err, c.Name())
	}
	if tax == nil {
		return nil
	}

	if key == "" {
		ls := []string{"name", "id", "rank", "correct", "parent"}
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
	case "parent":
		fmt.Printf("%s\n", tax.Parent())
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

// GetTaxon returns a taxon from the options.
func getTaxon(nm string) (biodv.Taxon, error) {
	if nm == "" {
		return nil, errors.New("either a --id or a taxon name, should be given")
	}

	var param string
	dbName, param = biodv.ParseDriverString(dbName)

	db, err := biodv.OpenTax(dbName, param)
	if err != nil {
		return nil, err
	}

	if id {
		return db.TaxID(nm)
	}
	ls, err := biodv.TaxList(db.Taxon(nm))
	if err != nil {
		return nil, err
	}
	if len(ls) == 0 {
		return nil, nil
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
		return nil, nil
	}
	return ls[0], nil
}
