// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package records

import (
	"strings"
	"testing"

	"github.com/js-arias/biodv"
)

var scannerBlob = `
taxon:	Larus argentatus
ID:	Larus argentatus:1
basis:	machine
date:	2016-01-01T11:50:15Z00:00
country: FR
latlon:	50.223982 1.596802
elevation: 2
geosource: gps
dataset: gbif:83e20573-f7dd-4852-9159-21566e1e691e
sex:	female
stage:	adult
organism: H903607
extern: gbif:1494057472
%%
taxon:	Felis concolor couguar
ID:	Felis concolor couguar:1
basis:	preserved
date:	1957-06-30T00:00:00Z00:00
country: MX
state:	Chihuhua
locality: San Francisco
dataset: gbif:2dad0cd2-e880-4ec3-90e5-d3f479528cbd
extern:	gbif:1893501987
%%
taxon:	FELIS CONCOLOR
ID:	MSU:MR:MR.8672
basis:	preserved
dataset: gbif:22a66350-7947-4a49-84a3-39c7c1b0881f
catalog: MSU:MR:MR.8672
date:	1963-08-26T00:00:00Z00:00
coutry:	EC
state:	Manabí
locality: Montañas de Mache
latlon:	0.250000 -79.833300
extern:	gbif:919431660
%%
`

func TestScan(t *testing.T) {
	sc := NewScanner(strings.NewReader(scannerBlob))
	i := 0
	for sc.Scan() {
		rec := sc.Record()
		if rec.Taxon() != biodv.TaxCanon(testData[i].taxon) {
			t.Errorf("wrong taxon %q, want %q", rec.Taxon(), biodv.TaxCanon(testData[i].taxon))
		}
		geo := rec.GeoRef()
		if !geo.IsValid() && i != 1 {
			t.Errorf("invalid georeference for %q [lat:%.3f lon:%.3f], should be valid [%.3f %.3f]", rec.Taxon(), geo.Lat, geo.Lon, testData[i].lat, testData[i].lon)
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
	db := &DB{tids: make(map[string]*taxon), ids: make(map[string]*Record)}
	sc := NewScanner(strings.NewReader(scannerBlob))
	if err := db.scan(sc); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	for _, d := range testData {
		rec, _ := db.RecID(d.id)
		if rec == nil {
			t.Errorf("record %q not found", d.id)
		}
		if v := rec.Value(biodv.RecExtern); v != d.extern {
			t.Errorf("record %q, extern %q, want %q", rec.ID(), v, d.extern)
		}
		if v := rec.Value(biodv.RecDataset); v == "" {
			t.Errorf("record %q, want dataset", rec.ID())
		}
	}
}
