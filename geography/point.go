// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package geography

import "math"

// Maximum and minimum values for geographic coordinates
const (
	MinLat = -90
	MaxLat = 90
	MinLon = -180
	MaxLon = 180
)

// IsValidCoord returns true
// if a latitude, longitude pair
// is valid.
func IsValidCoord(lat, lon float64) bool {
	if lat < MinLat || lat > MaxLat {
		return false
	}
	if lon <= MinLon || lon > MaxLon {
		return false
	}
	return true
}

// A Position is a point with geographic coordinates.
type Position struct {
	Lat         float64
	Lon         float64
	Elevation   uint   // in meters
	Source      string // source of the position coordinates
	Uncertainty uint   // uncertainty in meters
	Validation  string // source of a validation for the coordinates
}

// NewPosition creates a new position
// with no coordinates.
// Use these function instead of an empty Position
// as zero coordinates are geographically valid.
func NewPosition() Position {
	return Position{Lat: MaxLat * 2, Lon: MaxLon * 2}
}

// IsValid returns true if a point
// is a valid geographic point.
func (p Position) IsValid() bool {
	return IsValidCoord(p.Lat, p.Lon)
}

// Precision returns the current precision level
// when comparing georeferences.
func Precision() float64 {
	return precision
}

// Precision is the precision level,
// used when comparing georeferences.
var precision = 0.000001

// SetPrecision sets a new precision level.
func SetPrecision(lv float64) {
	if lv <= 0 {
		return
	}
	precision = lv
}

// Equal returns true if two points are equal
// under the package precision level.
// Invalid points are considered equals.
func (p Position) Equal(op Position) bool {
	if !p.IsValid() {
		return !op.IsValid()
	}

	if !op.IsValid() {
		return false
	}

	if math.Abs(p.Lat-op.Lat) > precision || math.Abs(p.Lon-op.Lon) > precision {
		return false
	}
	return true
}
