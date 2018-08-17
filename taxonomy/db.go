// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package taxonomy

import (
	"os"
	"path/filepath"
	"sort"
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

// Taxon returns a list of taxons with a given name.
// This function is for compatibility with biodv.Taxonomy interface.
//
// When using an editable DB prefer TaxEd.
func (db *DB) Taxon(name string) *biodv.TaxScan {
	sc := biodv.NewTaxScan(1)
	name = getTaxonID(name)
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

// TaxID returns the taxon with a given ID.
//
// When using an editable DB prefer TaxEd.
func (db *DB) TaxID(id string) (biodv.Taxon, error) {
	id = getTaxonID(id)
	if id == "" {
		return nil, errors.Errorf("taxonomy: db: taxon: empty taxon ID")
	}
	if tax, ok := db.ids[id]; ok {
		return tax, nil
	}
	return nil, nil
}

// TaxEd returns an editable Taxon
func (db *DB) TaxEd(id string) *Taxon {
	id = getTaxonID(id)
	if id == "" {
		return nil
	}
	if tax, ok := db.ids[id]; ok {
		return tax
	}
	return nil
}

// GetTaxonID gets a valid ID
// either from a taxon name,
// or an external service.
func getTaxonID(id string) string {
	if getService(id) != "" {
		return strings.ToLower(id)
	}
	return biodv.TaxCanon(id)
}

// TaxList returns a list of taxons.
// It will return all the descendants
// (correct children and synonyms)
// attached to the taxon.
// If the taxon ID is empty,
// it will list the root of the taxonomy.
func (db *DB) TaxList(id string) []*Taxon {
	id = getTaxonID(id)
	if id == "" {
		ls := make([]*Taxon, len(db.root))
		copy(ls, db.root)
		return ls
	}
	tax, ok := db.ids[id]
	if !ok {
		return nil
	}
	ls := make([]*Taxon, len(tax.children))
	copy(ls, tax.children)
	return ls
}

// Children returns a list of taxon children of a given ID,
// if the ID is empty,
// it will return the taxons attached to the root
// of the taxonomy.
//
// When using an editable DB prefer TaxList.
func (db *DB) Children(id string) *biodv.TaxScan {
	sc := biodv.NewTaxScan(20)
	id = getTaxonID(id)
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

// Synonyms returns a list taxons synonyms of a given ID.
//
// When using an editable DB prefer TaxList.
func (db *DB) Synonyms(id string) *biodv.TaxScan {
	sc := biodv.NewTaxScan(20)
	id = getTaxonID(id)
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
	db       *DB
	data     map[string]string
	parent   *Taxon
	children []*Taxon
}

// Name returns the canonical name of the current taxon.
func (tax *Taxon) Name() string {
	return tax.data[nameKey]
}

// ID returns the ID of the current taxon.
func (tax *Taxon) ID() string {
	return tax.data[nameKey]
}

// Parent returns the ID of the taxon's parent.
func (tax *Taxon) Parent() string {
	return tax.data[parentKey]
}

// Rank returns the linnean rank of the current taxon.
func (tax *Taxon) Rank() biodv.Rank {
	return biodv.GetRank(tax.data[rankKey])
}

// IsCorrect returns true if the taxon
// is a correct name
// (i.e. not a synonym).
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

// Keys returns a list of additional fields
// stored in the taxon.
func (tax *Taxon) Keys() []string {
	return record(tax.data).Keys()
}

// Value returns the value
// of an additional field stored in the taxon.
func (tax *Taxon) Value(key string) string {
	key = strings.ToLower(strings.TrimSpace(key))
	if key == "" {
		return ""
	}
	return tax.data[key]
}

// Set sets a value from a given key.
// The value should be transformed into
// a string.
// When an empty string is used as value,
// the stored value will be deleted.
func (tax *Taxon) Set(key, value string) error {
	key = strings.ToLower(strings.Join(strings.Fields(key), "-"))
	if key == "" {
		return nil
	}
	value = strings.TrimSpace(value)

	switch key {
	case nameKey:
		fallthrough
	case parentKey:
		fallthrough
	case rankKey:
		fallthrough
	case correctKey:
		return errors.Errorf("taxonomy: taxon: invalid key value: %s", key)
	case biodv.TaxExtern:
		if value == "" {
			return nil
		}
		srv := getService(value)
		if srv == "" {
			return errors.New("taxonomy: taxon: invalid extern ID value")
		}
		ext := strings.Fields(tax.data[key])
		if srv+":" == value {
			// empty extern ID,
			// deletes the extern ID from database
			for i, e := range ext {
				if srv != getService(e) {
					continue
				}
				delete(tax.db.ids, e)
				n := append(ext[:i], ext[i+1:]...)
				tax.data[key] = strings.Join(n, " ")
				tax.db.changed = true
				return nil
			}
			return nil
		}

		// check if the given ID is already in use
		if _, dup := tax.db.ids[value]; dup {
			return errors.Errorf("taxonomy: taxon: extern ID %s already in use", value)
		}

		// if the service is already set
		for i, e := range ext {
			if srv != getService(e) {
				continue
			}
			delete(tax.db.ids, e)
			tax.db.ids[value] = tax
			ext[i] = value
			tax.data[key] = strings.Join(ext, " ")
			tax.db.changed = true
			return nil
		}

		// the service is new
		ext = append(ext, value)
		sort.Strings(ext)
		tax.db.ids[value] = tax
		tax.data[key] = strings.Join(ext, " ")
		tax.db.changed = true
		return nil
	case biodv.TaxAuthor:
		v := tax.data[key]
		if v == value {
			return nil
		}
		tax.data[key] = value
		if tax.parent == nil {
			sortTaxons(tax.db.root)
		} else {
			sortTaxons(tax.parent.children)
		}
		return nil
	default:
		v := tax.data[key]
		if v == value {
			return nil
		}
		tax.data[key] = value
		tax.db.changed = true
	}
	return nil
}

// SetRank sets the rank of a taxon.
// The new rank should be compatible with the taxonomy.
func (tax *Taxon) SetRank(rank biodv.Rank) error {
	if tax.Rank() == rank {
		return nil
	}
	if !tax.parent.isConsistentDown(tax.IsCorrect(), rank) {
		return errors.Errorf("taxonomy: db: setrank %s: inconsistent parent rank", tax.Name())
	}
	for _, c := range tax.children {
		if !c.isConsistentUp(tax.IsCorrect(), rank) {
			return errors.Errorf("taxonomy: db: setrank %s: inconsistent children rank", tax.Name())
		}
	}
	tax.data[rankKey] = rank.String()
	tax.db.changed = true
	return nil
}

// Move moves a taxon,
// with the indicated valid (true)/synonym(false) status
// to a new parent.
// The taxon should be consistent with the database taxonomy.
// It the taxon is set a synonym
// all of its children will be set as synonyms
// of the new parent.
func (tax *Taxon) Move(parent string, status bool) error {
	parent = getTaxonID(parent)
	p := tax.db.TaxEd(parent)
	if p == tax.parent && status == tax.IsCorrect() {
		return nil
	}
	if p == nil {
		if !status {
			return errors.Errorf("taxonomy: db: move %q: set as synonym without a parent", tax.Name())
		}
		if parent != "" {
			return errors.Errorf("taxonomy: db: move %q: parent %q not in database", tax.Name(), parent)
		}
		parent = ""
	} else {
		if !p.IsCorrect() {
			return errors.Errorf("taxonomy: db: move %q: parent %q is a synonym", tax.Name(), p.Name())
		}
		parent = p.ID()
	}
	if !p.isConsistentDown(status, tax.Rank()) {
		return errors.Errorf("taxonomy: db: move %q: inconsistent rank", tax.Name())
	}

	tax.removeFromParent()

	if status {
		tax.data[correctKey] = "true"
	} else {
		tax.data[correctKey] = "false"
	}

	tax.parent = p
	tax.data[parentKey] = parent
	if p != nil {
		p.children = append(p.children, tax)
		sortTaxons(p.children)
	} else {
		tax.db.root = append(tax.db.root, tax)
		sortTaxons(tax.db.root)
	}
	tax.moveChildren()
	tax.db.changed = true
	return nil
}

// MoveChildren moves the children taxa of a synonym
// to its parent.
func (tax *Taxon) moveChildren() {
	if len(tax.children) == 0 {
		return
	}
	if tax.IsCorrect() {
		return
	}
	if tax.parent == nil {
		return
	}

	for _, c := range tax.children {
		c.data[parentKey] = tax.parent.ID()
		c.parent = tax.parent
	}
	tax.parent.children = append(tax.parent.children, tax.children...)
	sortTaxons(tax.parent.children)
	tax.children = nil
}

// GetService returns the service
// (extern Taxonomy identifier)
// that provides an external ID.
func getService(id string) string {
	i := strings.Index(id, ":")
	if i <= 0 {
		return ""
	}
	return id[:i]
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

// IsConsistentUp returns true if a rank is consistent
// in a taxonomy.
func (tax *Taxon) isConsistentUp(correct bool, rank biodv.Rank) bool {
	if rank == biodv.Unranked {
		return true
	}
	if tax.Rank() == biodv.Unranked {
		for _, c := range tax.children {
			if !c.isConsistentUp(correct, rank) {
				return false
			}
		}
		return true
	}
	if tax.Rank() > rank {
		return true
	}
	if tax.Rank() == rank && !correct {
		return true
	}
	return false
}

// Delete removes a taxon from the taxonomy.
// If rec is true,
// it will also remove all the descendants of the taxon,
// otherwise it will move all the children
// to the parent of the taxon.
// If the taxon is on the root,
// the synonyms will be also removed.
func (tax *Taxon) Delete(rec bool) {
	if rec {
		ls := make([]*Taxon, len(tax.children))
		copy(ls, tax.children)
		for _, c := range ls {
			c.Delete(true)
		}
	}
	tax.removeFromParent()
	tax.remove()
	tax.db.changed = true
}

func (tax *Taxon) removeFromParent() {
	// remove the taxon from its previous parent
	if tax.parent != nil {
		for i, d := range tax.parent.children {
			if d != tax {
				continue
			}
			copy(tax.parent.children[i:], tax.parent.children[i+1:])
			tax.parent.children[len(tax.parent.children)-1] = nil
			tax.parent.children = tax.parent.children[:len(tax.parent.children)-1]
			break
		}
	} else {
		for i, d := range tax.db.root {
			if d != tax {
				continue
			}
			copy(tax.db.root[i:], tax.db.root[i+1:])
			tax.db.root[len(tax.db.root)-1] = nil
			tax.db.root = tax.db.root[:len(tax.db.root)-1]
			break
		}
	}
}

func (tax *Taxon) remove() {
	if tax.parent == nil {
		for _, c := range tax.children {
			if !c.IsCorrect() {
				c.parent = nil
				tax.db.deleteIDs(c)
				continue
			}
			c.data[parentKey] = ""
			c.parent = nil
			tax.db.root = append(tax.db.root, c)
		}
	} else {
		for _, c := range tax.children {
			c.data[parentKey] = tax.parent.ID()
			c.parent = tax.parent
		}
		tax.parent.children = append(tax.parent.children, tax.children...)
		sortTaxons(tax.parent.children)
	}
	tax.children = nil
	tax.parent = nil
	tax.db.deleteIDs(tax)
}

// SortTaxons sorts taxons
// first as correct-valid and synonyms.
// Valid taxons are sorted by name.
// Synonyms are sorted by its taxon year
// as given in the author field.
// If no year is given for the taxon,
// synonym is move down,
// and then sorted by its name.
func sortTaxons(ls []*Taxon) {
	sort.Sort(sortTax(ls))
}

type sortTax []*Taxon

func (s sortTax) Len() int      { return len(s) }
func (s sortTax) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s sortTax) Less(i, j int) bool {
	if s[i].IsCorrect() {
		if !s[j].IsCorrect() {
			return true
		}
		return s[i].Name() < s[j].Name()
	}

	// correct names are always "less" than synonyms
	if s[j].IsCorrect() {
		return false
	}

	yi, yj := biodv.TaxYear(s[i]), biodv.TaxYear(s[j])
	if yi == 0 {
		if yj == 0 {
			return s[i].Name() < s[j].Name()
		}
		return false
	}
	if yj == 0 {
		return true
	}

	if yi == yj {
		return s[i].Name() < s[j].Name()
	}
	return yi < yj
}

func (db *DB) deleteIDs(tax *Taxon) {
	delete(db.ids, tax.ID())

	ext := strings.Fields(tax.Value(biodv.TaxExtern))
	for _, e := range ext {
		delete(db.ids, e)
	}
}

func init() {
	biodv.RegisterTax("biodv", biodv.TaxDriver{open, nil, aboutBiodv})
}

// AboutBiodv returns a simple stated of the purpose of the driver.
func aboutBiodv() string {
	return "the default taxonomy driver"
}

// Open opens a DB
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
	if err := db.scan(OpenScanner(path)); err != nil {
		return nil, err
	}
	db.changed = false
	return db, nil
}

// Scan uses a scanner
// to load a database.
func (db *DB) scan(sc *Scanner) error {
	for sc.Scan() {
		r := sc.Taxon()
		tax, err := db.Add(r.Name(), r.Parent(), r.Rank(), r.IsCorrect())
		if err != nil {
			sc.Close()
			return err
		}
		keys := r.Keys()
		for _, k := range keys {
			if err := tax.Set(k, r.Value(k)); err != nil {
				return err
			}
		}
		if err := sc.Err(); err != nil {
			return err
		}
	}
	return nil
}

// Add adds a new taxon name to a DB.
func (db *DB) Add(name, parent string, rank biodv.Rank, correct bool) (*Taxon, error) {
	name = biodv.TaxCanon(name)
	if name == "" {
		return nil, errors.Errorf("taxonomy: db: add: empty taxon name")
	}
	if _, dup := db.ids[name]; dup {
		return nil, errors.Errorf("taxonomy: db: add %q: taxon already in database", name)
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
		return nil, errors.Errorf("taxonomy: db: add %q: synonym without a parent", name)
	}
	tax := &Taxon{
		db:   db,
		data: make(map[string]string),
	}
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
		sortTaxons(db.root)
	} else {
		p.children = append(p.children, tax)
		sortTaxons(p.children)
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
