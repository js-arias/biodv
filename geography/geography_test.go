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
}
