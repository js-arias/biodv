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

var testData = []struct {
	name    string
	parent  string
	correct bool
	rank    biodv.Rank
}{
	{"Hominidae", "", true, biodv.Family},
	{"Pongo", "Hominidae", true, biodv.Genus},
	{"Pan", "Hominidae", true, biodv.Genus},
	{"Pan troglodytes", "Pan", true, biodv.Species},
	{"Pan paniscus", "Pan", true, biodv.Species},
	{"Homo", "Hominidae", true, biodv.Genus},
	{"Homo sapiens", "Homo", true, biodv.Species},
	{"Pithecanthropus", "Homo", false, biodv.Genus},
}

func TestAdd(t *testing.T) {
	db := &DB{ids: make(map[string]*Taxon)}

	if _, err := db.Add(" ", "Primates", biodv.Class, true); err == nil {
		t.Errorf("adding an empty taxon, expecting error")
	}
	if _, err := db.Add("Tarsidae", "Primates", biodv.Family, true); err == nil {
		t.Errorf("adding a taxon with no parent on DB, expecting error")
	}
	for _, d := range testData {
		tax, err := db.Add(d.name, d.parent, d.rank, d.correct)
		if err != nil {
			t.Errorf("when adding %s: %v", d.name, err)
			continue
		}
		if tax == nil {
			t.Errorf("when adding %s: expecting taxon value", d.name)
		}
		if tax.Name() != d.name {
			t.Errorf("name %q, want %q", tax.Name(), d.name)
		}
		if tax.Parent() != d.parent {
			t.Errorf("taxon %s: parent %q, want %q", tax.Name(), tax.Parent(), d.parent)
		}
		if tax.Rank() != d.rank {
			t.Errorf("taxon %s: rank %s, want %s", tax.Name(), tax.Rank(), d.rank)
		}
		if tax.IsCorrect() != d.correct {
			t.Errorf("taxon %s: is correct %v", tax.Name(), tax.IsCorrect())
		}
	}
	if _, err := db.Add("Pithecanthropus erectus", "Pithecanthropus", biodv.Species, false); err == nil {
		t.Errorf("adding a synonym to a synonym, expecting error")
	}
	if _, err := db.Add("Gorilla", "Pan", biodv.Genus, true); err == nil {
		t.Errorf("adding a taxon with equal rank, expecting error")
	}
	for _, d := range testData {
		if _, err := db.Add(d.name, d.parent, d.rank, d.correct); err == nil {
			t.Errorf("adding %s, a duplicate, expecting error", d.name)
		}
	}
	if _, err := db.Add("Rhedosaurus", "", biodv.Genus, false); err == nil {
		t.Errorf("adding a synonym without a parent, expecting error")
	}

	ls, err := biodv.TaxList(db.Children(""))
	if err != nil {
		t.Errorf("when looking for root-childrens: %v", err)
	}
	if len(ls) != 1 {
		t.Errorf("number of root-children %d, want %d", len(ls), 1)
	}

	ls, err = biodv.TaxList(db.Children("Hominidae"))
	if err != nil {
		t.Errorf("when looking for \"Hominidae\" childrens: %v", err)
	}
	if len(ls) != 3 {
		t.Errorf("number of \"Hominidae\" children %d, want %d", len(ls), 3)
	}

	ls, err = biodv.TaxList(db.Taxon("Pan paniscus"))
	if err != nil {
		t.Errorf("when looking for \"Pan paniscus\": %v", err)
	}
	if len(ls) != 1 {
		t.Errorf("number of taxons with name \"Pan paniscus\" %d, want %d", len(ls), 1)
	}
	tax := ls[0]
	if !tax.IsCorrect() || tax.Parent() != "Pan" {
		t.Errorf("taxon %q with wrong data: %v parent %s", tax.Name(), tax.IsCorrect(), tax.Parent())
	}

	tax, err = db.TaxID("Pithecanthropus")
	if err != nil {
		t.Errorf("when looking for \"Pithecanthropus\": %v", err)
	}
	if tax.IsCorrect() || tax.Parent() != "Homo" {
		t.Errorf("taxon %q with wrong data: %v parent %s", tax.Name(), tax.IsCorrect(), tax.Parent())
	}

	if !db.changed {
		t.Errorf("database has changed, but no change recorded")
	}
}

func TestMove(t *testing.T) {
	db := &DB{ids: make(map[string]*Taxon)}
	sc := NewScanner(strings.NewReader(scannerBlob))
	if err := db.scan(sc); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	tax := db.TaxEd("Pan")
	if err := tax.Move("Homo", false); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if tax.IsCorrect() {
		t.Errorf("after movement \"Pan\" should be a synonym: %v", tax.data[correctKey])
	}
	if tax.Parent() != "Homo" {
		t.Errorf("taxon %q with wrong parent: %q, want %q", tax.Name(), tax.Parent(), "Homo")
	}
	ls, err := biodv.TaxList(db.Children("Homo"))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(ls) != 3 {
		t.Errorf("\"Homo\" children found %d records, want %d", len(ls), 3)
	}

	tax = db.TaxEd("Pithecanthropus")
	if err := tax.Move("Hominidae", true); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !tax.IsCorrect() {
		t.Errorf("after movement \"Pithecanthropus\" should be a correct")
	}
	if tax.Parent() == "Homo" {
		t.Errorf("taxon %q with wrong parent: %q, want %q", tax.Name(), tax.Parent(), "Hominidae")
	}
	ls, err = biodv.TaxList(db.Synonyms("Homo"))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(ls) != 1 {
		t.Errorf("\"Homo\" synonyms found %d records, want %d", len(ls), 1)
	}
}

func TestSetRank(t *testing.T) {
	db := &DB{ids: make(map[string]*Taxon)}
	sc := NewScanner(strings.NewReader(scannerBlob))
	if err := db.scan(sc); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	tax := db.TaxEd("Hominidae")
	if err := tax.SetRank(biodv.Class); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if tax.Rank() != biodv.Class {
		t.Errorf("taxon %q unnexpected rank: %v, want: %v", tax.Name(), tax.Rank(), biodv.Class)
	}

	tax = db.TaxEd("Pan")
	if err := tax.SetRank(biodv.Species); err == nil {
		t.Errorf("wanted an error: inconsistent rank on children")
	}
	if tax.Rank() != biodv.Genus {
		t.Errorf("taxon %q unnexpected rank: %v, want: %v", tax.Name(), tax.Rank(), biodv.Genus)
	}
}

func TestSet(t *testing.T) {
	db := &DB{ids: make(map[string]*Taxon)}
	sc := NewScanner(strings.NewReader(scannerBlob))
	if err := db.scan(sc); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	db.changed = false

	tax := db.TaxEd("Hominidae")
	if err := tax.Set(biodv.TaxExtern, "    "); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if db.changed {
		t.Errorf("database should be unchanged")
	}
}

func TestDelete(t *testing.T) {
	db := &DB{ids: make(map[string]*Taxon)}
	sc := NewScanner(strings.NewReader(scannerBlob))
	if err := db.scan(sc); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	db.changed = false
	tax := db.TaxEd("Homo")
	tax.Delete(false)

	tx := db.TaxEd("Homo sapiens")
	if tx == nil {
		t.Errorf("\"Homo sapiens\" should be present")
	}
	if tx.Parent() != "Hominidae" {
		t.Errorf("bad parent %q, want %q", tx.Parent(), "Homonidae")
	}

	tx = db.TaxEd("Pithecanthropus")
	if tx == nil {
		t.Errorf("\"Pithecanthropus\" should be present")
	}
	if tx.Parent() != "Hominidae" {
		t.Errorf("bad parent %q, want %q", tx.Parent(), "Homonidae")
	}

	tax = db.TaxEd("Pan")
	tax.Delete(true)
	if tx := db.TaxEd("Pan paniscus"); tx != nil {
		t.Errorf("\"Pan paniscus\" is on database")
	}
}

var orderBlob = `
name:	Mustela
rank:	genus
correct: true
author:	Linneaeus, 1758
%%
name:	Cabreragale
rank:	genus
parent:	Mustela
correct: false
author: Baryshnikov & Abramov, 1997
%%
name:	Cyanomyonax
rank:	genus
parent:	Mustela
correct: false
author: Trouessart, 1885
%%
name:	Kolonocus
rank:	genus
parent:	Mustela
correct: false
author:	Satunin, 1914
%%
name:	Mustela nigripes
rank:	species
parent: Mustela
correct: true
author: (Audubon & Bachman, 1851)
%%
name:	Mustela frenata
rank:	species
parent: Mustela
correct: true
author: Lichtenstein, 1831
%%
name:	Mustela nivalis
rank:	species
parent: Mustela
correct: true
author: Linnaeus, 1766
%%
name:	Mustela nivalis corsicana
rank:	unranked
parent:	Mustela nivalis
correct: false
%%
name:	Mustela rixosa
rank:	species
parent:	Mustela nivalis
correct: false
author: (Bangs, 1896)
%%
name:	Mustela vulgaris
rank:	species
parent:	Mustela nivalis
correct: false
author: Erxleben, 1777
%%
name:   Canis
rank:   genus
correct: true
author: Linnaeus, 1758
%%
name:   Canis latrans
parent: Canis
rank:   species
correct: true
author: Say, 1823
%%
name:   Canis latrans latrans
parent: Canis latrans
rank:   unranked
correct: false
author: Say, 1823
%%
name:   Canis ochropus
parent: Canis latrans
rank:   species
correct: false
author: Eschscholtz, 1829
%%
name:   Canis latrans ochropus
parent: Canis latrans
rank:   unranked
correct: false
author: Eschscholtz, 1829
%%
name:   Lyciscus cagottis
parent: Canis latrans
rank:   species
correct: false
author: Hamilton-Smith, 1839
%%
name:   Canis latrans cagottis
parent: Canis latrans
rank:   unranked
correct: false
author: C.E.H.Smith, 1839
%%
name:   Canis latrans frustror
parent: Canis latrans
rank:   unranked
correct: false
author: Woodhouse, 1851
%%
name:   Canis frustror
parent: Canis latrans
rank:   species
correct: false
author: Woodhouse, 1851
%%
`

func TestChildrenOrder(t *testing.T) {
	db := &DB{ids: make(map[string]*Taxon)}
	sc := NewScanner(strings.NewReader(orderBlob))
	if err := db.scan(sc); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// sorted childrens
	ch := []string{
		// correct-valid names first,
		// sorted by name
		"Mustela frenata",
		"Mustela nigripes",
		"Mustela nivalis",

		// synonyms sorted by date
		"Cyanomyonax", // 1885
		"Kolonocus",   // 1914
		"Cabreragale", // 1997
	}

	ls := db.TaxList("Mustela")
	for i, tax := range ls {
		if tax.Name() != ch[i] {
			t.Errorf("taxon %q, want %q", tax.Name(), ch[i])
		}
	}

	// sorted synonyms
	sn := []string{
		"Mustela vulgaris",          // 1777
		"Mustela rixosa",            // 1896
		"Mustela nivalis corsicana", // no-year
	}

	ls = db.TaxList("Mustela nivalis")
	for i, tax := range ls {
		if tax.Name() != sn[i] {
			t.Errorf("taxon %q, want %q", tax.Name(), sn[i])
		}
	}

	sn = []string{
		"Canis latrans latrans",  // 1823
		"Canis latrans ochropus", // 1829
		"Canis ochropus",         //1829
		"Canis latrans cagottis", // 1839
		"Lyciscus cagottis",      // 1839
		"Canis frustror",         // 1851
		"Canis latrans frustror", // 1851
	}
	ls = db.TaxList("Canis latrans")
	for i, tax := range ls {
		if tax.Name() != sn[i] {
			t.Errorf("taxon %q, want %q", tax.Name(), sn[i])
		}
	}
}
