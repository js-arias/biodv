// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package dbfill implements the tax.db.fill command,
// i.e. add taxons from an external DB.
package dbfill

import (
	"fmt"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/cmdapp"
	"github.com/js-arias/biodv/dataset"
	"github.com/js-arias/biodv/taxonomy"

	"github.com/pkg/errors"
)

var cmd = &cmdapp.Command{
	UsageLine: `tax.db.fill -e|--extern <database> [-u|--uprank <rank>]
		[<name>]`,
	Short: "add taxons from an external DB",
	Long: `
Command tax.db.fill adds additional taxons from an external DB to the
current database. By default only synonyms (of any rank), and children
(of taxa at or below species level) will be added.

The option -u or -uprank will add only parents up to the given rank.

Options are:

    -e <database>
    --extern <database>
      A required parameter. It will set the external database.
      To see the available databases use the command ‘db.drivers’.

    -u <rank>
    --uprank <rank>
      If set, parent taxons, up to the given rank, will be added to
      the database.
      Valid ranks are:
        kingdom
        class
        order
        family
        genus
        species

    <name>
      If set, only the indicated taxon, and its descendants (or its
      parents, in the case the -u or –unprank option is defined) will
      be filled.
	`,
	Run:           run,
	RegisterFlags: register,
}

func init() {
	cmdapp.Add(cmd)
}

var extName string
var param string
var upRank string
var mapParent map[string]string

func register(c *cmdapp.Command) {
	c.Flag.StringVar(&extName, "extern", "", "")
	c.Flag.StringVar(&extName, "e", "", "")
	c.Flag.StringVar(&upRank, "uprank", "", "")
	c.Flag.StringVar(&upRank, "u", "", "")
}

func run(c *cmdapp.Command, args []string) (err error) {
	if extName == "" {
		return errors.Errorf("%s: an external database should be defined", c.Name())
	}
	extName, param = biodv.ParseDriverString(extName)

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
	if upRank != "" {
		rank := biodv.GetRank(upRank)
		if rank == biodv.Unranked {
			return err
		}
		upProc(dbs, nm, rank)
		return err
	}
	if nm == "" {
		ls := dbs.db.TaxList("")
		for _, c := range ls {
			fillTaxon(dbs, c)
		}
		return err
	}

	tax := dbs.db.TaxEd(nm)
	if tax == nil {
		return nil
	}
	fillTaxon(dbs, tax)
	return nil
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

func upProc(dbs *databases, nm string, rank biodv.Rank) {
	if nm == "" {
		for {
			ls := dbs.db.TaxList("")
			ok := true
			for _, c := range ls {
				if fillUp(dbs, c, rank) != nil {
					ok = false
				}
			}
			if ok {
				return
			}
		}
	}
	tax := dbs.db.TaxEd(nm)
	for tax != nil {
		tax = fillUp(dbs, tax, rank)
	}
}

func fillUp(dbs *databases, tax *taxonomy.Taxon, rank biodv.Rank) *taxonomy.Taxon {
	if getRank(dbs.db, tax) <= rank {
		return nil
	}
	p := dbs.db.TaxEd(tax.Parent())
	if p != nil {
		return p
	}

	epID := getExternParentID(dbs.ext, getExternID(tax), rank)
	if epID == "" {
		return nil
	}
	p = dbs.db.TaxEd(extName + ":" + epID)
	if p == nil {
		ep := getExternTaxon(dbs.ext, epID)
		if ep == nil {
			return nil
		}
		if ep.Rank() < rank && ep.Rank() != biodv.Unranked {
			return nil
		}
		p = addExtern(dbs, ep, "")
		if p == nil {
			return nil
		}
	}
	if err := tax.Move(p.ID(), tax.IsCorrect()); err != nil {
		fmt.Fprintf(os.Stderr, "warning: when moving %q to %q: %v\n", tax.Name(), p.Name(), err)
		return nil
	}
	return p
}

func getExternParentID(ext biodv.Taxonomy, id string, rank biodv.Rank) string {
	if id == "" {
		return ""
	}
	ep := mapParent[id]
	if ep != "" {
		return ep
	}
	tx := getExternTaxon(ext, id)
	if tx == nil {
		return ""
	}
	if tx.Parent() == "" {
		return ""
	}
	if tx.Rank() <= rank && tx.Rank() != biodv.Unranked {
		return ""
	}
	return tx.Parent()
}

func getExternTaxon(ext biodv.Taxonomy, id string) biodv.Taxon {
	tx, err := ext.TaxID(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: when looking for %s:%s: %v\n", extName, id, err)
		return nil
	}
	if tx == nil {
		fmt.Fprintf(os.Stderr, "warning: when looking for %s:%s: not found\n", extName, id)
		return nil
	}
	mapParent[tx.ID()] = tx.Parent()
	return tx
}

func fillTaxon(dbs *databases, tax *taxonomy.Taxon) {
	// fill descendants
	defer func() {
		desc := dbs.db.TaxList(tax.ID())
		for _, d := range desc {
			fillTaxon(dbs, d)
		}
	}()

	eid := getExternID(tax)
	if eid == "" {
		return
	}

	// add children,
	// only if it at or below species.
	if getRank(dbs.db, tax) >= biodv.Species {
		ls, err := biodv.TaxList(dbs.ext.Children(eid))
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: when looking for %s children: %v\n", tax.Name(), err)
		}
		fillList(dbs, tax, ls)
	}

	// add synonyms
	ls, err := biodv.TaxList(dbs.ext.Synonyms(eid))
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: when looking for %s synonyms: %v\n", tax.Name(), err)
	}
	fillList(dbs, tax, ls)
}

func fillList(dbs *databases, tax *taxonomy.Taxon, ls []biodv.Taxon) {
	for _, d := range ls {
		// skip taxons already in database
		if tx, _ := dbs.db.TaxID(extName + ":" + d.ID()); tx != nil {
			continue
		}
		// skip taxons with invalid names,
		// such as "? spelaea"
		if nm, _ := utf8.DecodeRuneInString(d.Name()); !unicode.IsLetter(nm) {
			continue
		}

		addExtern(dbs, d, tax.ID())
	}
}

func addExtern(dbs *databases, tx biodv.Taxon, parent string) *taxonomy.Taxon {
	tax, err := dbs.db.Add(tx.Name(), parent, tx.Rank(), tx.IsCorrect())
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: when adding %q [%s:%s] to %q: %v\n", tx.Name(), extName, tx.ID(), parent, err)
		return nil
	}
	if err := tax.Set(biodv.TaxExtern, extName+":"+tx.ID()); err != nil {
		fmt.Fprintf(os.Stderr, "warning: when matching %s: %v\n", tax.Name(), err)
	}
	update(dbs, tax, tx)
	return tax
}

func getRank(db *taxonomy.DB, tax *taxonomy.Taxon) biodv.Rank {
	if tax.Rank() != biodv.Unranked {
		return tax.Rank()
	}
	for p := db.TaxEd(tax.Parent()); p != nil; p = db.TaxEd(p.Parent()) {
		if p.Rank() != biodv.Unranked {
			return p.Rank()
		}
	}
	return biodv.Unranked
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
