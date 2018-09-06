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

// EarthRadius is the WGS84 mean radius (in meters).
const EarthRadius = 6371009

// Distance returns the distance,
// in meters
// (using the WGS84 mean radius)
// between two points.
// This function ignores the uncertainty of the positions.
func (p Position) Distance(op Position) uint {
	l1, l2 := toRad(p.Lat), toRad(op.Lat)
	dLon := toRad(p.Lon) - toRad(op.Lon)
	dLat := l1 - l2
	s1 := math.Sin(dLat / 2)
	s1 *= s1
	f1 := math.Sin(dLon / 2)
	s2 := f1 * f1 * math.Cos(l1) * math.Cos(l2)
	v := math.Sqrt(s1 + s2)
	return uint(2 * EarthRadius * math.Asin(v))
}

// MaxDist returns the maximum distance,
// in meters
// (using the WGS84 mean radius)
// between two points.
// The maximum distance is equal to the distance
// plus the uncertainty of both points.
func (p Position) MaxDist(op Position) uint {
	return p.Distance(op) + p.Uncertainty + op.Uncertainty
}

func toRad(angle float64) float64 {
	return angle * math.Pi / 180
}
