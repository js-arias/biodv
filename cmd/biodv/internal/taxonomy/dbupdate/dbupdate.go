// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package dbupdate implements the tax.db.update command,
// i.e. update taxon information from an external DB.
package dbupdate

import (
	"fmt"
	"os"
	"strings"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/cmdapp"
	"github.com/js-arias/biodv/dataset"
	"github.com/js-arias/biodv/taxonomy"

	"github.com/pkg/errors"
)

var cmd = &cmdapp.Command{
	UsageLine: `tax.db.update -e|--extern <database> [-m|--match] [<name>]`,
	Short:     "update taxon information from an external DB",
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
var param string
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

	dbs, err := openDBs()
	if err != nil {
		return errors.Wrap(err, c.Name())
	}
	defer func() {
		if err == nil {
			err = commit(dbs)
		}
		if err != nil {
			err = errors.Wrap(err, c.Name())
		}
	}()

	mapParent = make(map[string]string)

	nm := strings.Join(args, " ")
	if nm == "" {
		ls := dbs.db.TaxList("")
		for _, c := range ls {
			procTaxon(dbs, c)
		}
		return err
	}

	tax := dbs.db.TaxEd(nm)
	if tax == nil {
		return nil
	}
	procTaxon(dbs, tax)
	return err
}

type databases struct {
	db     *taxonomy.DB
	sets   *dataset.DB
	ext    biodv.Taxonomy
	extSet biodv.SetDB
}

func openDBs() (*databases, error) {
	dbs := &databases{}
	extName, param = biodv.ParseDriverString(extName)
	var err error
	dbs.ext, err = biodv.OpenTax(extName, param)
	if err != nil {
		return nil, err
	}

	dbs.db, err = taxonomy.Open("")
	if err != nil {
		return nil, err
	}

	ls := biodv.SetDrivers()
	for _, s := range ls {
		if s == extName {
			dbs.extSet, err = biodv.OpenSet(extName, param)
			if err != nil {
				return nil, err
			}
			break
		}
	}
	if dbs.extSet == nil {
		return dbs, nil
	}
	dbs.sets, err = dataset.Open("")
	if err != nil {
		return nil, err
	}
	return dbs, nil
}

func commit(dbs *databases) error {
	if dbs.sets != nil {
		if err := dbs.sets.Commit(); err != nil {
			return err
		}
	}
	return dbs.db.Commit()
}

func procTaxon(dbs *databases, tax *taxonomy.Taxon) {
	txs := matchFn(dbs, tax)
	if len(txs) == 1 {
		tx := txs[0]
		if tx.IsCorrect() {
			mapParent[tx.ID()] = tx.Parent()
		}
		if !match {
			update(dbs, tax, tx)
		}
	}
	if len(txs) == 0 {
		fmt.Fprintf(os.Stderr, "warning: when searching %s: not in database\n", tax.Name())
	}
	ls := dbs.db.TaxList(tax.ID())
	for _, c := range ls {
		procTaxon(dbs, c)
	}
	if len(txs) > 1 {
		if tx := matchParent(dbs, tax, txs); tx != nil {
			if err := tax.Set(biodv.TaxExtern, extName+":"+tx.ID()); err != nil {
				fmt.Fprintf(os.Stderr, "warning: when matching %s: %v\n", tax.Name(), err)
			}
			if tx.IsCorrect() {
				mapParent[tx.ID()] = tx.Parent()
			}
			if !match {
				update(dbs, tax, tx)
			}
			return
		}
		if tx := matchChildren(dbs, tax, txs); tx != nil {
			if err := tax.Set(biodv.TaxExtern, extName+":"+tx.ID()); err != nil {
				fmt.Fprintf(os.Stderr, "warning: when matching %s: %v\n", tax.Name(), err)
			}
			if tx.IsCorrect() {
				mapParent[tx.ID()] = tx.Parent()
			}
			if !match {
				update(dbs, tax, tx)
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

func matchFn(dbs *databases, tax *taxonomy.Taxon) []biodv.Taxon {
	eid := getExternID(tax)
	if eid != "" {
		if match {
			return nil
		}
		tx, err := dbs.ext.TaxID(eid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: when looking for %s: %v\n", tax.Name(), err)
		}
		if tx == nil {
			fmt.Fprintf(os.Stderr, "warning: when looking for %s: not found\n", tax.Name())
		} else if tx.IsCorrect() {
			mapParent[tx.ID()] = tx.Parent()
		}
		return []biodv.Taxon{tx}
	}
	ls, err := biodv.TaxList(dbs.ext.Taxon(tax.Name()))
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
func matchParent(dbs *databases, tax *taxonomy.Taxon, ls []biodv.Taxon) biodv.Taxon {
	p := dbs.db.TaxEd(tax.Parent())
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
func matchChildren(dbs *databases, tax *taxonomy.Taxon, ls []biodv.Taxon) biodv.Taxon {
	children := dbs.db.TaxList(tax.ID())
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

func update(dbs *databases, tax *taxonomy.Taxon, tx biodv.Taxon) {
	for _, k := range tx.Keys() {
		v := tx.Value(k)
		if k == biodv.TaxSource {
			if err := addDataset(dbs, tax, v); err != nil {
				fmt.Fprintf(os.Stderr, "warning: when updating %s: %v\n", tax.Name(), err)
			}
			continue
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

func addDataset(dbs *databases, tax *taxonomy.Taxon, id string) error {
	if id == "" {
		return nil
	}
	set := dbs.sets.SetEd(extName + ":" + id)
	if set != nil {
		return tax.Set(biodv.TaxSource, set.ID())
	}

	// Adds the new set
	src, err := dbs.extSet.SetID(id)
	if err != nil {
		return err
	}
	if src == nil {
		return nil
	}
	set, err = dbs.sets.Add(src.Title())
	if err != nil {
		return err
	}
	set.Set(biodv.SetExtern, extName+":"+src.ID())
	for _, k := range src.Keys() {
		v := src.Value(k)
		if err := set.Set(k, v); err != nil {
			return err
		}
	}
	return tax.Set(biodv.TaxSource, set.ID())
}
