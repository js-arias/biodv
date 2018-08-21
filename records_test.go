// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package biodv

import (
	"testing"

	"github.com/pkg/errors"
)

type mockRecord string

func (mr mockRecord) ID() string {
	return string(mr)
}

func (mr mockRecord) Taxon() string              { return "mock" }
func (mr mockRecord) Basis() BasisOfRecord       { return UnknownBasis }
func (mr mockRecord) CollEvent() CollectionEvent { return CollectionEvent{} }
func (mr mockRecord) GeoRef() Point              { return InvalidPoint() }
func (mr mockRecord) Keys() []string             { return nil }
func (mr mockRecord) Value(key string) string    { return "" }

var mockRecList = []string{
	"gbif:1494120332",
	"gbif:1494120360",
	"gbif:1494120372",
	"gbif:1494120430",
	"gbif:1494120449",
	"gbif:1494120458",
	"gbif:1494120480",
}

func TestRecScan(t *testing.T) {

	// Expected RecScan usage
	sc := NewRecScan(10)
	go func(x *RecScan) {
		for _, s := range mockRecList {
			if !x.Add(mockRecord(s), nil) {
				break
			}
		}
		x.Add(nil, nil)
	}(sc)

	c := 0
	for sc.Scan() {
		rec := sc.Record()
		if rec.ID() != mockRecList[c] {
			t.Errorf("record ID %q, want %q", rec.ID(), mockRecList[c])
		}
		c++
	}
	if err := sc.Err(); err != nil {
		t.Errorf("recscan unexpected error: %v", err)
	}
	if len(mockRecList) != c {
		t.Errorf("scanned records %d, want %d", c, len(mockRecList))
	}

	// Closing RecScan before finish
	sc = NewRecScan(10)
	go func(x *RecScan) {
		for _, s := range mockRecList {
			if !x.Add(mockRecord(s), nil) {
				break
			}
		}
		x.Add(nil, nil)
	}(sc)

	c = 0
	for sc.Scan() {
		rec := sc.Record()
		if rec.ID() != mockRecList[c] {
			t.Errorf("record ID %q, want %q", rec.ID(), mockRecList[c])
		}
		c++
		if rec.ID() == "gbif:1494120372" {
			sc.Close()
		}
	}
	if err := sc.Err(); err != nil {
		t.Errorf("recscan unexpected error: %v", err)
	}
	if c > 3 {
		t.Errorf("scanned records %d, want %d", c, 3)
	}

	// An error received during iteration
	sc = NewRecScan(10)
	go func(x *RecScan) {
		for _, s := range mockRecList {
			if s == "gbif:1494120430" {
				x.Add(nil, errors.New("mock error"))
				return
			}
			if !x.Add(mockRecord(s), nil) {
				break
			}
		}
		x.Add(nil, nil)
	}(sc)
	c = 0
	for sc.Scan() {
		rec := sc.Record()
		if rec.ID() != mockRecList[c] {
			t.Errorf("record ID %q, want %q", rec.ID(), mockRecList[c])
		}
		c++
	}
	if err := sc.Err(); err == nil {
		t.Errorf("recscan expecting error")
	}
	if c > 3 {
		t.Errorf("scanned records %d, want %d", c, 3)
	}
}
