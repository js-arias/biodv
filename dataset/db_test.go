// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package dataset

import "testing"

var testData = []struct {
	title   string
	license string
	pub     string
}{
	{"GBIF Backbone Taxonomy", "CC BY 4.0", ""},
	{"NMNH Extant Specimen Records", "CC0-1.0", "National Museum of Natural History, Smithsonian Institution"},
	{"Geographically tagged INSDC sequences", "CC BY 4.0", "European Bioinformatics Institute (EMBL-EBI)"},
}

func TestAdd(t *testing.T) {
	db := &DB{ids: make(map[string]*Dataset)}

	if _, err := db.Add("    "); err == nil {
		t.Errorf("adding an empty dataset, expecting error")
	}

	for _, d := range testData {
		set, err := db.Add(d.title)
		if err != nil {
			t.Errorf("when adding %q: %v", d.title, err)
		}
		if set == nil {
			t.Errorf("when adding %q: expecting a dataset.Dataset", d.title)
		}
		if set.ID() != d.title {
			t.Errorf("ID %q, want %q", set.ID(), d.title)
		}
		if set.Title() != d.title {
			t.Errorf("title %q, want %q", set.ID(), d.title)
		}
	}

	for _, d := range testData {
		set, err := db.SetID(d.title)
		if err != nil {
			t.Errorf("when looking for %q: %v", d.title, err)
		}
		if set == nil {
			t.Errorf("when looking for %q: no dataset retrieved", d.title)
		}
		if set.Title() != d.title {
			t.Errorf("dataset %q, want %q", set.Title(), d.title)
		}
	}

	for _, d := range testData {
		if _, err := db.Add(d.title); err == nil {
			t.Errorf("adding %q, a repeated dataset", d.title)
		}
	}
}
