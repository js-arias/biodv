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

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/cmdapp"
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
      Available databases are:
        gbif	GBIF webservice (requires internet connection)

    -u <rank>
    --uprank <rank>
      If set, parent taxons, up to the given rank, will be added to
      the database.

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
	var param string
	extName, param = biodv.ParseDriverString(extName)
	ext, err := biodv.OpenTax(extName, param)
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	db, err := taxonomy.Open("")
	if err != nil {
		return errors.Wrap(err, c.Name())
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
	if upRank != "" {
		rank := biodv.GetRank(upRank)
		if rank == biodv.Unranked {
			return err
		}
		upProc(db, ext, nm, rank)
		return err
	}
	if nm == "" {
		ls := db.TaxList("")
		for _, c := range ls {
			fillTaxon(db, ext, c)
		}
		return err
	}

	tax := db.TaxEd(nm)
	if tax == nil {
		return nil
	}
	fillTaxon(db, ext, tax)
	return err
}

func upProc(db *taxonomy.DB, ext biodv.Taxonomy, nm string, rank biodv.Rank) {
	if nm == "" {
		for {
			ls := db.TaxList("")
			ok := true
			for _, c := range ls {
				if fillUp(db, ext, c, rank) != nil {
					ok = false
				}
			}
			if ok {
				return
			}
		}
	}
	tax := db.TaxEd(nm)
	for tax != nil {
		tax = fillUp(db, ext, tax, rank)
	}
}

func fillUp(db *taxonomy.DB, ext biodv.Taxonomy, tax *taxonomy.Taxon, rank biodv.Rank) *taxonomy.Taxon {
	if getRank(db, tax) <= rank {
		return nil
	}
	p := db.TaxEd(tax.Parent())
	if p != nil {
		return p
	}

	epID := getExternParentID(ext, getExternID(tax), rank)
	if epID == "" {
		return nil
	}
	p = db.TaxEd(extName + ":" + epID)
	if p == nil {
		ep := getExternTaxon(ext, epID)
		if ep == nil {
			return nil
		}
		if ep.Rank() < rank && ep.Rank() != biodv.Unranked {
			return nil
		}
		p = addExtern(db, ep, "")
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

func fillTaxon(db *taxonomy.DB, ext biodv.Taxonomy, tax *taxonomy.Taxon) {
	eid := getExternID(tax)
	if eid == "" {
		return
	}

	// add synonyms
	ls, err := biodv.TaxList(ext.Synonyms(eid))
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: when looking for %s synonyms: %v\n", tax.Name(), err)
	}
	fillList(db, tax, ls)

	// add children,
	// only if it at or below species.
	if getRank(db, tax) < biodv.Species {
		return
	}
	ls, err = biodv.TaxList(ext.Children(eid))
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: when looking for %s children: %v\n", tax.Name(), err)
	}
	fillList(db, tax, ls)

	desc := db.TaxList(tax.ID())
	for _, d := range desc {
		fillTaxon(db, ext, d)
	}
}

func fillList(db *taxonomy.DB, tax *taxonomy.Taxon, ls []biodv.Taxon) {
	for _, d := range ls {
		// skip taxons already in database
		if tx, _ := db.TaxID(extName + ":" + d.ID()); tx != nil {
			continue
		}
		addExtern(db, d, tax.ID())
	}
}

func addExtern(db *taxonomy.DB, tx biodv.Taxon, parent string) *taxonomy.Taxon {
	tax, err := db.Add(tx.Name(), parent, tx.Rank(), tx.IsCorrect())
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: when adding %q [%s:%s] to %q: %v\n", tx.Name(), extName, tx.ID(), parent, err)
		return nil
	}
	if err := tax.Set(biodv.TaxExtern, extName+":"+tx.ID()); err != nil {
		fmt.Fprintf(os.Stderr, "warning: when matching %s: %v\n", tax.Name(), err)
	}
	update(tax, tx)
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
