// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package records

import (
	"math"
	"strings"
	"testing"

	"github.com/js-arias/biodv"
)

var testData = []struct {
	taxon  string
	basis  biodv.BasisOfRecord
	id     string
	lat    float64
	lon    float64
	extern string
}{
	{"Larus argentatus", biodv.Machine, "Larus argentatus:1", 50.223982, 1.596802, "gbif:1494057472"},
	{"Felis concolor couguar", biodv.Preserved, "Felis concolor couguar:1", 360, 360, "gbif:1893501987"},
	{"Felis concolor", biodv.Preserved, "MSU:MR:MR.8672", 0.25, -79.8333, "gbif:919431660"},
}

func TestTaxFileName(t *testing.T) {
	names := []struct {
		taxon string
		file  string
	}{
		{"Larus argentatus", "Larus-argentatus.stz"},
		{"felis concolor couguar", "Felis-concolor-couguar.stz"},
		{"FELIS CONCOLOR", "Felis-concolor.stz"},
	}

	for _, p := range names {
		if f := taxFileName(p.taxon); f != p.file {
			t.Errorf("wrong file name %q, want %q", f, p.file)
		}
	}
}

func TestAdd(t *testing.T) {
	db := &DB{tids: make(map[string]*taxon), ids: make(map[string]*Record)}

	if _, err := db.Add("", "", "", biodv.UnknownBasis, 360, 360); err == nil {
		t.Errorf("adding a record without a taxon, expecting error")
	}
	for _, d := range testData {
		rec, err := db.Add(d.taxon, d.id, "", d.basis, d.lat, d.lon)
		if err != nil {
			t.Errorf("when adding %q: %v", d.id, err)
		}
		if rec.ID() != d.id {
			t.Errorf("record %q, want %q", rec.ID(), d.id)
		}
		if rec.Taxon() != d.taxon {
			t.Errorf("record %q, taxon %q, want %q", rec.ID(), rec.Taxon(), d.taxon)
		}
		geo := rec.GeoRef()
		if d.id == "Felis concolor couguar:1" {
			if geo.IsValid() {
				t.Errorf("record %q, valid georef", d.id)
			}
			continue
		}
		if !geo.IsValid() {
			t.Errorf("record %q, invalid georef", d.id)
		}
		if math.Abs(geo.Lat-d.lat) > 0.01 {
			t.Errorf("record %q, lat %f, want %f", d.id, geo.Lat, d.lat)
		}
		if math.Abs(geo.Lon-d.lon) > 0.01 {
			t.Errorf("record %q, lon %f, want %f", d.id, geo.Lon, d.lon)
		}

	}

	for _, d := range testData {
		rec, _ := db.RecID(d.id)
		if rec == nil {
			t.Errorf("record %q, not found", d.id)
		}
	}

	for _, d := range testData {
		if _, err := db.Add(d.taxon, d.id, "", d.basis, d.lat, d.lon); err == nil {
			t.Errorf("when adding %q: expecting error", d.id)
		}
	}
}

var taxBlob = `

Struthio camelus
Rhea americana
Pterocnemia pennata
Casuarius casuarius
Casuarius bennetti
Emuarius guljaruba
Emuarius gidju
Dromaius baudinianus
Dromaius novaehollandiae
Apteryx owenii
Apteryx haastii
Apteryx mantelli
Mullerornis agilis
Aepyornis hildebrandti

Crypturellus tataupa
Tinamus major
Eudromia elegans
Lithornis hookeri
Lithornis celetius
Lithornis vulturinus
Anomalopteryx didiformis
Emeus crassus

# Rhedosaurus
; Indominus rex

Euryapteryx curtus
Pachyornis geranoides
Pachyornis elephantopus
Pachyornis australis
Dinornis robustus
Megalapteryx didinus

`

func TestTaxonList(t *testing.T) {
	db := &DB{tids: make(map[string]*taxon), ids: make(map[string]*Record)}
	r := strings.NewReader(taxBlob)
	if err := db.readTaxList(r); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if _, ok := db.tids["Lithornis hookeri"]; !ok {
		t.Errorf("taxon list unread")
	}
	if _, ok := db.tids[biodv.TaxCanon("# Rhedosaurus")]; ok {
		t.Errorf("lines bigining with '#' should be left unread")
	}
	if _, ok := db.tids[biodv.TaxCanon("; Indominus rex")]; ok {
		t.Errorf("lines bigining with ';' should be left unread")
	}
}
