// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package biodv contains
// main interfaces and types
// for a basic biodiversity database.
package biodv

import "strings"

// A Taxonomy is a taxonomy database.
type Taxonomy interface {
	// Taxon returns a list of taxons with a given name.
	Taxon(name string) *TaxScan

	// TaxID returns the taxon with a given ID.
	TaxID(id string) (Taxon, error)

	// Synonyms returns a list taxons synonyms of a given ID.
	Synonyms(id string) *TaxScan

	// Children returns a list of taxon children of a given ID,
	// if the ID is empty,
	// it will return the taxons attached to the root
	// of the taxonomy.
	Children(id string) *TaxScan
}

// A Taxon is a taxon name in a taxonomy.
type Taxon interface {
	// Name returns the canonical name of the current taxon.
	Name() string

	// ID returns the ID of the current taxon.
	ID() string

	// Parent returns the ID of the taxon's parent.
	Parent() string

	// Rank returns the linnean rank of the current taxon.
	Rank() Rank

	// IsCorrect returns true if the taxon
	// is a correct name
	// (i.e. not a synonym).
	IsCorrect() bool

	// Keys returns a list of additional fields
	// stored in the taxon.
	Keys() []string

	// Value returns the value
	// of an additional field stored in the taxon.
	Value(key string) string
}

// Common keys used for a Taxon.
const (
	TaxAuthor = "author"    // Author of the taxon name
	TaxExtern = "extern"    // Extern IDs
	TaxRef    = "reference" // A bibliographic reference
	TaxSource = "source"    // Source of taxonomic data
)

// RecDB is a record database.
type RecDB interface {
	// TaxRecs returns a list of records from a given taxon ID.
	TaxRecs(id string) *RecScan

	// RecID returns the record with a given ID.
	RecID(id string) (Record, error)
}

// A Record is an specimen record.
type Record interface {
	// Taxon returns the ID of the taxon
	// assigned to the specimen.
	Taxon() string

	// ID return the ID of the current specimen.
	ID() string

	// Basis returns the basis of the specimen record.
	Basis() BasisOfRecord

	// CollEvent is the collection event of the record.
	CollEvent() CollectionEvent

	// GeoRef returns a geographic point.
	//
	// If the record is not georeferenced
	// is should return an invalid Point.
	GeoRef() Point

	// Keys returns a list of additional fields
	// stored in the record.
	Keys() []string

	// Value returns the value
	// of an additional field stored in the record.
	Value(key string) string
}

// Common keys used for Record.
const (
	RecRef      = "reference"  // A bibliographic reference
	RecDataset  = "dataset"    // Source of the specimen data
	RecCatalog  = "catalog"    // museum catalog number
	RecDeterm   = "determiner" // the person who identified the specimen
	RecExtern   = "extern"     // Extern IDs
	RecComment  = "comment"    // A free text comment
	RecOrganism = "organism"   // An ID of the organism
	RecSex      = "sex"        // Sex of the organism
	RecStage    = "stage"      // Life stage of the organism
	RecAltitude = "altitude"   // Elevation of flying organism
)

// ParseDriverString separates a driver
// and its parameter if it is set
// in the form <driver>:<param>.
func ParseDriverString(str string) (driver, param string) {
	i := strings.Index(str, ":")
	if i < 0 {
		return str, ""
	}
	return str[:i], str[i+1:]
}
