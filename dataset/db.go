// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package dataset implements
// a database of dataset metadata.
package dataset

import (
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
