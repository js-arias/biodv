// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package records implements
// a records database.
package records

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/js-arias/biodv"

	"github.com/pkg/errors"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

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

// DB is a records database
// for reading and writing data.
// DB implements the biodv.RecDB interface.
type DB struct {
	tids map[string]*taxon
	ids  map[string]*Record
}

// TaxRecs returns a list of records from a given taxon ID.
// This function is for compatibility with biodv.RecDB interface.
func (db *DB) TaxRecs(id string) *biodv.RecScan {
	sc := biodv.NewRecScan(1)
	id = strings.TrimSpace(id)
	if id == "" {
		sc.Add(nil, errors.Errorf("records: db: taxrec: empty taxon ID"))
		return sc
	}
	tax, ok := db.tids[id]
	if !ok {
		sc.Add(nil, nil)
		return sc
	}
	go func() {
		for _, r := range tax.recs {
			sc.Add(r, nil)
		}
		sc.Add(nil, nil)
	}()
	return sc
}

// RecID returns the record with a given ID.
func (db *DB) RecID(id string) (biodv.Record, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.Errorf("records: db: recid: empty record ID")
	}
	if rec, ok := db.ids[id]; ok {
		return rec, nil
	}
	return nil, nil
}

// Taxon stores a list of records.
type taxon struct {
	id   string
	recs []*Record
}

// Record is a record stored in a DB.
// Record implements the biodv.Record interface.
type Record struct {
	data map[string]string
}

// Taxon returns the ID of the taxon
// assigned to the specimen.
func (rec *Record) Taxon() string {
	return rec.data[taxonKey]
}

// ID return the ID of the current specimen.
func (rec *Record) ID() string {
	return rec.data[idKey]
}

// Basis returns the basis of the specimen record.
func (rec *Record) Basis() biodv.BasisOfRecord {
	return biodv.GetBasis(rec.data[basisKey])
}

// CollEvent is the collection event of the record.
func (rec *Record) CollEvent() biodv.CollectionEvent {
	return recmap(rec.data).CollEvent()
}

// GeoRef returns a geographic point.
//
// If the record is not georeferenced
// is should return an invalid Point.
func (rec *Record) GeoRef() biodv.Point {
	return recmap(rec.data).GeoRef()
}

// Keys returns a list of additional fields
// stored in the record.
func (rec *Record) Keys() []string {
	return recmap(rec.data).Keys()
}

// Value returns the value
// of an additional field stored in the record.
func (rec *Record) Value(key string) string {
	key = strings.ToLower(strings.TrimSpace(key))
	if key == "" {
		return ""
	}
	return rec.data[key]
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

// Add adds a new record to a DB.
func (db *DB) Add(taxID, id, catalog string, basis biodv.BasisOfRecord, lat, lon float64) (*Record, error) {
	taxID = strings.TrimSpace(taxID)
	if taxID == "" {
		return nil, errors.New("records: db: add: empty taxon")
	}

	id = strings.TrimSpace(id)
	catalog = strings.TrimSpace(catalog)
	if id == "" {
		if catalog == "" {
			id = db.generateID(taxID)
		} else {
			id = catalog
		}
	}

	if _, dup := db.ids[id]; dup {
		return nil, errors.Errorf("records: db: add %q: record already in database")
	}

	tax, ok := db.tids[taxID]
	if !ok {
		tax = &taxon{id: taxID}
		db.tids[taxID] = tax
	}
	data := make(map[string]string)
	data[taxonKey] = taxID
	data[idKey] = id
	if catalog != "" {
		data[biodv.RecCatalog] = catalog
	}
	data[basisKey] = basis.String()
	storeLatLon(data, lat, lon)
	rec := &Record{data}
	tax.recs = append(tax.recs, rec)
	db.ids[id] = rec
	if catalog != "" && catalog != id {
		db.ids[catalog] = rec
	}
	return rec, nil
}

// GenerateID generates a new ID for a record
// of a given taxon.
func (db *DB) generateID(taxID string) string {
	t := time.Now()
	root := taxID + ":" + t.Format("20060102150405")
	for {
		id := fmt.Sprintf("%s-%d", root, rand.Intn(100000000))
		if _, ok := db.ids[id]; !ok {
			return id
		}
	}
}
