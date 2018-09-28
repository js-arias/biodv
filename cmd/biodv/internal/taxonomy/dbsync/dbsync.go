// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package dbsync implements the tax.db.sync command,
// i.e. synchronize the local DB to an external taxonomy.
package dbsync

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
	UsageLine: "tax.db.sync -e|--extern <database> [<name>]",
	Short:     "synchronize the local DB to an external taxonomy",
	Long: `
Command tax.db.sync synchronize two taxonomies (i.e. made it compatible),
one external and the local DB.

It require that the local DB has already assigned the external IDs. The
process will only add required taxons (for example, the parent of a
synonym). Taxons without an external ID will be left untouched.

Options are:

    -e <database>
    --extern <database>
      A required parameter. It will set the external database.
      To see the available databases use the command ‘db.drivers’.

    <name>
      If set, only the indicated taxon, and its descendants will
      be synchronized.
	`,
	Run:           run,
	RegisterFlags: register,
}

func init() {
	cmdapp.Add(cmd)
}

var extName string
var param string
var extTaxons map[string]externTaxon
var toMove map[string]bool
var reRank map[string]bool
var maxIts = 5

func register(c *cmdapp.Command) {
	c.Flag.StringVar(&extName, "extern", "", "")
	c.Flag.StringVar(&extName, "e", "", "")
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

	extTaxons = make(map[string]externTaxon)
	toMove = make(map[string]bool)
	reRank = make(map[string]bool)

	nm := strings.Join(args, " ")
	if nm == "" {
		ls := dbs.db.TaxList("")
		for _, c := range ls {
			addTaxons(dbs, c)
		}
	} else {
		tax := dbs.db.TaxEd(nm)
		if tax == nil {
			return err
		}
		addTaxons(dbs, tax)
	}

	if err = makeMoves(dbs); err != nil {
		return err
	}
	if err = makeRankUpdates(dbs.db); err != nil {
		return err
	}
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

// MakeMoves make all movements
func makeMoves(dbs *databases) error {
	for i := 0; i < maxIts; i++ {
		if len(toMove) == 0 {
			break
		}
		del := make(map[string]bool)
		for id := range toMove {
			tax := dbs.db.TaxEd(id)
			if move(dbs, tax) {
				del[id] = true
			}
		}
		for id := range del {
			delete(toMove, id)
		}
	}
	if len(toMove) > 0 {
		for id := range toMove {
			fmt.Fprintf(os.Stderr, "unomoved: %s\n", id)
		}
		return errors.Errorf("%d taxons not moved", len(toMove))
	}
	return nil
}

// MakeRankUpdates update the ranks of all taxons.
func makeRankUpdates(db *taxonomy.DB) error {
	for i := 0; i < maxIts; i++ {
		if len(reRank) == 0 {
			break
		}
		del := make(map[string]bool)
		for id := range reRank {
			tax := db.TaxEd(id)
			if updateRank(tax) {
				del[id] = true
			}
		}
		for id := range del {
			delete(reRank, id)
		}
	}
	if len(reRank) > 0 {
		for id := range reRank {
			fmt.Fprintf(os.Stderr, "left unranked: %s\n", id)
		}
		return errors.Errorf("%d taxons left unranked", len(reRank))
	}
	return nil
}

type externTaxon struct {
	parent  string
	rank    biodv.Rank
	correct bool
}

// AddTaxons add taxons to maps indicating
// if they are to be reranked,
// set a new status
// or moved to a new parent.
func addTaxons(dbs *databases, tax *taxonomy.Taxon) {
	tx, ok := getExternTaxon(dbs.ext, getExternID(tax))
	if ok {
		if tx.rank != tax.Rank() {
			toMove[tax.ID()] = true
			reRank[tax.ID()] = true

			// unranked taxons can be moved freely
			if err := tax.SetRank(biodv.Unranked); err != nil {
				fmt.Fprintf(os.Stderr, "warning: %v\n", err)
			}
		}
		if tx.correct != tax.IsCorrect() {
			toMove[tax.ID()] = true
		}
		if tx.parent != "" {
			p := dbs.db.TaxEd(extName + ":" + tx.parent)
			if p != nil && p.ID() != tax.Parent() {
				toMove[tax.ID()] = true
			}
		}
	}

	for _, d := range dbs.db.TaxList(tax.ID()) {
		addTaxons(dbs, d)
	}
}

// Move will move a taxon to its new parent.
// If the taxon is set as a synonym,
// then it will add the new parent,
// if it does not exist.
func move(dbs *databases, tax *taxonomy.Taxon) bool {
	tx, ok := extTaxons[getExternID(tax)]
	if !ok {
		return true
	}

	var p *taxonomy.Taxon
	if !tx.correct {
		p = getSeniorTaxon(dbs, tx.parent)
	} else {
		p = dbs.db.TaxEd(extName + ":" + tx.parent)
	}
	if p == nil {
		return true
	}
	if !p.IsCorrect() {
		return false
	}
	if err := tax.Move(p.ID(), tx.correct); err != nil {
		return false
	}
	return true
}

func getSeniorTaxon(dbs *databases, parent string) *taxonomy.Taxon {
	p := dbs.db.TaxEd(extName + ":" + parent)
	if p != nil {
		return p
	}
	et, err := dbs.ext.TaxID(parent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: when looking for %s:%s: %v\n", extName, parent, err)
		return nil
	}
	if et == nil {
		fmt.Fprintf(os.Stderr, "warning: when looking for %s:%s: not found\n", extName, parent)
		return nil
	}
	extTaxons[parent] = externTaxon{et.Parent(), et.Rank(), et.IsCorrect()}

	// The taxon exists,
	// but it was not matched
	// or matched to another taxon,
	// so it can be say that is the same taxon.
	p = dbs.db.TaxEd(et.Name())
	if p != nil {
		fmt.Fprintf(os.Stderr, "warning: ambiguous parent %q [%s:%s], on DB [%s:%s]", et.Name(), extName, et.ID(), extName, getExternID(p))
		return nil
	}

	pID := ""
	if gp := dbs.db.TaxEd(extName + ":" + et.Parent()); gp != nil {
		pID = gp.ID()
	}

	p, err = dbs.db.Add(et.Name(), pID, et.Rank(), et.IsCorrect())
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: when adding %q [%s:%s] to %q: %v\n", et.Name(), extName, et.ID(), pID, err)
		return nil
	}
	if err := p.Set(biodv.TaxExtern, extName+":"+et.ID()); err != nil {
		fmt.Fprintf(os.Stderr, "warning: when matching %s: %v\n", p.Name(), err)
	}
	for _, k := range et.Keys() {
		v := et.Value(k)
		if k == biodv.TaxSource {
			if err := addDataset(dbs, p, v); err != nil {
				fmt.Fprintf(os.Stderr, "warning: when updating %s: %v\n", p.Name(), err)
			}
			continue
		}
		if err := p.Set(k, v); err != nil {
			fmt.Fprintf(os.Stderr, "warning: when updating %s: %v\n", p.Name(), err)
		}
	}
	return p
}

// updateRank sets the new rank of a taxon.
func updateRank(tax *taxonomy.Taxon) bool {
	tx, ok := extTaxons[getExternID(tax)]
	if !ok {
		return true
	}
	if err := tax.SetRank(tx.rank); err != nil {
		fmt.Fprintf(os.Stderr, "warning: %v\n", err)
		return false
	}
	return true
}

func getExternTaxon(ext biodv.Taxonomy, id string) (externTaxon, bool) {
	if id == "" {
		return externTaxon{}, false
	}
	tx, err := ext.TaxID(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: when looking for %s:%s: %v\n", extName, id, err)
		return externTaxon{}, false
	}
	if tx == nil {
		fmt.Fprintf(os.Stderr, "warning: when looking for %s:%s: not found\n", extName, id)
		return externTaxon{}, false
	}
	et := externTaxon{tx.Parent(), tx.Rank(), tx.IsCorrect()}
	extTaxons[id] = et
	return et, true
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
