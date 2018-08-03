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
	"os"
	"path/filepath"
	"strings"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/encoding/stanza"

	"github.com/pkg/errors"
)

// Default database directory and file.
const taxDir = "taxonomy"
const taxFile = "taxonomy.stz"

// DB is a taxonomy database,
// for reading and writing data.
// DB implements the biodv.Taxonomy interface.
type DB struct {
	path    string
	ids     map[string]*Taxon
	changed bool // tire if the database was modified
	root    []*Taxon
}

func (db *DB) Taxon(name string) *biodv.TaxScan {
	sc := biodv.NewTaxScan(1)
	name = biodv.TaxCanon(name)
	if name == "" {
		sc.Add(nil, errors.Errorf("taxonomy: db: taxon: empty taxon name"))
		return sc
	}
	if tax, ok := db.ids[name]; ok {
		sc.Add(tax, nil)
	}
	sc.Add(nil, nil)
	return sc
}

func (db *DB) TaxID(id string) (biodv.Taxon, error) {
	id = biodv.TaxCanon(id)
	if id == "" {
		return nil, errors.Errorf("taxonomy: db: taxon: empty taxon ID")
	}
	if tax, ok := db.ids[id]; ok {
		return tax, nil
	}
	return nil, nil
}

func (db *DB) Children(id string) *biodv.TaxScan {
	sc := biodv.NewTaxScan(20)
	id = biodv.TaxCanon(id)
	var ls []*Taxon
	if id == "" {
		ls = db.root
	} else {
		tax, ok := db.ids[id]
		if !ok {
			sc.Add(nil, nil)
			return sc
		}
		ls = tax.children
	}
	go func() {
		for _, c := range ls {
			if c.IsCorrect() {
				sc.Add(c, nil)
			}
		}
		sc.Add(nil, nil)
	}()
	return sc
}

func (db *DB) Synonyms(id string) *biodv.TaxScan {
	sc := biodv.NewTaxScan(20)
	id = biodv.TaxCanon(id)
	if id == "" {
		sc.Add(nil, errors.Errorf("taxonomy: db: taxon: invalid ID for a synonym"))
		return sc
	}
	tax, ok := db.ids[id]
	if !ok {
		sc.Add(nil, nil)
		return sc
	}
	go func() {
		for _, sn := range tax.children {
			if !sn.IsCorrect() {
				sc.Add(sn, nil)
			}
		}
		sc.Add(nil, nil)
	}()
	return sc
}

// Taxon is a taxon stored in a DB.
// Taxon implements the biodv.Taxon interface.
type Taxon struct {
	data     map[string]string
	parent   *Taxon
	children []*Taxon
}

func (tax *Taxon) Name() string {
	return tax.data[nameKey]
}

func (tax *Taxon) ID() string {
	return tax.data[nameKey]
}

func (tax *Taxon) Parent() string {
	return tax.data[parentKey]
}

func (tax *Taxon) Rank() biodv.Rank {
	return biodv.GetRank(tax.data[rankKey])
}

func (tax *Taxon) IsCorrect() bool {
	if tax.data[correctKey] == "false" {
		return false
	}
	return true
}

// Basic keys for the taxonomy database
const (
	nameKey    = "name"
	parentKey  = "parent"
	rankKey    = "rank"
	correctKey = "correct"
)

func (tax *Taxon) Keys() []string {
	return record(tax.data).Keys()
}

func (tax *Taxon) Value(key string) string {
	key = strings.ToLower(strings.TrimSpace(key))
	if key == "" {
		return ""
	}
	return tax.data[key]
}

// Encode writes a taxon
// into a stanza writer.
func (tax *Taxon) encode(w *stanza.Writer) error {
	fields := []string{nameKey, parentKey, rankKey, correctKey}
	fields = append(fields, tax.Keys()...)
	w.SetFields(fields)

	if err := w.Write(tax.data); err != nil {
		return errors.Wrapf(err, "unable to writer %s", tax.Name())
	}
	for _, c := range tax.children {
		if err := c.encode(w); err != nil {
			return err
		}
	}
	return nil
}

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

func init() {
	biodv.RegisterTax("biodv", open)
}

// open opens a DB
// as a biodv.Taxonomy.
func open(path string) (biodv.Taxonomy, error) {
	return Open(path)
}

// Open opens a DB
// on a given path.
func Open(path string) (*DB, error) {
	db := &DB{
		path: path,
		ids:  make(map[string]*Taxon),
	}
	sc := OpenScanner(path)
	for sc.Scan() {
		r := sc.Taxon()
		_, err := db.Add(r.Name(), r.Parent(), r.Rank(), r.IsCorrect())
		if err != nil {
			sc.Close()
			return nil, err
		}
		if err := sc.Err(); err != nil {
			return nil, err
		}
	}
	db.changed = false
	return db, nil
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
	tax.data[nameKey] = name
	tax.data[parentKey] = parent
	tax.data[rankKey] = rank.String()
	tax.data[correctKey] = "true"
	if !correct {
		tax.data[correctKey] = "false"
	}
	tax.parent = p
	if p == nil {
		db.root = append(db.root, tax)
	} else {
		p.children = append(p.children, tax)
	}
	db.ids[name] = tax
	db.changed = true
	return tax, nil
}

// Commit saves a taxonomy to a file.
func (db *DB) Commit() (err error) {
	if !db.changed {
		return nil
	}

	if _, err := os.Lstat(filepath.Join(db.path, taxDir)); err != nil {
		if err := os.Mkdir(filepath.Join(db.path, taxDir), os.ModeDir|os.ModePerm); err != nil {
			return errors.Wrapf(err, "taxonomy: db: commit: unable to create %s directory", taxDir)
		}
	}

	file := filepath.Join(db.path, taxDir, taxFile)
	f, err := os.Create(file)
	if err != nil {
		return errors.Wrap(err, "taxonomy: db: commit")
	}
	defer func() {
		e1 := f.Close()
		if err == nil {
			err = errors.Wrap(e1, "taxonomy: db: commit")
		}
	}()

	w := stanza.NewWriter(f)
	defer w.Flush()

	for _, tax := range db.root {
		if err := tax.encode(w); err != nil {
			return errors.Wrap(err, "taxonomy: db: commit")
		}
	}
	db.changed = false
	return nil
}
