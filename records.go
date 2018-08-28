// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package biodv

import (
	"io"
	"sort"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

// RecDriver contains components
// of a RecDB driver.
type RecDriver struct {
	// Open is a function to open
	// a RecDB.
	Open func(string) (RecDB, error)

	// URL is a function to return
	// an URL of a given record ID.
	// This value can be nil.
	URL func(id string) string

	// About is a function that return
	// a short description of the driver.
	About func() string
}

var (
	recDriversMu sync.RWMutex
	recDrivers   = make(map[string]RecDriver)
)

// RegisterRec makes a recDB driver
// availablre by the provided name.
// If Register is called twice with the same name
// or if driver is nil,
// it panics.
func RegisterRec(name string, driver RecDriver) {
	recDriversMu.Lock()
	defer recDriversMu.Unlock()
	if driver.Open == nil {
		panic("biodv: RecDB driver Open is nil")
	}
	if _, dup := recDrivers[name]; dup {
		panic("biodv: RegisterRec called twice for driver " + name)
	}
	recDrivers[name] = driver
}

// RecDrivers returns a sorted list
// of names of the registered drivers.
func RecDrivers() []string {
	recDriversMu.RLock()
	defer recDriversMu.RUnlock()
	var ls []string
	for name := range recDrivers {
		ls = append(ls, name)
	}
	sort.Strings(ls)
	return ls
}

// OpenRec opens a RecDB database
// by its driver,
// and a driver specific parameter string.
func OpenRec(driver, param string) (RecDB, error) {
	if driver == "" {
		return nil, errors.New("biodv: empty recDB driver")
	}
	recDriversMu.RLock()
	dr, ok := recDrivers[driver]
	recDriversMu.RUnlock()
	if !ok {
		return nil, errors.Errorf("biodv: unknonw recDB driver %q", driver)
	}
	return dr.Open(param)
}

// RecURL returns the URL of a given Record ID
// in a given database.
func RecURL(driver, id string) string {
	if driver == "" {
		return ""
	}
	recDriversMu.RLock()
	dr, ok := recDrivers[driver]
	recDriversMu.RUnlock()
	if !ok {
		return ""
	}
	if dr.URL == nil {
		return ""
	}
	return dr.URL(id)
}

// RecAbout returns the short message
// describing the driver.
func RecAbout(driver string) string {
	if driver == "" {
		return ""
	}
	recDriversMu.RLock()
	dr, ok := recDrivers[driver]
	recDriversMu.RUnlock()
	if !ok {
		return ""
	}
	if dr.About == nil {
		return ""
	}
	return dr.About()
}

// A RecScan is a record scanner
// to stream the results of a query
// that is expected to producece
// a list of records.
//
// Use scan to advance the stream:
//
//	sc := recs.TaxRecs("Rhinella")
//	for sc.Scan() {
//		rec := sc.Record()
//		...
//	}
//	if err := sc.Err(); err != nil {
//		...	// process the error
//	}
type RecScan struct {
	// the record channel
	c chan Record

	// an error encountered during iteration
	err error

	// closed if true is the scanner is closed
	closed bool

	// rec is the last read record
	rec Record
}

// NewRecScan creates a record scanner,
// with a buffer of the indicated size.
func NewRecScan(sz int) *RecScan {
	if sz < 10 {
		sz = 10
	}
	return &RecScan{c: make(chan Record, sz)}
}

// Add adds a record or an error
// to a record scanner.
// It should be used by clients that
// return the scanner.
//
// It returns true,
// if the element is added successfully.
func (rsc *RecScan) Add(rec Record, err error) bool {
	if rsc.err != nil {
		return false
	}
	if rsc.closed {
		return false
	}
	if err != nil {
		close(rsc.c)
		rsc.err = err
		return true
	}
	rsc.c <- rec
	return true
}

// Close closes the scanner.
// If Scan is called and returns false
// the scanner is closed automatically.
func (rsc *RecScan) Close() {
	if rsc.closed {
		return
	}
	if rsc.err != nil {
		return
	}
	close(rsc.c)
	rsc.closed = true
}

// Err returns the error,
// if any,
// that was encountered during iteration.
func (rsc *RecScan) Err() error {
	if !rsc.closed {
		return nil
	}
	if errors.Cause(rsc.err) == io.EOF {
		return nil
	}
	return rsc.err
}

// Record returns the last read record.
// Every call to Record must be preceded
// by a call to Scan.
func (rsc *RecScan) Record() Record {
	if rsc.closed {
		panic("biodv: accessing a closed record scanner")
	}
	rec := rsc.rec
	rsc.rec = nil
	if rec == nil {
		panic("biodv: calling Record without an Scan call")
	}
	return rec
}

// Scan advances the scanner to the next result.
// It returns false when there is no more records,
// or an error happens when preparing it.
// Err should be consulted to distinguish
// between the two cases.
//
// Every call to Record,
// even the first one,
// must be precede by a call to Scan.
func (rsc *RecScan) Scan() bool {
	if rsc.closed {
		return false
	}
	rsc.rec = <-rsc.c
	if rsc.rec == nil {
		rsc.closed = true
		if rsc.err == nil {
			rsc.err = io.EOF
			close(rsc.c)
		}
		return false
	}
	return true
}

// BasisOfRecord indicates the physic basis
// of an specimen record.
type BasisOfRecord uint

// Valid basis of record.
const (
	UnknownBasis BasisOfRecord = iota
	Preserved                  // a preserved (museum) specimen
	Fossil                     // a fossilized specimen
	Observation                // a human observation
	Machine                    // a machine "observation"
)

// basis holds a list of the accepted basis of record names.
var basis = []string{
	"unknown",
	"preserved",
	"fossil",
	"observation",
	"machine",
}

// GetBasis returns a BasisOfRecord value from a string.
func GetBasis(s string) BasisOfRecord {
	s = strings.ToLower(s)
	for i, b := range basis {
		if b == s {
			return BasisOfRecord(i)
		}
	}
	return UnknownBasis
}

// String returns the basis string of a basis of record.
func (b BasisOfRecord) String() string {
	i := int(b)
	if i >= len(basis) {
		return basis[UnknownBasis]
	}
	return basis[i]
}
