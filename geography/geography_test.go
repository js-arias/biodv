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
