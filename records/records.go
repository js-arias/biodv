// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package records

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/encoding/stanza"
	"github.com/js-arias/biodv/geography"

	"github.com/pkg/errors"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Default database directory
const recDir = "records"
const recTaxList = "taxons.lst"

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
	path    string
	tids    map[string]*taxon
	ids     map[string]*Record
	changed bool
}

// TaxRecs returns a list of records from a given taxon ID.
// This function is for compatibility with biodv.RecDB interface.
//
// When using an editable DB prefer RecList.
func (db *DB) TaxRecs(id string) *biodv.RecScan {
	sc := biodv.NewRecScan(1)
	id = biodv.TaxCanon(id)
	if id == "" {
		sc.Add(nil, errors.Errorf("records: db: taxrec: empty taxon ID"))
		return sc
	}
	tax, ok := db.tids[id]
	if !ok {
		sc.Add(nil, nil)
		return sc
	}
	if !tax.sorted {
		sortRecords(tax.recs)
		tax.sorted = true
	}

	go func() {
		for _, r := range tax.recs {
			sc.Add(r, nil)
		}
		sc.Add(nil, nil)
	}()
	return sc
}

// RecList returns a list of records
// for a given taxon name.
func (db *DB) RecList(id string) []*Record {
	id = biodv.TaxCanon(id)
	if id == "" {
		return nil
	}
	tax, ok := db.tids[id]
	if !ok {
		return nil
	}
	if !tax.sorted {
		sortRecords(tax.recs)
		tax.sorted = true
	}
	ls := make([]*Record, len(tax.recs))
	copy(ls, tax.recs)
	return ls
}

// Move moves a record,
// to another taxon.
// If the destination taxon is not in the database,
// it will be added to it.
func (db *DB) Move(id, taxID string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil
	}
	rec := db.ids[id]
	if rec == nil {
		return nil
	}

	taxID = biodv.TaxCanon(taxID)
	if taxID == "" {
		return errors.Errorf("records: db: move: empty taxon ID")
	}

	old := rec.taxon
	if old.id == taxID {
		return nil
	}

	tax, ok := db.tids[taxID]
	if !ok {
		tax = &taxon{id: taxID, db: db}
		db.tids[taxID] = tax
		db.changed = true
	}

	old.removeRecord(rec)
	rec.data[taxonKey] = taxID
	tax.recs = append(tax.recs, rec)
	tax.changed = true
	tax.sorted = false
	return nil
}

// Delete removes a record from the records database.
func (db *DB) Delete(id string) {
	id = strings.TrimSpace(id)
	if id == "" {
		return
	}
	rec := db.ids[id]
	if rec == nil {
		return
	}

	tax := rec.taxon

	// remove all IDs
	ext := strings.Fields(rec.data[biodv.RecExtern])
	for _, e := range ext {
		delete(db.ids, e)
	}
	if cat := rec.data[biodv.RecCatalog]; cat != "" {
		delete(db.ids, cat)
	}
	delete(db.ids, rec.ID())

	tax.removeRecord(rec)
}

// Record returns a Record.
func (db *DB) Record(id string) *Record {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil
	}
	return db.ids[id]
}

// RecID returns the record with a given ID.
//
// When using an editable DB prefer Record.
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
	sorted  bool
}

// Commit saves a list of taxon records
// to a file.
func (tax *taxon) commit() (err error) {
	if !tax.changed {
		return nil
	}

	taxFile := taxFileName(tax.id)
	file := filepath.Join(tax.db.path, recDir, taxFile)

	// If there is no records in the taxon
	// removes the file with that records
	if len(tax.recs) == 0 {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			tax.changed = false
			return nil
		}
		if err = os.Remove(file); err != nil {
			return errors.Wrap(err, "records: db: commit")
		}
		tax.changed = false
		return nil
	}

	if !tax.sorted {
		sortRecords(tax.recs)
		tax.sorted = true
	}

	var f *os.File
	f, err = os.Create(file)
	if err != nil {
		return errors.Wrap(err, "records: db: commit")
	}
	defer func() {
		e1 := f.Close()
		if err == nil && e1 != nil {
			err = errors.Wrap(e1, "records: db: commit")
		}
	}()

	w := stanza.NewWriter(f)
	defer w.Flush()

	for _, rec := range tax.recs {
		if err = rec.encode(w); err != nil {
			return
		}
	}
	tax.changed = false
	return nil
}

func (tax *taxon) removeRecord(rec *Record) {
	for i, r := range tax.recs {
		if r != rec {
			continue
		}
		copy(tax.recs[i:], tax.recs[i+1:])
		tax.recs[len(tax.recs)-1] = nil
		tax.recs = tax.recs[:len(tax.recs)-1]
		tax.changed = true
		break
	}
	if len(tax.recs) == 0 {
		tax.db.changed = true
	}
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
func (rec *Record) GeoRef() geography.Position {
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
	elevationKey   = "elevation"
	geosourceKey   = "geosource"
	validationKey  = "validation"
	zKey           = "z"
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

	if old.Country() != event.Country() {
		if event.Country() == "" {
			delete(rec.data, countryKey)
			delete(rec.data, stateKey)
			delete(rec.data, countyKey)
			rec.taxon.changed = true
		} else if event.Country() != "" {
			rec.data[countryKey] = strings.ToUpper(event.CountryCode())
			rec.taxon.changed = true
		}
	}

	if old.Admin.State != event.Admin.State && event.Country() != "" {
		if event.State() == "" {
			delete(rec.data, stateKey)
		} else {
			rec.data[stateKey] = event.State()
		}
		rec.taxon.changed = true
	}

	if old.Admin.County != event.Admin.County && event.Country() != "" {
		if event.County() == "" {
			delete(rec.data, countyKey)
		} else {
			rec.data[countyKey] = event.County()
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

	if old.Z != event.Z {
		if event.Z == 0 {
			delete(rec.data, zKey)
		} else {
			rec.data[zKey] = strconv.Itoa(event.Z)
		}
	}
}

// SetGeoRef sets the values
// of a georeference.
func (rec *Record) SetGeoRef(geo geography.Position) {
	old := rec.GeoRef()

	if !geo.IsValid() {
		if rec.data[latlonKey] != "" {
			delete(rec.data, latlonKey)
			rec.taxon.changed = true
		}
	} else if !old.IsValid() {
		storeLatLon(rec.data, geo.Lat, geo.Lon)
		rec.taxon.changed = true
	} else if !geo.Equal(old) {
		storeLatLon(rec.data, geo.Lat, geo.Lon)
		rec.taxon.changed = true
	}

	if geo.Elevation == 0 {
		if rec.data[elevationKey] != "" {
			delete(rec.data, elevationKey)
			rec.taxon.changed = true
		}
	} else if geo.Elevation != old.Elevation {
		rec.data[elevationKey] = strconv.Itoa(int(geo.Elevation))
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
	case zKey:
		fallthrough
	case latlonKey:
		fallthrough
	case uncertaintyKey:
		fallthrough
	case elevationKey:
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
		rec.taxon.sorted = false
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
		if value == "" {
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

// Encode writes a record
// into a stanza writer.
func (rec *Record) encode(w *stanza.Writer) error {
	fields := []string{
		taxonKey,
		idKey,
		basisKey,
		dateKey,
		countryKey,
		stateKey,
		countyKey,
		localityKey,
		collectorKey,
		zKey,
		latlonKey,
		uncertaintyKey,
		elevationKey,
		geosourceKey,
		validationKey,
	}
	fields = append(fields, rec.Keys()...)
	w.SetFields(fields)

	if err := w.Write(rec.data); err != nil {
		return errors.Wrapf(err, "unable to write %s [taxon %s]", rec.ID(), rec.Taxon())
	}
	return nil
}

// SortRecords sorts records
// first from the catalog code,
// and then by its ID.
func sortRecords(ls []*Record) {
	sort.Sort(sortRec(ls))
}

type sortRec []*Record

func (s sortRec) Len() int      { return len(s) }
func (s sortRec) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s sortRec) Less(i, j int) bool {
	ci, cj := s[i].Value(biodv.RecCatalog), s[j].Value(biodv.RecCatalog)
	if ci != "" {
		if cj == "" {
			return true
		}
		return ci < cj
	}

	// catalog code are always "less" than catalog-less records
	if cj != "" {
		return false
	}

	return s[i].ID() < s[j].ID()
}

func init() {
	biodv.RegisterRec("biodv", biodv.RecDriver{open, nil, aboutBiodv})
}

// AboutBiodv returns a simple stated of the purpose of the driver.
func aboutBiodv() string {
	return "the default records database driver"
}

// Open opens a DB
// as a biodv.Taxonomy.
func open(path string) (biodv.RecDB, error) {
	return Open(path)
}

// Open opens a DB
// on a given path.
func Open(path string) (*DB, error) {
	db := &DB{
		path: path,
		ids:  make(map[string]*Record),
		tids: make(map[string]*taxon),
	}
	file := filepath.Join(path, recDir, recTaxList)
	f, err := os.Open(file)
	if err != nil {
		return db, nil
	}
	defer f.Close()
	if err := db.readTaxList(f); err != nil {
		return nil, errors.Wrap(err, "records: open: when reading taxon list")
	}

	for _, tax := range db.tids {
		sc := OpenScanner(path, tax.id)
		if err := db.scan(sc); err != nil {
			return nil, err
		}
		tax.changed = false
		tax.sorted = true
	}
	db.changed = false
	return db, nil
}

// ReadTaxList reads taxon names from a file.
func (db *DB) readTaxList(r io.Reader) error {
	s := bufio.NewScanner(r)
	for s.Scan() {
		name := biodv.TaxCanon(s.Text())
		if name == "" {
			continue
		}
		if nm, _ := utf8.DecodeRuneInString(name); nm == '#' || nm == ';' || !unicode.IsLetter(nm) {
			continue
		}
		if _, dup := db.tids[name]; dup {
			continue
		}
		tax := &taxon{id: name, db: db}
		db.tids[name] = tax
	}
	if err := s.Err(); err != nil {
		return err
	}
	db.changed = true
	return nil
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
		rec.SetGeoRef(r.GeoRef())
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
	taxID = biodv.TaxCanon(taxID)
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
	tax.sorted = false
	return rec, nil
}

// Commit saves a record database to hard disk.
func (db *DB) Commit() error {
	if db.changed {
		if err := db.saveTaxList(); err != nil {
			return err
		}
	}

	for _, tax := range db.tids {
		if err := tax.commit(); err != nil {
			return errors.Wrap(err, "records: db: commit")
		}
	}
	db.changed = false
	return nil
}

func (db *DB) saveTaxList() (err error) {
	if _, err := os.Lstat(filepath.Join(db.path, recDir)); err != nil {
		if err := os.Mkdir(filepath.Join(db.path, recDir), os.ModeDir|os.ModePerm); err != nil {
			return errors.Wrapf(err, "records: db: commit: unable to create %s directory", recDir)
		}
	}

	file := filepath.Join(db.path, recDir, recTaxList)
	var f *os.File
	f, err = os.Create(file)
	if err != nil {
		return errors.Wrap(err, "records: db: commit")
	}
	defer func() {
		e1 := f.Close()
		if err == nil && e1 != nil {
			err = errors.Wrap(e1, "records: db: commit")
		}
	}()

	ls := make([]string, 0, len(db.tids))
	for _, tax := range db.tids {
		if len(tax.recs) == 0 {
			continue
		}
		ls = append(ls, tax.id)
	}
	sort.Strings(ls)

	for _, v := range ls {
		fmt.Fprintf(f, "%s\n", v)
	}
	return nil
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
