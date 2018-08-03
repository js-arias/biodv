// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package add

import (
	"strings"
	"testing"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/taxonomy"
)

var blob = `

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

Eurypteryx curtus
Pachyornis geranoides
Pachyornis elephantopus
Pachyornis australis
Dinornis robustus
Megalapteryx didinus

`

func TestTaxAddCmdRead(t *testing.T) {
	r := strings.NewReader(blob)
	db, err := taxonomy.Open("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := read(db, r, biodv.Species); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tax, _ := db.TaxID("Rhedosaurus"); tax != nil {
		t.Errorf("taxon \"Rhedosaurus\" should not be on database")
	}
	if tax, _ := db.TaxID("Tinamus major"); tax == nil {
		t.Errorf("taxon \"Tinamus major\" should be on database")
	}
	if tax, _ := db.TaxID("Pachyornis"); tax == nil {
		t.Errorf("taxon \"Pachyornis\" should be on database")
	}
}
