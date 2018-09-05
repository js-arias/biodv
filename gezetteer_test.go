// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package biodv

import (
	"testing"

	"github.com/js-arias/biodv/geography"

	"github.com/pkg/errors"
)

var pointList = []geography.Position{
	{Lat: -34, Lon: -58},
	{Lat: -34, Lon: -56},
	{Lat: -33, Lon: -70},
	{Lat: -25, Lon: -57},
	{Lat: -15, Lon: -47},
	{Lat: -12, Lon: -77},
	{Lat: -16, Lon: -68},
	{Lat: 0, Lon: -78},
	{Lat: 4, Lon: -74},
	{Lat: 10, Lon: -66},
}

func TestGeoScan(t *testing.T) {

	// Expected GeoScan usage
	sc := NewGeoScan(10)
	go func(x *GeoScan) {
		for _, p := range pointList {
			if !x.Add(p, nil) {
				break
			}
		}
		x.Add(geography.NewPosition(), nil)
	}(sc)

	c := 0
	for sc.Scan() {
		p := sc.Position()
		if !p.Equal(pointList[c]) {
			t.Errorf("point (%f,%f) want (%f,%f)", p.Lat, p.Lon, pointList[c].Lat, pointList[c].Lon)
		}
		c++
	}
	if err := sc.Err(); err != nil {
		t.Errorf("geoscan unexpected error: %v", err)
	}
	if len(pointList) != c {
		t.Errorf("scanned points %d, want %d", c, len(pointList))
	}

	// Closing GeoScan before finish
	sc = NewGeoScan(10)
	go func(x *GeoScan) {
		for _, p := range pointList {
			if !x.Add(p, nil) {
				break
			}
		}
		x.Add(geography.NewPosition(), nil)
	}(sc)

	c = 0
	for sc.Scan() {
		p := sc.Position()
		if !p.Equal(pointList[c]) {
			t.Errorf("point (%f,%f) want (%f,%f)", p.Lat, p.Lon, pointList[c].Lat, pointList[c].Lon)
		}
		c++
		if p.Equal(geography.Position{Lat: -15, Lon: -47}) {
			sc.Close()
		}
	}
	if err := sc.Err(); err != nil {
		t.Errorf("geoscan unexpected error: %v", err)
	}
	if c > 5 {
		t.Errorf("scanned points %d, want %d", c, 5)
	}

	// Expected GeoScan usage
	sc = NewGeoScan(10)
	go func(x *GeoScan) {
		for i, p := range pointList {
			if i == 4 {
				x.Add(geography.NewPosition(), errors.New("mock error"))
				return
			}
			if !x.Add(p, nil) {
				break
			}
		}
		x.Add(geography.NewPosition(), nil)
	}(sc)

	c = 0
	for sc.Scan() {
		p := sc.Position()
		if !p.Equal(pointList[c]) {
			t.Errorf("point (%f,%f) want (%f,%f)", p.Lat, p.Lon, pointList[c].Lat, pointList[c].Lon)
		}
		c++
	}
	if err := sc.Err(); err == nil {
		t.Errorf("geoscan expecting error")
	}
	if c > 4 {
		t.Errorf("scanned points %d, want %d", c, 4)
	}
}
