// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package stanza

import (
	"bytes"
	"strings"
	"testing"
)

var blob = `
# Country data facts
Name:	República Argentina
Common:	Argentina
ISO3166: AR
Capital: Buenos Aires
Population: 42669500
Anthem:	Ya su trono dignísimo abrieron
	las Provincias Unidas del Sud
	y los libres del mundo responden:
	"¡Al gran pueblo argentino, salud!"
%
Name:	대한민국
Common:	South Korea
ISO3166: KR
Capital: Seoul
Population: 51302044
Anthem:	무궁화 삼천리 화려강산
	대한 사람, 대한으로 길이 보전하세
%%
Name:	中华人民共和国
Common:	China
ISO3166: CN
Capital: Beijing
Population: 1339724852
%
Name:	Росси́я
Common:	Russia
ISO3166: RU
Capital: Moscow
Population: 144192450
Anthem: Славься, Отечество наше свободное,
	Братских народов союз вековой,
	Предками данная мудрость народная!
	Славься, страна! Мы гордимся тобой!
%%
`

func TestScan(t *testing.T) {
	sc := NewScanner(strings.NewReader(blob))
	i := 0
	for sc.Scan() {
		rec := sc.Record()
		if _, ok := rec["common"]; !ok {
			t.Errorf("field %q not found", "common")
		}
		if _, ok := rec["iso3166"]; !ok {
			t.Errorf("field %q not found", "iso3166")
		}
		if rec["common"] != "China" {
			if len(rec) != 6 {
				t.Errorf("fields %d, want 6", len(rec))
			}
			an, ok := rec["anthem"]
			if !ok {
				t.Errorf("field %q not found", "anthem")
			}
			if v := len(strings.Split(an, "\n")); v < 2 {
				t.Errorf("field %q should be multiline", "anthem")
			}
		} else if len(rec) != 5 {
			t.Errorf("fields %d, want 5", len(rec))
		}
		i++
	}
	if err := sc.Err(); err != nil {
		t.Errorf("unexpected scanner error: %v", err)
	}
	if i != 4 {
		t.Errorf("found %d records, want 4", i)
	}
}

func TestWrite(t *testing.T) {
	sc := NewScanner(strings.NewReader(blob))
	country := make(map[string]map[string]string)
	out := &bytes.Buffer{}
	w := NewWriter(out)
	w.SetFields([]string{"name", "common", "iso3166", "capital", "population", "anthem"})
	for sc.Scan() {
		rec := sc.Record()
		err := w.Write(rec)
		if err != nil {
			t.Errorf("writing error: %v", err)
		}
		country[rec["common"]] = rec
	}
	if err := w.Flush(); err != nil {
		t.Errorf("flushing error: %v", err)
	}

	sc = NewScanner(strings.NewReader(blob))

	i := 0
	for sc.Scan() {
		rec := sc.Record()
		p, ok := country[rec["common"]]
		if !ok {
			t.Errorf("country %q not found", rec["common"])
			continue
		}
		for f, v := range rec {
			if p[f] != v {
				t.Errorf("country %q: get %q, want %q", p["common"], p[f], v)
			}
		}
		i++
	}
	if err := sc.Err(); err != nil {
		t.Errorf("unexpected scanner error: %v", err)
	}
	if i != 4 {
		t.Errorf("found %d records, want 4", i)
	}
}
