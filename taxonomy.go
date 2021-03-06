// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package biodv

import (
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/pkg/errors"
)

// TaxDriver contains components
// of a Taxonomy driver.
type TaxDriver struct {
	// Open is a function to open
	// a taxonomy.
	Open func(string) (Taxonomy, error)

	// URL is a function to return
	// an URL of a given taxon ID.
	// This value can be nil.
	URL func(id string) string

	// About is a function that return
	// a short description of the driver.
	About func() string
}

var (
	taxDriversMu sync.RWMutex
	taxDrivers   = make(map[string]TaxDriver)
)

// RegisterTax makes a taxonomy driver
// available by the provided name.
// If Register is called twice with the same name
// or if driver is nil,
// it panics.
func RegisterTax(name string, driver TaxDriver) {
	taxDriversMu.Lock()
	defer taxDriversMu.Unlock()
	if driver.Open == nil {
		panic("biodv: Taxonomy driver Open is nil")
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
	if driver == "" {
		return nil, errors.New("biodv: empty taxonomy driver")
	}
	taxDriversMu.RLock()
	dr, ok := taxDrivers[driver]
	taxDriversMu.RUnlock()
	if !ok {
		return nil, errors.Errorf("biodv: unknown taxonomy driver %q", driver)
	}
	return dr.Open(param)
}

// TaxURL returns the URL of a given taxon ID
// in a given database.
func TaxURL(driver, id string) string {
	if driver == "" {
		return ""
	}
	taxDriversMu.RLock()
	dr, ok := taxDrivers[driver]
	taxDriversMu.RUnlock()
	if !ok {
		return ""
	}
	if dr.URL == nil {
		return ""
	}
	return dr.URL(id)
}

// TaxAbout returns the short message
// describing the driver.
func TaxAbout(driver string) string {
	if driver == "" {
		return ""
	}
	taxDriversMu.RLock()
	dr, ok := taxDrivers[driver]
	taxDriversMu.RUnlock()
	if !ok {
		return ""
	}
	if dr.About == nil {
		return ""
	}
	return dr.About()
}

// A TaxScan is a taxon scanner
// to stream the results of a query
// that is expected to produce
// a taxon list.
//
// Use Scan to advance the stream:
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

	// closed if true is the scanner is closed
	closed bool

	// tax is the last read taxon
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
// if the element is added successfully.
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

// Err returns the error,
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
		panic("biodv: calling Taxon withon an Scan call")
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

// TaxParents returns a list with the parents
// in a Taxonomy
// for a given taxon ID
// (included the given taxon)
// sorted from the most inclusive
// to the most exclusive taxon.
func TaxParents(txm Taxonomy, id string) ([]Taxon, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, nil
	}
	tax, err := txm.TaxID(id)
	if err != nil {
		return nil, err
	}
	pls, err := TaxParents(txm, tax.Parent())
	if err != nil {
		return nil, err
	}
	pls = append(pls, tax)
	return pls, nil
}

// TaxCanon returns a taxon name into its canonical form.
func TaxCanon(name string) string {
	name = strings.Join(strings.Fields(name), " ")
	if name == "" {
		return ""
	}
	name = strings.ToLower(name)
	r, n := utf8.DecodeRuneInString(name)
	return string(unicode.ToTitle(r)) + name[n:]
}

// TaxYear returns the year of the taxon description
// as given from the author field.
// If no year is defined,
// it will return 0.
func TaxYear(tx Taxon) int {
	if tx == nil {
		return 0
	}
	return getYearFromAuthor(tx.Value(TaxAuthor))
}

func getYearFromAuthor(author string) int {
	if author == "" {
		return 0
	}
	author = strings.TrimRight(author, ")")
	if len(author) < 4 {
		return 0
	}
	year, err := strconv.Atoi(author[len(author)-4:])
	if err != nil {
		return 0
	}
	if year < 1750 || year > time.Now().Year() {
		return 0
	}
	return year
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
