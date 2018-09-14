// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package table implements the rec.table command,
// i.e. print a table of records.
package table

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/cmdapp"

	"github.com/pkg/errors"
)

var cmd = &cmdapp.Command{
	UsageLine: `rec.table [--db <database>] [--id] [-e|--exact]
		[-g|--georef] [-n|--noheader] [<taxon>]`,
	Short: "print a table of records",
	Long: `
Command rec.table prints a table (separated by tabs) of the records of
a given taxon in a given database.  If no taxon is given, it will make
the table based on the names given in the standard input.

By default, records assigned to the given taxon (including synonyms and
correct/valid children) will be printed. If the option -e or --exact is
defined, then only the records assigned explicitly to the taxon will be
printed.

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

    -id
    --id
      If set, the search of the taxon will be based on the taxon ID,
      instead of the taxon name. This will affect either if the taxon
      is given on the command line, or read from the standard input.

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

    <taxon>
      If set, the table will be based on the indicated taxon. If the
      name is ambiguous, the ID of the ambiguous taxa will be printed.
      If the option --id is set, it must be a taxon ID instead of a
      taxon name.
	`,
	Run:           run,
	RegisterFlags: register,
}

func init() {
	cmdapp.Add(cmd)
}

var dbName string
var id bool
var exact bool
var georef bool
var nohead bool

func register(c *cmdapp.Command) {
	c.Flag.StringVar(&dbName, "db", "biodv", "")
	c.Flag.BoolVar(&id, "id", false, "")
	c.Flag.BoolVar(&exact, "exact", false, "")
	c.Flag.BoolVar(&exact, "e", false, "")
	c.Flag.BoolVar(&georef, "georef", false, "")
	c.Flag.BoolVar(&georef, "g", false, "")
	c.Flag.BoolVar(&nohead, "noheader", false, "")
	c.Flag.BoolVar(&nohead, "n", false, "")
}

var ids map[string]bool
var rows map[string][]string

func run(c *cmdapp.Command, args []string) error {
	ids = make(map[string]bool)
	rows = make(map[string][]string)
	if dbName == "" {
		dbName = "biodv"
	}
	var param string
	dbName, param = biodv.ParseDriverString(dbName)

	txm, err := biodv.OpenTax(dbName, param)
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	recs, err := biodv.OpenRec(dbName, param)
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	w, err := setupTable()
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	nm := strings.Join(args, " ")
	if nm != "" {
		if err := taxonTable(w, txm, recs, nm); err != nil {
			return errors.Wrap(err, c.Name())
		}
		w.Flush()
		if err := w.Error(); err != nil {
			return errors.Wrap(err, c.Name())
		}
		return nil
	}
	if err := read(w, txm, recs); err != nil {
		return errors.Wrap(err, c.Name())
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return errors.Wrap(err, c.Name())
	}
	return nil
}

func read(w *csv.Writer, txm biodv.Taxonomy, recs biodv.RecDB) error {
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		name := biodv.TaxCanon(s.Text())
		if name == "" {
			continue
		}
		if nm, _ := utf8.DecodeRuneInString(name); nm == '#' || nm == ';' {
			continue
		}
		if err := taxonTable(w, txm, recs, name); err != nil {
			return err
		}
	}
	return s.Err()
}

// GetTaxon returns a taxon from the options.
func getTaxon(txm biodv.Taxonomy, nm string) (biodv.Taxon, error) {
	if id {
		return txm.TaxID(nm)
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

func setupTable() (*csv.Writer, error) {
	w := csv.NewWriter(os.Stdout)
	w.Comma = '\t'
	w.UseCRLF = true
	if !nohead {
		if err := w.Write([]string{"ID", "Taxon", "Lat", "Lon", "Catalog"}); err != nil {
			return nil, err
		}
	}
	return w, nil
}

func taxonTable(w *csv.Writer, txm biodv.Taxonomy, recs biodv.RecDB, name string) error {
	tax, err := getTaxon(txm, name)
	if err != nil {
		return errors.Wrapf(err, "while searching for '%s'", name)
	}
	if tax == nil {
		return nil
	}
	return printSearch(w, tax.ID(), txm, recs)
}

// PrintSearch search for the records of a given taxon
// and print it on the table.
func printSearch(w *csv.Writer, id string, txm biodv.Taxonomy, recs biodv.RecDB) error {
	sr := recs.TaxRecs(id)
	for sr.Scan() {
		r := sr.Record()
		row := []string{
			r.ID(),
			r.Taxon(),
			"NA",
			"NA",
			r.Value(biodv.RecCatalog),
		}

		geo := r.GeoRef()
		if geo.IsValid() {
			row[2] = strconv.FormatFloat(geo.Lat, 'f', -1, 64)
			row[3] = strconv.FormatFloat(geo.Lon, 'f', -1, 64)
		} else if georef {
			continue
		}

		if r.Taxon() != id {
			ids[r.Taxon()] = true
			rows[r.ID()] = row
			continue
		}

		if err := w.Write(row); err != nil {
			return err
		}
	}
	if err := sr.Err(); err != nil {
		return err
	}
	if exact {
		return nil
	}
	return searchChildren(w, id, txm, recs)
}

// PrintStored use stored records of a given taxon
// and print it on the table.
func printStored(w *csv.Writer, id string, txm biodv.Taxonomy, recs biodv.RecDB) error {
	todel := make(map[string]bool)
	for _, row := range rows {
		if row[1] != id {
			continue
		}
		if err := w.Write(row); err != nil {
			return err
		}
		todel[row[0]] = true
	}
	for rid := range todel {
		delete(rows, rid)
	}
	return searchChildren(w, id, txm, recs)
}

// SearchChildren search for reconds on children
func searchChildren(w *csv.Writer, id string, txm biodv.Taxonomy, recs biodv.RecDB) error {
	children, err := biodv.TaxList(txm.Children(id))
	if err != nil {
		return err
	}
	syns, err := biodv.TaxList(txm.Synonyms(id))
	if err != nil {
		return err
	}
	children = append(children, syns...)

	for _, c := range children {
		if ids[c.ID()] {
			if err := printStored(w, c.ID(), txm, recs); err != nil {
				return err
			}
			continue
		}
		if err := printSearch(w, c.ID(), txm, recs); err != nil {
			return err
		}
	}
	return nil
}
