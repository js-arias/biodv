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

type mockTaxon string

func (mt mockTaxon) Name() string {
	return string(mt)
}

func (mt mockTaxon) ID() string              { return "" }
func (mt mockTaxon) Parent() string          { return "" }
func (mt mockTaxon) Rank() Rank              { return Unranked }
func (mt mockTaxon) IsCorrect() bool         { return true }
func (mt mockTaxon) Keys() []string          { return nil }
func (mt mockTaxon) Value(key string) string { return "" }

var mockTaxList = []string{
	"Homo",
	"Pan",
	"Pongo",
	"Gorilla",
	"Hylobates",
}

func TestTaxScan(t *testing.T) {

	// Expected TaxScan usage
	sc := NewTaxScan(10)
	go func(x *TaxScan) {
		for _, s := range mockTaxList {
			if !x.Add(mockTaxon(s), nil) {
				break
			}
		}
		x.Add(nil, nil)
	}(sc)

	c := 0
	for sc.Scan() {
		tax := sc.Taxon()
		if tax.Name() != mockTaxList[c] {
			t.Errorf("taxon name %q, want %q", tax.Name(), mockTaxList[c])
		}
		c++
	}
	if err := sc.Err(); err != nil {
		t.Errorf("taxscan unexpected error: %v", err)
	}
	if len(mockTaxList) != c {
		t.Errorf("scanned taxons %d, want %d", len(mockTaxList), c)
	}

	// Closing TaxScan before finish
	sc = NewTaxScan(10)
	go func(x *TaxScan) {
		for _, s := range mockTaxList {
			if !x.Add(mockTaxon(s), nil) {
				break
			}
		}
		x.Add(nil, nil)
	}(sc)

	c = 0
	for sc.Scan() {
		tax := sc.Taxon()
		if tax.Name() != mockTaxList[c] {
			t.Errorf("taxon name %q, want %q", tax.Name(), mockTaxList[c])
		}
		c++
		if tax.Name() == "Pongo" {
			sc.Close()
		}
	}
	if err := sc.Err(); err != nil {
		t.Errorf("taxscan unexpected error: %v", err)
	}
	if c > 3 {
		t.Errorf("scanned taxons %d, want %d", c, 3)
	}

	// An error received during iteration
	sc = NewTaxScan(10)
	go func(x *TaxScan) {
		for _, s := range mockTaxList {
			if s == "Pongo" {
				x.Add(nil, errors.New("mock error"))
				return
			}
			if !x.Add(mockTaxon(s), nil) {
				break
			}
		}
		x.Add(nil, nil)
	}(sc)

	c = 0
	for sc.Scan() {
		tax := sc.Taxon()
		if tax.Name() != mockTaxList[c] {
			t.Errorf("taxon name %q, want %q", tax.Name(), mockTaxList[c])
		}
		c++
	}
	if err := sc.Err(); err == nil {
		t.Errorf("taxscan expecting error")
	}
	if c > 2 {
		t.Errorf("scanned taxons %d, want %d", c, 2)
	}
}

func TestBiodvTaxCanon(t *testing.T) {
	testData := []struct {
		text string
		name string
	}{
		{"Hominidae", "Hominidae"},
		{"Pongo   ", "Pongo"},
		{"PAN", "Pan"},
		{"Pan troglodytes", "Pan troglodytes"},
		{"pan paniscus", "Pan paniscus"},
		{"   Homo", "Homo"},
		{"Homo sapiens", "Homo sapiens"},
		{"  Pithecanthropus   ", "Pithecanthropus"},
	}

	for _, d := range testData {
		if cn := TaxCanon(d.text); cn != d.name {
			t.Fatalf("Canon name %q, want %q", cn, d.name)
		}
	}
}

func TestAuthorYear(t *testing.T) {
	testData := []struct {
		text string
		year int
	}{
		{"", 0},
		{"L.", 0},
		{"Spreng.", 0},
		{"Hodgson, 1842", 1842},
		{"(Domaniewski, 1926)", 1926},
		{"invalid-year, 1001", 0},
		{"far-future, 9899", 0},
		{"Trouessart, 1885", 1885},
	}

	for _, d := range testData {
		if y := getYearFromAuthor(d.text); y != d.year {
			t.Errorf("Year %d, want %d", y, d.year)
		}
	}
}
