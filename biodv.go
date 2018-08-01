// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package biodv contains
// main interfaces and types
// for a basic biodiversity database.
package biodv

import (
	"io"
	"sort"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

// A Taxonomy is a taxonomy database.
type Taxonomy interface {
	// Taxon returns a list of taxons with a given name.
	Taxon(name string) *TaxScan

	// TaxID returns the taxon with a given ID.
	TaxID(id string) (Taxon, error)

	// Synonyms returns a list taxons synonyms of a given ID.
	Synonyms(id string) *TaxScan

	// Children returns a list of taxon children of a given ID,
	// if the ID is empty,
	// it will return the taxons attached to the root
	// of the taxonomy.
	Children(id string) *TaxScan
}

// DriverTax is a function for
// open a taxonomy database.
type DriverTax func(string) (Taxonomy, error)

var (
	taxDriversMu sync.RWMutex
	taxDrivers   = make(map[string]DriverTax)
)

// RegisterTax makes a taxonomy driver
// available by the provided name.
// If Register is called twice with the same name
// or if drives is nil,
// it panics.
func RegisterTax(name string, driver DriverTax) {
	taxDriversMu.Lock()
	defer taxDriversMu.Unlock()
	if driver == nil {
		panic("biodv: Taxonomy drivers is nil")
	}
	if _, dup := taxDrivers[name]; dup {
		panic("biodv: RegisterTax called twice for driver " + name)
	}
	taxDrivers[name] = driver
}

// TaxDrivers returns a sorted list
// of names of the registered drivers.
func TaxDrivers() []string {
	taxDriversMu.RLock()
	defer taxDriversMu.RUnlock()
	var ls []string
	for name := range taxDrivers {
		ls = append(ls, name)
	}
	sort.Strings(ls)
	return ls
}

// OpenTax opens a taxonomy database
// by its driver,
// and a driver specific parameter string.
func OpenTax(driver, param string) (Taxonomy, error) {
	taxDriversMu.RLock()
	fn, ok := taxDrivers[driver]
	taxDriversMu.RUnlock()
	if !ok {
		return nil, errors.Errorf("biodv: unknown taxonomy driver %q", driver)
	}
	return fn(param)
}

// A Taxon is a taoxn name in a taxonomy.
type Taxon interface {
	// Name returns the canonical name of the current taxon.
	Name() string

	// ID returns the ID of the current taxon.
	ID() string

	// Parent returns the ID of the taxon's parent.
	Parent() string

	// Rank returns the linnean rank of the current taxon.
	Rank() Rank

	// IsCorrect returns true if the taxon
	// is a correct name
	// (i.e. not a synonym).
	IsCorrect() bool

	// Keys returns a list of additional fields
	// stored in the taxon.
	Keys() []string

	// Value returns the value
	// of an additional field stored in the taxon.
	Value(key string) string
}

// Common keys used for a Taxon.
const (
	TaxAuthor = "author"
	TaxRef    = "reference"
	TaxSource = "source"
)

// A TaxScan is a taxon scanner
// to stream the results of a quary
// that are expected to produce
// a taxon list.
//
// Use Scan to advance the stram:
//
//	sc, err := txm.Taxon("Rhinella")
//	for sc.Scan() {
//		tax := sc.Taxon()
//		...
//	}
//	if err := sc.Err(); err != nil {
//		...	// process the error
//	}
type TaxScan struct {
	// the taxon channel
	c chan Taxon

	// an error encountered during iteration
	err error

	// closed if true is the scanner is closed.
	closed bool

	// tax is the last read taxon.
	tax Taxon
}

// NewTaxScan creates a taxon scanner,
// with a buffer of the indicated size.
func NewTaxScan(sz int) *TaxScan {
	if sz < 10 {
		sz = 10
	}
	return &TaxScan{c: make(chan Taxon, sz)}
}

// Add adds a taxon or an error
// to a taxon scanner.
// It should be used by clients that
// return the scanner.
//
// It returns true,
// if the element is added succesfully.
func (tsc *TaxScan) Add(tax Taxon, err error) bool {
	if tsc.err != nil {
		return false
	}
	if tsc.closed {
		return false
	}
	if err != nil {
		close(tsc.c)
		tsc.err = err
		return true
	}
	tsc.c <- tax
	return true
}

// Close closes the scanner.
// If Scan is called and returns false
// the scanner is closed automatically.
func (tsc *TaxScan) Close() {
	if tsc.closed {
		return
	}
	if tsc.err != nil {
		return
	}
	close(tsc.c)
	tsc.closed = true
}

// Err returns the errors,
// if any,
// that was encountered during iteration.
func (tsc *TaxScan) Err() error {
	if !tsc.closed {
		return nil
	}
	if errors.Cause(tsc.err) == io.EOF {
		return nil
	}
	return tsc.err
}

// Scan advances the scanner to the next result.
// It returns false when there is no more taxons,
// or an error happens when preparing it.
// Err should be consulted to distinguish
// between the two cases.
//
// Every call to Taxon,
// even the first one,
// must be precede by a call to Scan.
func (tsc *TaxScan) Scan() bool {
	if tsc.closed {
		return false
	}
	tsc.tax = <-tsc.c
	if tsc.tax == nil {
		tsc.closed = true
		if tsc.err == nil {
			tsc.err = io.EOF
			close(tsc.c)
		}
		return false
	}
	return true
}

// Taxon returns the last read taxon.
// Every call to Taxon must be preceded
// by a call to Scan.
func (tsc *TaxScan) Taxon() Taxon {
	if tsc.closed {
		panic("biodv: accessing a closed taxon scanner")
	}
	tax := tsc.tax
	tsc.tax = nil
	if tax == nil {
		panic("biodv: calling taxon withon an Scan call")
	}
	return tax
}

// TaxList creates a list of Taxon,
// from a taxon scanner.
func TaxList(tsc *TaxScan) ([]Taxon, error) {
	var ls []Taxon
	for tsc.Scan() {
		tax := tsc.Taxon()
		ls = append(ls, tax)
	}
	if err := tsc.Err(); err != nil {
		return nil, err
	}
	return ls, nil
}

// Rank is a linnean rank.
// Ranks are arranged in a way that
// an inclusive rank in the taxonomy
// is always smaller
// than more exclusive ranks.
//
// Then it is possible to use the form:
//
//	if rank < biodv.Genus {
//		// do something
//	}
type Rank uint

// Valid taxonomic ranks.
const (
	Unranked Rank = iota
	Kingdom
	Phylum
	Class
	Order
	Family
	Genus
	Species
)

// ranks holds a list of the accepted rank names.
var ranks = []string{
	Unranked: "unranked",
	Kingdom:  "kingdom",
	Phylum:   "phylum",
	Class:    "class",
	Order:    "order",
	Family:   "family",
	Genus:    "genus",
	Species:  "species",
}

// GetRank returns a rank value from a string.
func GetRank(s string) Rank {
	s = strings.ToLower(s)
	for i, r := range ranks {
		if r == s {
			return Rank(i)
		}
	}
	return Unranked
}

// String returns the rank string of a rank.
func (r Rank) String() string {
	i := int(r)
	if i >= len(ranks) {
		return ranks[Unranked]
	}
	return ranks[i]
}
