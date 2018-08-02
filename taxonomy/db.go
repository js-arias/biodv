// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package taxonomy implements
// a hierarchical,
// linnean ranked taxonomy.
package taxonomy

import (
	"github.com/js-arias/biodv"

	"github.com/pkg/errors"
)

// DB is a taxonomy database,
// for reading and writing data.
type DB struct {
	ids map[string]*Taxon
}

// Taxon is a taxon stored in a DB.
// Taxon implements the biodv.Taxon interface.
type Taxon struct {
	data   map[string]string
	parent *Taxon
}

func (tax *Taxon) Name() string {
	return tax.data["name"]
}

func (tax *Taxon) ID() string {
	return tax.data["name"]
}

func (tax *Taxon) Parent() string {
	return tax.data["parent"]
}

func (tax *Taxon) Rank() biodv.Rank {
	return biodv.GetRank(tax.data["rank"])
}

func (tax *Taxon) IsCorrect() bool {
	if tax.data["correct"] == "false" {
		return false
	}
	return true
}

func (tax *Taxon) Keys() []string          { return nil }
func (tax *Taxon) Value(key string) string { return "" }

// IsConsistentDown returns true if a rank is consistent
// in a taxonomy.
func (tax *Taxon) isConsistentDown(correct bool, rank biodv.Rank) bool {
	if rank == biodv.Unranked {
		return true
	}
	for p := tax; p != nil; p = p.parent {
		r := p.Rank()
		if r == biodv.Unranked {
			continue
		}
		if rank > r {
			return true
		}
		if rank == r && !correct {
			return true
		}
		return false
	}
	return true
}

// Open opens a DB.
func Open() *DB {
	return &DB{ids: make(map[string]*Taxon)}
}

// Add adds a new taxon name to a DB.
func (db *DB) Add(name, parent string, rank biodv.Rank, correct bool) (*Taxon, error) {
	name = biodv.TaxCanon(name)
	if name == "" {
		return nil, errors.Errorf("taxonomy: db: add: empty taxon name")
	}
	if _, dup := db.ids[name]; dup {
		return nil, errors.Errorf("taxonomy: db: add %q: taxon already in database")
	}
	parent = biodv.TaxCanon(parent)
	var p *Taxon
	if parent != "" {
		var ok bool
		p, ok = db.ids[parent]
		if !ok {
			return nil, errors.Errorf("taxonomy: db: add %q: parent %q not in database", name, parent)
		}
		if !p.IsCorrect() {
			return nil, errors.Errorf("taxonomy: db: add %q: parent %q is a synonym", name, parent)
		}
		if !p.isConsistentDown(correct, rank) {
			return nil, errors.Errorf("taxonomy: db: add %q: inconsistent rank", name)
		}
	}
	if p == nil && !correct {
		return nil, errors.Errorf("taxonomy: db: add %q: synonym without a parent")
	}
	tax := &Taxon{data: make(map[string]string)}
	tax.data["name"] = name
	tax.data["parent"] = parent
	tax.data["rank"] = rank.String()
	tax.data["correct"] = "true"
	if !correct {
		tax.data["correct"] = "false"
	}
	tax.parent = p
	db.ids[name] = tax
	return tax, nil
}
