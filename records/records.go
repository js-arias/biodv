// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package records implements
// a records database.
package records

import (
	"strings"

	"github.com/js-arias/biodv"
)

// Default datbase directory
const recDir = "records"

// TaxFileName returns the expected file name,
// for the records of a given taxon.
func taxFileName(name string) string {
	name = biodv.TaxCanon(name)
	if name == "" {
		return ""
	}
	name = strings.Join(strings.Fields(name), "-")
	return name + ".stz"
}

// Basic keys for the records database
const (
	taxonKey       = "taxon"
	idKey          = "id"
	basisKey       = "basis"
	dateKey        = "date"
	countryKey     = "country"
	stateKey       = "state"
	countyKey      = "county"
	localityKey    = "locality"
	collectorKey   = "collector"
	latlonKey      = "latlon"
	uncertaintyKey = "uncertainty"
	altitudeKey    = "altitude"
	depthKey       = "depth"
	geosourceKey   = "geosource"
	validationKey  = "validation"
)
