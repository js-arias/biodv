// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package dbupdate implements the tax.db.update command,
// i.e. update taxon information from an extern DB.
package dbupdate

import (
	"fmt"
	"os"
	"strings"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/cmdapp"
	"github.com/js-arias/biodv/taxonomy"

	"github.com/pkg/errors"
)

var cmd = &cmdapp.Command{
	UsageLine: `tax.db.update -e|--extern <database> [-m|--match] [<name>]`,
	Short:     "update taxon information from an extern DB",
	Long: `
Command tax.db.update reads an external database and update the
additional fields stored on the external database. Neither the name,
nor the rank, the parent of the correct-synonym status will be
modified.

If option -m, or --match is used, only the external ID will be set.

If a taxon is defined, then only that taxon and its descendants will
be updated.

Options are:

    -e <database>
    --extern <database>
      A requiered parameter. It will set the external database.
      Available databases are:
        gbif	GBIF webservice (requires internet connection)

    -m
    --match
      If set, only the external ID, and no other data, will be stored.

    <name>
      If set, only the indicated taxon, and its descendants, will be
      updated.
	`,
	Run:           run,
	RegisterFlags: register,
}

func init() {
	cmdapp.Add(cmd)
}

var extName string
var match bool
var mapParent map[string]string

func register(c *cmdapp.Command) {
	c.Flag.StringVar(&extName, "extern", "", "")
	c.Flag.StringVar(&extName, "e", "", "")
	c.Flag.BoolVar(&match, "match", false, "")
	c.Flag.BoolVar(&match, "m", false, "")
}

func run(c *cmdapp.Command, args []string) (err error) {
	if extName == "" {
		return errors.Errorf("%s: an external database should be defined", c.Name())
	}
	var param string
	extName, param = biodv.ParseDriverString(extName)
	ext, err := biodv.OpenTax(extName, param)
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	db, err := taxonomy.Open("")
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = db.Commit()
		}
		if err != nil {
			err = errors.Wrap(err, c.Name())
		}
	}()

	mapParent = make(map[string]string)

	nm := strings.Join(args, " ")
	if nm == "" {
		ls := db.TaxList("")
		for _, c := range ls {
			procTaxon(db, ext, c)
		}
		return err
	}

	tax := db.TaxEd(nm)
	if tax == nil {
		return nil
	}
	procTaxon(db, ext, tax)
	return err
}

func procTaxon(db *taxonomy.DB, ext biodv.Taxonomy, tax *taxonomy.Taxon) {
	txs := matchFn(db, ext, tax)
	if len(txs) == 1 {
		tx := txs[0]
		mapParent[tx.ID()] = tx.Parent()
		if !match {
			update(tax, tx)
		}
	}
	if len(txs) == 0 {
		fmt.Fprintf(os.Stderr, "warning: when searching %s: not in database\n", tax.Name())
	}
	ls := db.TaxList(tax.ID())
	for _, c := range ls {
		procTaxon(db, ext, c)
	}
	if len(txs) > 1 {
		if tx := matchParent(db, ext, tax, txs); tx != nil {
			if err := tax.Set(biodv.TaxExtern, extName+":"+tx.ID()); err != nil {
				fmt.Fprintf(os.Stderr, "warning: when matching %s: %v\n", tax.Name(), err)
			}
			mapParent[tx.ID()] = tx.Parent()
			if !match {
				update(tax, tx)
			}
			return
		}
		if tx := matchChildren(db, ext, tax, txs); tx != nil {
			if err := tax.Set(biodv.TaxExtern, extName+":"+tx.ID()); err != nil {
				fmt.Fprintf(os.Stderr, "warning: when matching %s: %v\n", tax.Name(), err)
			}
			mapParent[tx.ID()] = tx.Parent()
			if !match {
				update(tax, tx)
			}
			return
		}

		fmt.Fprintf(os.Stderr, "warning: ambiguous name:\n")
		for _, tx := range txs {
			fmt.Fprintf(os.Stderr, "\t%s:%s\t%s %s\t", extName, tx.ID(), tx.Name(), tx.Value(biodv.TaxAuthor))
			if tx.IsCorrect() {
				fmt.Fprintf(os.Stderr, "correct name\n")
			} else {
				fmt.Fprintf(os.Stderr, "synonym\n")
			}
		}
	}
}

func matchFn(db *taxonomy.DB, ext biodv.Taxonomy, tax *taxonomy.Taxon) []biodv.Taxon {
	eid := getExternID(tax)
	if eid != "" {
		if match {
			return nil
		}
		tx, err := ext.TaxID(eid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: when looking for %s: %v\n", tax.Name(), err)
		}
		if tx == nil {
			fmt.Fprintf(os.Stderr, "warning: when looking for %s: not found\n", tax.Name())
		} else {
			mapParent[tx.ID()] = tx.Parent()
		}
		return []biodv.Taxon{tx}
	}
	ls, err := biodv.TaxList(ext.Taxon(tax.Name()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: when searching %s: %v\n", tax.Name(), err)
		return nil
	}
	if len(ls) == 1 {
		tx := ls[0]
		if err := tax.Set(biodv.TaxExtern, extName+":"+tx.ID()); err != nil {
			fmt.Fprintf(os.Stderr, "warning: when matching %s: %v\n", tax.Name(), err)
		}
	}
	return ls
}

// MathcParent search for an extern taxon
// which is parent,
// is also the parent of the current taxon in the database.
func matchParent(db *taxonomy.DB, ext biodv.Taxonomy, tax *taxonomy.Taxon, ls []biodv.Taxon) biodv.Taxon {
	p := db.TaxEd(tax.Parent())
	if p == nil {
		return nil
	}
	eid := getExternID(p)
	if eid == "" {
		return nil
	}

	for _, tx := range ls {
		if tx.Parent() == eid {
			return tx
		}
	}
	return nil
}

// MatchChildren search for an extern taxon
// which is children
// is also a children of the current taxon in the database.
func matchChildren(db *taxonomy.DB, ext biodv.Taxonomy, tax *taxonomy.Taxon, ls []biodv.Taxon) biodv.Taxon {
	children := db.TaxList(tax.ID())
	var pID string
	for _, c := range children {
		eid := getExternID(c)
		if eid == "" {
			continue
		}
		p, ok := mapParent[eid]
		if !ok {
			continue
		}
		if pID == "" {
			pID = p
			continue
		}
		if pID != p {
			// there is no agreement between different taxons
			return nil
		}
	}
	if pID == "" {
		// no parent has assigned to the map
		return nil
	}

	for _, tx := range ls {
		if tx.ID() == pID {
			return tx
		}
	}
	return nil
}

func update(tax *taxonomy.Taxon, tx biodv.Taxon) {
	for _, k := range tx.Keys() {
		v := tx.Value(k)
		if k == biodv.TaxSource {
			v = extName + ":" + v
		}
		if err := tax.Set(k, v); err != nil {
			fmt.Fprintf(os.Stderr, "warning: when updating %s: %v\n", tax.Name(), err)
		}
	}
}

func getExternID(tax *taxonomy.Taxon) string {
	ext := strings.Fields(tax.Value(biodv.TaxExtern))
	for _, e := range ext {
		i := strings.Index(e, ":")
		if i <= 0 {
			continue
		}
		if e[:i] == extName {
			return e[i+1:]
		}
	}
	return ""
}
