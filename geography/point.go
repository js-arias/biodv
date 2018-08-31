// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package geography

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
