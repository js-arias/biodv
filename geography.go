// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package biodv

import (
	"math"

	"github.com/js-arias/biodv/geography"
)

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

// InvalidPoint returns a new Point without a valid georeference.
func InvalidPoint() Point {
	return Point{Lat: 180, Lon: 360}
}

// IsValid returns true if a geographic point is valid.
func (p Point) IsValid() bool {
	return geography.IsValidCoord(p.Lat, p.Lon)
}

// Precision is the default precision level
// when comparing georeferences.
const GeoPrecision = 0.000001

// Equal returns true if two points are equal
// under a precision level.
// Invalid points are considered equals.
func (p Point) Equal(op Point, prec float64) bool {
	if !geography.IsValidCoord(p.Lat, p.Lon) {
		return !geography.IsValidCoord(op.Lat, op.Lon)
	}

	if !geography.IsValidCoord(op.Lat, op.Lon) {
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
