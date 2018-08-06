// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package taxonomy

import (
	"strings"
	"testing"

	"github.com/js-arias/biodv"
)

var scannerBlob = `
name:	HOMINIDAE
parent:	
rank:	family
correct: true
extern:	gbif:5483
%%
name:	PONGO
parent:	Hominidae
rank:	genus
correct: true
extern:	gbif:5219531
%%
name:	PAN
parent: Hominidae
rank:	genus
correct: true
extern: gbif:2436437
%%
name:	Pan troglodytes
parent: Pan
rank:	species
correct: true
extern:	gbif:5219534
%%
name:	Pan paniscus
parent:	Pan
rank:	species
correct: true
extern: gbif:5219533
%%
name:	HOMO
parent:	Hominidae
rank:	genus
correct: true
extern:	gbif:2436435
%%
name:	Homo sapiens
parent:	HOMO
rank:	species
correct: true
extern:	gbif:2436436
%%
name:	Pithecanthropus
parent:	HOMO
rank:	genus
correct: FALSE
extern: gbif:4827617
%%
`

func TestScan(t *testing.T) {
	sc := NewScanner(strings.NewReader(scannerBlob))
	i := 0
	for sc.Scan() {
		tax := sc.Taxon()
		if tax.Name() != testData[i].name {
			t.Errorf("wrong name %q, want %q", tax.Name(), testData[i].name)
		}
		if tax.Parent() != testData[i].parent {
			t.Errorf("taxon %s, wrong parent %q, want %q", tax.Name(), tax.Parent(), testData[i].parent)
		}
		if tax.Rank() != testData[i].rank {
			t.Errorf("taxon %s, wrong rank %v, want %v", tax.Name(), tax.Rank(), testData[i].rank)
		}
		if tax.IsCorrect() != testData[i].correct {
			t.Errorf("taxon %s is correct == %v", tax.Name(), tax.IsCorrect())
		}
		if tax.Value(biodv.TaxExtern) == "" {
			t.Errorf("taxon %s does not have \"extern\" key", tax.Name())
		}
		i++
	}
	if err := sc.Err(); err != nil {
		t.Errorf("unexpected scanner error: %v", err)
	}
	if i != len(testData) {
		t.Errorf("found %d records, want %d", i, len(testData))
	}
}

func TestDBScan(t *testing.T) {
	db := &DB{ids: make(map[string]*Taxon)}
	sc := NewScanner(strings.NewReader(scannerBlob))
	if err := db.scan(sc); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	tax, _ := db.TaxID("gbif:2436436")
	if tax == nil {
		t.Error("no taxon associated with key: \"gbif:2436436\"")
	}
	tax, _ = db.TaxID("Pan troglodytes")
	if tax == nil {
		t.Error("no taxon associated with key: \"Pan troglodytes\"")
	}
	if v := tax.Value(biodv.TaxExtern); v != "gbif:5219534" {
		t.Errorf("taxon %s extern %q, want %q", tax.Name(), v, "gbif:5219534")
	}
}
