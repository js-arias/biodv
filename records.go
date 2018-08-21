// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package biodv

import (
	"io"
	"strings"
	"time"

	"github.com/pkg/errors"
)

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

// A CollectionEvent stores the information
// of a collection event for a record.
type CollectionEvent struct {
	Date      time.Time
	Country   string
	State     string
	County    string
	Locality  string
	Collector string
}

// A Point is a georeferencenced point record.
type Point struct {
	Lon         float64
	Lat         float64
	Altitude    float64
	Depth       float64
	Source      string // source of the reference
	Uncertainty uint   // georeference uncertainty in meters
	Validation  string // source of the validation
}

// Maximum and minimum values for geographic coordinates
const (
	MinLon = -180
	MaxLon = 180
	MinLat = -90
	MaxLat = 90
)

// InvalidPoint returns a new Point without a valid georeference.
func InvalidPoint() Point {
	return Point{Lon: 360, Lat: 180}
}

// IsValid returns true if a geographic point is valid.
func (p Point) IsValid() bool {
	if (p.Lon <= MaxLon) && (p.Lon > MinLon) {
		if (p.Lat < MaxLat) && (p.Lat > MinLat) {
			return true
		}
	}
	return false
}
