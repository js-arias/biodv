// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package biodv

import "math"

// A Point is a georeferencenced point record.
type Point struct {
	Lat         float64
	Lon         float64
	Elevation   uint   // in meters
	Depth       uint   // in meters
	Source      string // source of the reference
	Uncertainty uint   // georeference uncertainty in meters
	Validation  string // source of the validation
}

// Maximum and minimum values for geographic coordinates
const (
	MinLat = -90
	MaxLat = 90
	MinLon = -180
	MaxLon = 180
)

// InvalidPoint returns a new Point without a valid georeference.
func InvalidPoint() Point {
	return Point{Lat: 180, Lon: 360}
}

// IsValid returns true if a geographic point is valid.
func (p Point) IsValid() bool {
	if (p.Lon <= MaxLon) && (p.Lon > MinLon) {
		if (p.Lat < MaxLat) && (p.Lat > MinLat) {
			return true
		}
	}
	return false
}

// Precision is the default precision level
// when comparing georeferences.
const GeoPrecision = 0.000001

// Equal returns true if two points are equal
// under a precision level.
// Invalid points are considered equals.
func (p Point) Equal(op Point, prec float64) bool {
	if !p.IsValid() {
		return !op.IsValid()
	}

	if !op.IsValid() {
		return false
	}

	if prec == 0 {
		prec = GeoPrecision
	}

	if math.Abs(p.Lat-op.Lat) > prec || math.Abs(p.Lon-op.Lon) > prec {
		return false
	}
	return true
}
