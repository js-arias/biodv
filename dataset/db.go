// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package dataset implements
// a database of dataset metadata.
package dataset

import (
	"sort"
	"strings"

	"github.com/js-arias/biodv"

	"github.com/pkg/errors"
)

// Default database directory and file.
const setDir = "sources"
const setFile = "dataset.stz"

// Basic keys for dataset database.
const (
	titleKey = "title"
)

// DB is a dataset metadata database,
// for reading and writing data.
// DB implements the biodv.SetDB interface.
type DB struct {
	path    string
	ids     map[string]*Dataset
	changed bool // true if the database was modified
}

// SetID returns a dataset with a given ID.
func (db *DB) SetID(id string) (biodv.Dataset, error) {
	id = strings.Join(strings.Fields(id), " ")
	if id == "" {
		return nil, errors.New("dataset: db: set: empty set ID")
	}
	if set, ok := db.ids[id]; ok {
		return set, nil
	}
	return nil, nil
}

// Dataset is a dataset metadata stored in a DB.
// Dataset implements the biodv.Dataset interface.
type Dataset struct {
	db   *DB
	data map[string]string
}

// ID returns the ID of the dataset
func (set *Dataset) ID() string {
	return dataset(set.data).ID()
}

// Title returns the title of the dataset.
func (set *Dataset) Title() string {
	return dataset(set.data).Title()
}

// Keys return the list of additional fields
// stored in the dataset metadata.
func (set *Dataset) Keys() []string {
	return dataset(set.data).Keys()
}

// Value returns the value
// of the indicated key stored in the dataset metadata.
func (set *Dataset) Value(key string) string {
	return dataset(set.data).Value(key)
}

// Set sets a value from a given key.
// The value should be transformed into
// a string
// When an empty string is used as value,
// the stored value will be deleted.
func (set *Dataset) Set(key, value string) error {
	key = strings.ToLower(strings.Join(strings.Fields(key), "-"))
	if key == "" {
		return nil
	}
	value = strings.TrimSpace(value)

	switch key {
	case titleKey:
		return errors.Errorf("dataset: set: invalid key: %s", key)
	case biodv.SetExtern:
		if value == "" {
			return nil
		}
		srv := getService(value)
		if srv == "" {
			return errors.New("dataset: set: invalid extern ID value")
		}
		ext := strings.Fields(set.data[key])
		if srv+":" == value {
			// empty extern ID
			// deletes the extern ID from database
			for i, e := range ext {
				if srv != getService(e) {
					continue
				}
				delete(set.db.ids, e)
				n := append(ext[:i], ext[i+1:]...)
				set.data[key] = strings.Join(n, " ")
				set.db.changed = true
				return nil
			}
			return nil
		}

		// check if the given ID is already in use
		if _, dup := set.db.ids[value]; dup {
			return errors.Errorf("dataset: set: extern ID %q already in use", value)
		}

		// if the service is already set
		for i, e := range ext {
			if srv != getService(e) {
				continue
			}
			delete(set.db.ids, e)
			set.db.ids[value] = set
			ext[i] = value
			set.data[key] = strings.Join(ext, " ")
			set.db.changed = true
			return nil
		}

		// the service is new
		ext = append(ext, value)
		sort.Strings(ext)
		set.db.ids[value] = set
		set.data[key] = strings.Join(ext, " ")
		set.db.changed = true
		return nil
	default:
		v := set.data[key]
		if v == value {
			return nil
		}
		set.data[key] = value
		set.db.changed = true
	}
	return nil
}

// GetService returns the service
// (extern Dataset identifier)
// that provides an external ID.
func getService(id string) string {
	i := strings.Index(id, ":")
	if i <= 0 {
		return ""
	}
	return id[:i]
}

func init() {
	biodv.RegisterSet("biodv", biodv.SetDriver{open, nil, aboutBiodv})
}

// AboutBiodv returns a simple statement of the purpouse of the driver.
func aboutBiodv() string {
	return "the default dataset driver"
}

// Open opens a DB
// as a biodv.Dataset.
func open(path string) (biodv.SetDB, error) {
	return Open(path)
}

// Open opens a DB
// on a given path.
func Open(path string) (*DB, error) {
	db := &DB{
		path: path,
		ids:  make(map[string]*Dataset),
	}
	if err := db.scan(OpenScanner(path)); err != nil {
		return nil, err
	}
	db.changed = false
	return db, nil
}

// Scan uses a scanner
// to load a database.
func (db *DB) scan(sc *Scanner) error {
	for sc.Scan() {
		r := sc.Dataset()
		set, err := db.Add(r.Title())
		if err != nil {
			sc.Close()
			return err
		}
		keys := r.Keys()
		for _, k := range keys {
			if err := set.Set(k, r.Value(k)); err != nil {
				return err
			}
		}
	}
	return sc.Err()
}

// Add adds a new dataset metadata to the DB.
func (db *DB) Add(title string) (*Dataset, error) {
	title = strings.Join(strings.Fields(title), " ")
	if title == "" {
		return nil, errors.New("dataset: db: add: empty dataset name")
	}
	if _, dup := db.ids[title]; dup {
		return nil, errors.New("dataset: db: add %q: dataset already in database")
	}
	set := &Dataset{
		db:   db,
		data: make(map[string]string),
	}
	set.data[titleKey] = title
	db.ids[title] = set
	db.changed = true
	return set, nil
}
