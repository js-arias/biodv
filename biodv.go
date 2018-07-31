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
	"sort"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

// A Taxonomy is a taxonomy database.
type Taxonomy interface {
	// Taxon returns a list of taxons with a given name.
	Taxon(name string) ([]Taxon, error)

	// TaxID returns the taxon with a given ID.
	TaxID(id string) (Taxon, error)
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
