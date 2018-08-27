// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package dbadd implements the rec.db.add command,
// i.e. add records from an external DB.
package dbadd

import (
	"fmt"
	"os"
	"strings"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/cmdapp"
	"github.com/js-arias/biodv/records"

	"github.com/pkg/errors"
)

var cmd = &cmdapp.Command{
	UsageLine: "rec.db.add -e|--extern <database> [-g|--georef] [<name>]",
	Short:     "add records from an external DB",
	Long: `
Command rec.db.add adds one or more records from the indicated database.
Only the taxons on the local taxon database that are already matched to
the external DB will be search. Only taxons at or below species rank will
be searched.

If the option -g or --georef  is defined, only records with valid
georeferences will be added.

Options are:

    -e <database>
    --extern <database>
      A required parameter. If will set the external database.
      To see the available databases use the command ‘db.drivers’.

    -g
    --georef
      If set, only the records with a valid georefence will be added.

    <name>
      If set, only the records for the indicated taxon (and its
      descendats) will be added.
	`,
	Run:           run,
	RegisterFlags: register,
}

func init() {
	cmdapp.Add(cmd)
}

var extName string
var georef bool

func register(c *cmdapp.Command) {
	c.Flag.StringVar(&extName, "extern", "", "")
	c.Flag.StringVar(&extName, "e", "", "")
	c.Flag.BoolVar(&georef, "georef", false, "")
	c.Flag.BoolVar(&georef, "g", false, "")
}

var ids = make(map[string][]biodv.Record)

func run(c *cmdapp.Command, args []string) error {
	if extName == "" {
		return errors.Errorf("%s: an external database should be defined", c.Name())
	}
	var param string
	extName, param = biodv.ParseDriverString(extName)
	ext, err := biodv.OpenRec(extName, param)
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	txm, err := biodv.OpenTax("biodv", "")
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	recs, err := records.Open("")
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	if len(args) > 0 {
		nm := strings.Join(args, " ")
		tax, _ := txm.TaxID(nm)
		if tax == nil {
			return nil
		}
		procTaxon(txm, ext, recs, tax)
		if err := recs.Commit(); err != nil {
			return errors.Wrap(err, c.Name())
		}
		return nil
	}

	ls, err := biodv.TaxList(txm.Children(""))
	if err != nil {
		return errors.Wrap(err, c.Name())
	}
	for _, tax := range ls {
		procTaxon(txm, ext, recs, tax)
	}
	if err := recs.Commit(); err != nil {
		return errors.Wrap(err, c.Name())
	}
	return nil

}

// ProcTaxon add records of a given taxon.
func procTaxon(txm biodv.Taxonomy, ext biodv.RecDB, recs *records.DB, tax biodv.Taxon) {
	if getRank(txm, tax) < biodv.Species {
		procChildren(txm, ext, recs, tax)
		return
	}

	eid := getExternID(tax)
	if eid == "" {
		procChildren(txm, ext, recs, tax)
		return
	}

	// records for this taxon
	// are already stored
	if ls := ids[eid]; ls != nil {
		addStored(recs, tax, ls)
		delete(ids, eid)
		procChildren(txm, ext, recs, tax)
		return
	}

	sr := ext.TaxRecs(eid)
	for sr.Scan() {
		r := sr.Record()
		geo := r.GeoRef()
		if georef && !geo.IsValid() {
			continue
		}

		if r.Taxon() != eid {
			ls := ids[r.Taxon()]
			ls = append(ls, r)
			ids[r.Taxon()] = ls
			continue
		}

		addRecord(recs, tax, r)
	}
	if err := sr.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: when reading results for %s: %v\n", tax.Name(), err)
	}
	procChildren(txm, ext, recs, tax)
}

// AddStored adds stored records of a given taxon.
func addStored(recs *records.DB, tax biodv.Taxon, ls []biodv.Record) {
	for _, r := range ls {
		addRecord(recs, tax, r)
	}
}

func addRecord(recs *records.DB, tax biodv.Taxon, r biodv.Record) {
	eid := extName + ":" + r.ID()
	cat := r.Value(biodv.RecCatalog)

	// check if the record is already added
	if ot, _ := recs.RecID(eid); ot != nil {
		return
	}

	// if the catalog number is already in use,
	// print a warning and ignore the catalog code.
	if ot, _ := recs.RecID(cat); ot != nil {
		fmt.Fprintf(os.Stderr, "warning: catalog code %q [(%s) on %s - %s] already in use [by %s - %s]\n", cat, eid, tax.Name(), r.Value(biodv.RecDataset), ot.Taxon(), ot.Value(biodv.RecDataset))
		cat = ""
	}
	geo := r.GeoRef()

	rec, err := recs.Add(tax.Name(), "", cat, r.Basis(), geo.Lat, geo.Lon)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: when adding %q [%s]: %v\n", eid, tax.Name(), err)
		return
	}
	rec.SetCollEvent(r.CollEvent())
	rec.SetGeoRef(r.GeoRef(), records.Precision)
	keys := r.Keys()
	for _, k := range keys {
		if k == biodv.RecCatalog {
			continue
		}
		v := r.Value(k)
		if v == "" {
			continue
		}
		if k == biodv.RecDataset {
			v = extName + ":" + v
		}
		if err := rec.Set(k, v); err != nil {
			fmt.Fprintf(os.Stderr, "warning: when updating %q [%s]: %v\n", eid, tax.Name(), err)
		}
	}
	if err := rec.Set(biodv.RecExtern, eid); err != nil {
		fmt.Fprintf(os.Stderr, "warning: when updating %q [%s]: %v\n", eid, tax.Name(), err)
	}
}

func procChildren(txm biodv.Taxonomy, ext biodv.RecDB, recs *records.DB, tax biodv.Taxon) {
	children, _ := biodv.TaxList(txm.Children(tax.ID()))
	syns, _ := biodv.TaxList(txm.Synonyms(tax.ID()))
	children = append(children, syns...)

	for _, c := range children {
		procTaxon(txm, ext, recs, c)
	}
}

// GetRank returns the rank of a taxon,
// or the rank of a ranked parent
// if the taxon is unranked.
func getRank(txm biodv.Taxonomy, tax biodv.Taxon) biodv.Rank {
	for p := tax; p != nil; p, _ = txm.TaxID(p.Parent()) {
		if p.Rank() != biodv.Unranked {
			return p.Rank()
		}
	}
	return biodv.Unranked
}

func getExternID(tax biodv.Taxon) string {
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