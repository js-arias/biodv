// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package geography

import "testing"

func TestCountry(t *testing.T) {
	testData := []struct {
		code string
		want string
	}{
		{"GR", "Greece"},
		{"cd", "Congo, the Democratic Republic of the"},
		{"MAC", ""},
		{"mdg", ""},
		{"", ""},
	}

	for _, d := range testData {
		if Country(d.code) != d.want {
			t.Errorf("code %q = %q, want %q", d.code, Country(d.code), d.want)
		}
	}
}

func TestVadidCoord(t *testing.T) {
	testData := []struct {
		lat  float64
		lon  float64
		want bool
	}{
		{26, -115, true},
		{-34, -71, true},
		{41, 13, true},
		{-17, 129, true},
		{90, 112, true},
		{-90, -16, true},
		{180, -78, false},
		{-200, -118, false},
		{16, -180, false},
		{45, 180, true},
		{-31, 270, false},
		{11, -311, false},
	}

	for _, d := range testData {
		if IsValidCoord(d.lat, d.lon) != d.want {
			t.Errorf("point [%.1f %.1f] %v", d.lat, d.lon, IsValidCoord(d.lat, d.lon))
		}
	}
	if p := NewPosition(); p.IsValid() {
		t.Errorf("a new position should be invalid")
	}

	for _, d := range testData {
		p := Position{Lat: d.lat, Lon: d.lon}
		if p.IsValid() != d.want {
			t.Errorf("position [%.1f %.1f] %v", p.Lat, p.Lon, p.IsValid())
		}
	}
}

func TestEqual(t *testing.T) {
	testData := []struct {
		p    Position
		op   Position
		want bool
	}{
		{Position{Lat: 26, Lon: -115}, Position{Lat: 26, Lon: -115}, true},
		{Position{Lat: -34.01, Lon: -71.11}, Position{Lat: -34.01, Lon: -71.11}, true},
		{Position{Lat: 41, Lon: 13}, Position{Lat: 41.0000001, Lon: 12.9999999}, true},
		{Position{Lat: -17.76, Lon: 129.49}, Position{Lat: -17.64, Lon: 129.49}, false},
		{Position{Lat: -34.01, Lon: -71.11}, Position{Lat: -34.010000001, Lon: -71.118}, false},
		{Position{Lat: 26, Lon: -115}, NewPosition(), false},
		{NewPosition(), NewPosition(), true},
	}

	for _, d := range testData {
		if d.p.Equal(d.op) != d.want {
			t.Errorf("position [%.5f %.5f] - [%.5f %.5f] %v", d.p.Lat, d.p.Lon, d.op.Lat, d.op.Lon, d.p.Equal(d.op))
		}
	}
}

func TestDistance(t *testing.T) {
	testData := []struct {
		p       Position
		op      Position
		want    uint
		wantMax uint
	}{
		{Position{Lat: 40.714269, Lon: -74.005973, Uncertainty: 3036}, Position{Lat: 45.639725, Lon: 5.737419, Uncertainty: 3036}, 6215190, 6221262},
		{Position{Lat: 40.714269, Lon: -74.005973, Uncertainty: 3036}, Position{Lat: 51.514125, Lon: -0.093689, Uncertainty: 301}, 5572131, 5575468},
		{Position{Lat: 45.639725, Lon: 5.737419, Uncertainty: 3036}, Position{Lat: 51.514125, Lon: -0.093689, Uncertainty: 301}, 780916, 784253},
	}

	for _, d := range testData {
		if d.p.Distance(d.op) != d.want {
			t.Errorf("distance [%.5f %.5f] - [%.5f %.5f] %d, want %d", d.p.Lat, d.p.Lon, d.op.Lat, d.op.Lon, d.p.Distance(d.op), d.want)
		}
		if d.p.MaxDist(d.op) != d.wantMax {
			t.Errorf("maximum distance [%.5f %.5f] - [%.5f %.5f] %d, want %d", d.p.Lat, d.p.Lon, d.op.Lat, d.op.Lon, d.p.MaxDist(d.op), d.wantMax)
		}
	}
}
