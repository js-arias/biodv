// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package taxonomy

import (
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
