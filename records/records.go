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
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/js-arias/biodv"

	"github.com/pkg/errors"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Default database directory
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
	tids    map[string]*taxon
	ids     map[string]*Record
	changed bool
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
	db      *DB
	id      string
	recs    []*Record
	changed bool
}

// Record is a record stored in a DB.
// Record implements the biodv.Record interface.
type Record struct {
	taxon *taxon
	data  map[string]string
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

// SetCollEvent sets the values
// of a collection event.
func (rec *Record) SetCollEvent(event biodv.CollectionEvent) {
	old := rec.CollEvent()

	if !old.Date.Equal(event.Date) {
		if event.Date.IsZero() {
			delete(rec.data, dateKey)
		} else {
			rec.data[dateKey] = event.Date.Format(time.RFC3339)
		}
		rec.taxon.changed = true
	}

	if old.Country != event.Country {
		if event.Country == "" {
			delete(rec.data, countryKey)
		} else {
			rec.data[countryKey] = event.Country
		}
		rec.taxon.changed = true
	}

	if old.State != event.State {
		if event.State == "" {
			delete(rec.data, stateKey)
		} else {
			rec.data[stateKey] = event.State
		}
		rec.taxon.changed = true
	}

	if old.County != event.County {
		if event.County == "" {
			delete(rec.data, countyKey)
		} else {
			rec.data[countyKey] = event.County
		}
		rec.taxon.changed = true
	}

	if old.Locality != event.Locality {
		if event.Locality == "" {
			delete(rec.data, localityKey)
		} else {
			rec.data[localityKey] = event.Locality
		}
		rec.taxon.changed = true
	}

	if old.Collector != event.Collector {
		if event.Collector == "" {
			delete(rec.data, collectorKey)
		} else {
			rec.data[collectorKey] = event.Collector
		}
		rec.taxon.changed = true
	}
}

// Precision is the default precision level
// when comparing georeferences.
const Precision = 0.000001

// SetGeoRef sets the values
// of a georeference.
// The prec parameter controls the precision
// used to compare the old and new
// georeferences
// (if any).
func (rec *Record) SetGeoRef(geo biodv.Point, prec float64) {
	old := rec.GeoRef()

	if !geo.IsValid() {
		if rec.data[latlonKey] != "" {
			delete(rec.data, latlonKey)
			rec.taxon.changed = true
		}
	} else if !old.IsValid() {
		storeLatLon(rec.data, geo.Lat, geo.Lon)
		rec.taxon.changed = true
	} else if math.Abs(geo.Lat-old.Lat) > prec || math.Abs(geo.Lon-old.Lon) > prec {
		storeLatLon(rec.data, geo.Lat, geo.Lon)
		rec.taxon.changed = true
	}

	if geo.Altitude > 0 && geo.Depth < 0 {
		geo.Altitude = old.Altitude
		geo.Depth = old.Depth
	}

	if geo.Altitude == 0 {
		if rec.data[altitudeKey] != "" {
			delete(rec.data, altitudeKey)
			rec.taxon.changed = true
		}
	} else if math.Abs(old.Altitude-geo.Altitude) > prec {
		rec.data[altitudeKey] = strconv.FormatFloat(geo.Altitude, 'f', -1, 64)
		rec.taxon.changed = true
	}

	if geo.Depth == 0 {
		if rec.data[depthKey] != "" {
			delete(rec.data, depthKey)
			rec.taxon.changed = true
		}
	} else if math.Abs(old.Depth-geo.Depth) > prec {
		rec.data[depthKey] = strconv.FormatFloat(geo.Depth, 'f', -1, 64)
		rec.taxon.changed = true
	}

	if geo.Source != old.Source {
		if geo.Source == "" {
			delete(rec.data, geosourceKey)
		} else {
			rec.data[geosourceKey] = geo.Source
		}
		rec.taxon.changed = true
	}

	if geo.Uncertainty != old.Uncertainty {
		if geo.Uncertainty == 0 {
			delete(rec.data, uncertaintyKey)
		} else {
			rec.data[uncertaintyKey] = strconv.Itoa(int(geo.Uncertainty))
		}
		rec.taxon.changed = true
	}

	if geo.Validation != old.Validation {
		if geo.Validation == "" {
			delete(rec.data, validationKey)
		} else {
			rec.data[validationKey] = geo.Validation
		}
		rec.taxon.changed = true
	}
}

// Set sets a value from a given key.
// The value should be transformed into
// a string.
// When an empty string is used as value,
// the stored value will be deleted.
func (rec *Record) Set(key, value string) error {
	key = strings.ToLower(strings.Join(strings.Fields(key), "-"))
	if key == "" {
		return nil
	}
	value = strings.TrimSpace(value)

	switch key {
	case taxonKey:
		fallthrough
	case idKey:
		fallthrough
	case dateKey:
		fallthrough
	case countryKey:
		fallthrough
	case stateKey:
		fallthrough
	case localityKey:
		fallthrough
	case collectorKey:
		fallthrough
	case latlonKey:
		fallthrough
	case uncertaintyKey:
		fallthrough
	case altitudeKey:
		fallthrough
	case depthKey:
		fallthrough
	case geosourceKey:
		fallthrough
	case validationKey:
		return errors.Errorf("records: record: invalid key value: %s", key)
	case basisKey:
		b := biodv.GetBasis(value)
		if rec.Basis() == biodv.UnknownBasis && b != biodv.UnknownBasis {
			rec.data[basisKey] = b.String()
			rec.taxon.changed = true
		}
		return nil
	case biodv.RecCatalog:
		v := rec.data[biodv.RecCatalog]
		if value == v {
			return nil
		}
		if v != "" {
			return errors.Errorf("records: record: catalog number cannot be changed")
		}
		db := rec.taxon.db
		if _, dup := db.ids[value]; dup {
			return errors.Errorf("records: record: catalog %s already in use", value)
		}
		rec.data[biodv.RecCatalog] = value
		rec.taxon.changed = true
		db.ids[value] = rec
	case biodv.RecExtern:
		if value == "" {
			return nil
		}
		srv := getService(value)
		if srv == "" {
			return errors.New("records: record: invalid extern ID value")
		}
		ext := strings.Fields(rec.data[key])
		db := rec.taxon.db
		if srv+":" == value {
			// empty extern ID
			// deletes the extern ID from database
			for i, e := range ext {
				if srv != getService(e) {
					continue
				}
				delete(db.ids, e)
				n := append(ext[:i], ext[i+1:]...)
				rec.data[key] = strings.Join(n, " ")
				rec.taxon.changed = true
				return nil
			}
			return nil
		}

		// check if the given ID is already in use
		if _, dup := db.ids[value]; dup {
			return errors.Errorf("records: record: extern ID %s already in use", value)
		}

		// if the service is already set
		for i, e := range ext {
			if srv != getService(e) {
				continue
			}
			delete(db.ids, e)
			db.ids[value] = rec
			ext[i] = value
			rec.data[key] = strings.Join(ext, " ")
			rec.taxon.changed = true
			return nil
		}

		// the service is new
		ext = append(ext, value)
		sort.Strings(ext)
		db.ids[value] = rec
		rec.data[key] = strings.Join(ext, " ")
		rec.taxon.changed = true
		return nil
	default:
		v := rec.data[key]
		if v == value {
			return nil
		}
		if v == "" {
			delete(rec.data, key)
			rec.taxon.changed = true
			return nil
		}
		rec.data[key] = value
		rec.taxon.changed = true
		return nil
	}
	return nil
}

// GetService returns the service
// (extern Record identifier)
// that provides an external ID.
func getService(id string) string {
	i := strings.Index(id, ":")
	if i <= 0 {
		return ""
	}
	return id[:i]
}

// Scan uses a scanner
// to load data to a database.
func (db *DB) scan(sc *Scanner) error {
	for sc.Scan() {
		r := sc.Record()
		geo := r.GeoRef()
		rec, err := db.Add(r.Taxon(), r.ID(), r.Value(biodv.RecCatalog), r.Basis(), geo.Lat, geo.Lon)
		if err != nil {
			sc.Close()
			return err
		}
		rec.SetCollEvent(r.CollEvent())
		rec.SetGeoRef(r.GeoRef(), Precision)
		keys := r.Keys()
		for _, k := range keys {
			if err := rec.Set(k, r.Value(k)); err != nil {
				return err
			}
		}
	}
	return sc.Err()
}

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
		return nil, errors.Errorf("records: db: add %q: record already in database", id)
	}
	if catalog != "" && db.ids[catalog] != nil {
		return nil, errors.Errorf("records: db: add %q: catalog %q: catalog number already in database", id, catalog)
	}

	tax, ok := db.tids[taxID]
	if !ok {
		tax = &taxon{id: taxID, db: db}
		db.tids[taxID] = tax
		db.changed = true
	}
	data := make(map[string]string)
	data[taxonKey] = taxID
	data[idKey] = id
	if catalog != "" {
		data[biodv.RecCatalog] = catalog
	}
	data[basisKey] = basis.String()
	storeLatLon(data, lat, lon)
	rec := &Record{tax, data}
	tax.recs = append(tax.recs, rec)
	db.ids[id] = rec
	if catalog != "" && catalog != id {
		db.ids[catalog] = rec
	}
	tax.changed = true
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
