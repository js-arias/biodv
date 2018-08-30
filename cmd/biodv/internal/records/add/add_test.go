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
	"github.com/js-arias/biodv/records"
)

var blob = `
Taxon	Longitude	Latitude	Comment
Rhododendron maddenii	98.296110	27.713055	gbif:543321173
Rhododendron maddenii	105.736390	17.644444	gbif:574874259
Rhododendron maddenii	103.600000	22.420000	gbif:322085347
Rhododendron vaccinioides	87.669600	27.459300	Brown et al. 2006 (ooo)
Rhododendron vaccinioides	88.057300	27.821500	Brown et al. 2006 (ooo)
Rhododendron vaccinioides	89.026400	27.459300	Brown et al. 2006 (ooo)
Rhododendron euonumifolium	104.122690	22.607246	Brown et al. 2006 (q)
Rhododendron euonumifolium	106.607942	25.660870	Brown et al. 2006 (q)
Rhododendron euonumifolium	107.119611	25.376812	Brown et al. 2006 (q)
Rhododendron kawakamii	121.362500	24.220655	Brown et al. 2006 (cc)
Rhododendron kawakamii	121.250000	23.542754	Brown et al. 2006 (cc)
Rhododendron kawakamii	120.950000	23.485305	Brown et al. 2006 (cc)
Rhododendron retusum	97.384939	4.205607	Brown et al. 2006 (aaa)
Rhododendron retusum	98.541331	3.196262	Brown et al. 2006 (aaa)
Rhododendron retusum	98.354816	3.158879	Brown et al. 2006 (aaa)
Rhododendron baenitzianum	142.890203	-6.410127	Brown et al. 2006 (h)
Rhododendron baenitzianum	141.192568	-5.207089	Brown et al. 2006 (h)
Rhododendron baenitzianum	141.905574	-3.168608	Brown et al. 2006 (h)
`

func TestRecAddCmdRead(t *testing.T) {
	r := strings.NewReader(blob)
	recs, err := records.Open("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := read(recs, r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ls, err := getRecList(recs.TaxRecs("Rhododendron kawakamii"))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(ls) != 3 {
		t.Errorf("%d records read, want %d", len(ls), 3)
	}
	ls, err = getRecList(recs.TaxRecs("Rhododendron maddenii"))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	for _, r := range ls {
		if !strings.HasPrefix(r.Value(biodv.RecComment), "gbif:") {
			t.Errorf("on taxon %q, comment %q, want \"gbif:xxx\"", r.Taxon(), r.Value(biodv.RecComment))
		}
	}
}

func getRecList(sr *biodv.RecScan) ([]biodv.Record, error) {
	var ls []biodv.Record
	for sr.Scan() {
		r := sr.Record()
		ls = append(ls, r)
	}
	if err := sr.Err(); err != nil {
		return nil, err
	}
	return ls, nil
}
