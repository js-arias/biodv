// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package biodv

import (
	"sort"
	"sync"

	"github.com/pkg/errors"
)

// SetDriver contains components
// of a SetDB driver.
type SetDriver struct {
	// Open is a function to open
	// a SetDB.
	Open func(string) (SetDB, error)

	// URL is a function to return
	// an URL of a given dataset ID.
	// This value can be nil.
	URL func(id string) string

	// About is a function that return
	// a short description of the driver.
	About func() string
}

var (
	setDriversMu sync.RWMutex
	setDrivers   = make(map[string]SetDriver)
)

// RegisterSet makes a setDB driver
// available by the provided name.
// If Register is called twice with the same name
// or if the driver is nil
// it panics.
func RegisterSet(name string, driver SetDriver) {
	setDriversMu.Lock()
	defer setDriversMu.Unlock()
	if driver.Open == nil {
		panic("biodv: SetDB driver Open is nil")
	}
	if _, dup := setDrivers[name]; dup {
		panic("biodv: RegisterSet called twice for driver " + name)
	}
	setDrivers[name] = driver
}

// SetDrivers returns a sorted list
// of names of the registered drivers.
func SetDrivers() []string {
	setDriversMu.RLock()
	defer setDriversMu.RUnlock()
	var ls []string
	for name := range setDrivers {
		ls = append(ls, name)
	}
	sort.Strings(ls)
	return ls
}

// OpenSet opens a SetDB database
// by its driver,
// and a driver speciefic parameter string.
func OpenSet(driver, param string) (SetDB, error) {
	if driver == "" {
		return nil, errors.New("biodv: empty setDB driver")
	}
	setDriversMu.RLock()
	dr, ok := setDrivers[driver]
	setDriversMu.RUnlock()
	if !ok {
		return nil, errors.Errorf("biodv: unknown setDB driver %q", driver)
	}
	return dr.Open(param)
}

// SetURL returns the URL of a given dataset ID
// in a given database.
func SetURL(driver, id string) string {
	if driver == "" {
		return ""
	}
	setDriversMu.RLock()
	dr, ok := setDrivers[driver]
	setDriversMu.RUnlock()
	if !ok {
		return ""
	}
	if dr.URL == nil {
		return ""
	}
	return dr.URL(id)
}

// SetAbout returns the short message
// describing the driver.
func SetAbout(driver string) string {
	if driver == "" {
		return ""
	}
	setDriversMu.RLock()
	dr, ok := setDrivers[driver]
	setDriversMu.RUnlock()
	if !ok {
		return ""
	}
	if dr.About == nil {
		return ""
	}
	return dr.About()
}
