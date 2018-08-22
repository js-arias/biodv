// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package table implements the rec.table command,
// i.e. print a table of records.
package table

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/cmdapp"

	"github.com/pkg/errors"
)

var cmd = &cmdapp.Command{
	UsageLine: `rec.table [--db <database>] [--id <value>] [-e|--exact]
		[-g|--georef] [-n|--noheader] [<name>]`,
	Short: "print a table of records",
	Long: `
Command rec.table prints a table (separated by tabs) of the records of
a given taxon in a given database.

If the option -e or --exact is defined, and the database returns all the
records of a given taxon (including synonyms and correct/valid children)
then only the records assigned explicitly to the taxon will be printed.

If the option -g or --georef is defined, only records with valid
georeferences will be printed.

By default, the table will be printed with the column header. If the
option -n or --noheader is defined, then no header will be printed. The
order of columns is:
	ID       record ID
	Taxon    Taxon ID
	Lat      Geographical latitude
	Lon      Geographical longitude
	Catalog  Catalog code of the record

Options are:

    -db <database>
    --db <database>
      If set, the indicated database will be used to produce the table.
      To see the available databases use the command ‘db.drivers’.
      The database should include drivers for a taxonomy and records.

    -id <value>
    --id <value>
      If set, the table will be based on the indicated taxon.

    -e
    --exact
      If set, only the records explicitly assigned to the indicated
      taxon will be printed.

    -g
    --georef
      If set, only the records with a valid georeference will be
      printed.

    -n
    --noheader
      If set, the table will be printed without the columns header.

    <name>
     If set, the table will be based on the indicated taxon. If the
     name is ambiguous, the ID of the ambigous taxa will be printed.
	`,
	Run:           run,
	RegisterFlags: register,
}

func init() {
	cmdapp.Add(cmd)
}

var dbName string
var id string
var exact bool
var georef bool
var nohead bool

func register(c *cmdapp.Command) {
	c.Flag.StringVar(&dbName, "db", "", "")
	c.Flag.StringVar(&id, "id", "", "")
	c.Flag.BoolVar(&exact, "exact", false, "")
	c.Flag.BoolVar(&exact, "e", false, "")
	c.Flag.BoolVar(&georef, "georef", false, "")
	c.Flag.BoolVar(&georef, "g", false, "")
	c.Flag.BoolVar(&nohead, "noheader", false, "")
	c.Flag.BoolVar(&nohead, "n", false, "")
}

func run(c *cmdapp.Command, args []string) error {
	if dbName == "" {
		return errors.Errorf("%s: no database defined", c.Name())
	}
	var param string
	dbName, param = biodv.ParseDriverString(dbName)

	nm := strings.Join(args, " ")
	if id == "" && nm == "" {
		return errors.Errorf("%s: either a --id or a taxon name, should be given", c.Name())
	}

	txm, err := biodv.OpenTax(dbName, param)
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	recs, err := biodv.OpenRec(dbName, param)
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	tax, err := getTaxon(txm, nm)
	if err != nil {
		return errors.Wrap(err, c.Name())
	}
	if err := printTable(tax, recs); err != nil {
		return errors.Wrap(err, c.Name())
	}
	return nil
}

// GetTaxon returns a taxon from the options.
func getTaxon(txm biodv.Taxonomy, nm string) (biodv.Taxon, error) {
	if id != "" {
		return txm.TaxID(id)
	}
	ls, err := biodv.TaxList(txm.Taxon(nm))
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

func printTable(tax biodv.Taxon, recs biodv.RecDB) error {
	w := csv.NewWriter(os.Stdout)
	w.Comma = '\t'
	w.UseCRLF = true

	sr := recs.TaxRecs(tax.ID())
	i := 0
	for sr.Scan() {
		if i == 0 && !nohead {
			if err := w.Write([]string{"ID", "Taxon", "Lat", "Lon", "Catalog"}); err != nil {
				return err
			}
		}

		r := sr.Record()
		row := []string{
			r.ID(),
			r.Taxon(),
			"NA",
			"NA",
			r.Value(biodv.RecCatalog),
		}

		if exact && r.ID() != tax.ID() {
			continue
		}

		geo := r.GeoRef()
		if geo.IsValid() {
			row[2] = strconv.FormatFloat(geo.Lat, 'f', -1, 64)
			row[3] = strconv.FormatFloat(geo.Lon, 'f', -1, 64)
		} else if georef {
			continue
		}

		if err := w.Write(row); err != nil {
			return err
		}
		i++
	}

	w.Flush()
	return w.Error()
}
