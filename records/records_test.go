// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package records

import (
	"testing"

	"github.com/js-arias/biodv"
)

var testData = []struct {
	taxon string
	basis biodv.BasisOfRecord
	id    string
	lat   float64
	lon   float64
}{
	{"Larus argentatus", biodv.Machine, "Larus-argentatus:1", 50.223982, 1.596802},
	{"Felis concolor couguar", biodv.Preserved, "Felis-concolor-couguar:1", 360, 360},
	{"Felis concolor", biodv.Preserved, "MSU:MR:MR.8672", 0.25, -79.8333},
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
