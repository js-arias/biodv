// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package dataset

import (
	"strings"
	"testing"

	"github.com/js-arias/biodv"
)

var scannerBlob = `
title:	GBIF Backbone Taxonomy
about:	The GBIF Backbone Taxonomy, often called the Nub taxonomy.
reference: GBIF Secretariat (2017). GBIF Backbone Taxonomy. Checklist dataset https://doi.org/10.15468/39omei accessed via GBIF.org on 2018-09-19.
license: CC BY 4.0
url:	https://www.gbif.org/dataset/d7dddbf4-2cf0-4f39-9b2a-bb099caae36c
extern:	gbif:d7dddbf4-2cf0-4f39-9b2a-bb099caae36c
%%
title:	NMNH Extant Specimen Records
about:	Public records of accessioned specimens and observations curated by the National Museum of Natural History, Smithsonian Institution.
license: CC0-1.0
url:	http://collections.nmnh.si.edu
publisher: National Museum of Natural History, Smithsonian Institution
extern: gbif:821cc27a-e3bb-4bc5-ac34-89ada245069d
%%
title:	Geographically tagged INSDC sequences
about:	Metadata for INSDC sequences that have been geographically tagged.
reference: European Molecular Biology Laboratory (EMBL) (2014). Geographically tagged INSDC sequences. Occurrence dataset https://doi.org/10.15468/cndomv accessed via GBIF.org on 2018-09-19.
license: CC BY 4.0
url:	http://www.ebi.ac.uk
publisher: European Bioinformatics Institute (EMBL-EBI)
extern: gbif:ad43e954-dd79-4986-ae34-9ccdbd8bf568
%%
`

func TestScan(t *testing.T) {
	sc := NewScanner(strings.NewReader(scannerBlob))
	i := 0
	for sc.Scan() {
		set := sc.Dataset()
		if set.Title() != testData[i].title {
			t.Errorf("wrong title %q, want %q", set.Title(), testData[i].title)
		}
		if set.Value(biodv.SetExtern) == "" {
			t.Errorf("dataset %q does not have \"extern\" key", set.Title())
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
	db := &DB{ids: make(map[string]*Dataset)}

	sc := NewScanner(strings.NewReader(scannerBlob))
	if err := db.scan(sc); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	set, _ := db.SetID("gbif:821cc27a-e3bb-4bc5-ac34-89ada245069d")
	if set == nil {
		t.Errorf("no dataset assocaited with key: \"gbif:821cc27a-e3bb-4bc5-ac34-89ada245069d\"")
	}
	set, _ = db.SetID("GBIF Backbone Taxonomy")
	if v := set.Value(biodv.SetExtern); v != "gbif:d7dddbf4-2cf0-4f39-9b2a-bb099caae36c" {
		t.Errorf("set %q extern %q, want %q", set.Title(), v, "gbif:d7dddbf4-2cf0-4f39-9b2a-bb099caae36c")
	}
}
