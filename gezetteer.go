// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package biodv

import (
	"io"
	"sort"
	"sync"

	"github.com/js-arias/biodv/geography"

	"github.com/pkg/errors"
)

// A Gazetteer is a georeferencing service.
type Gazetteer interface {
	// Locate returns a set of points
	// for a given locality.
	Locate(adm geography.Admin, locality string) *GeoScan

	// Reverse returns the administrative data
	// for a given point.
	// It is not implemented by all gazetteers.
	Reverse(p geography.Position) (geography.Admin, error)
}

// GzDriver contains components
// of a Gazetteer driver.
type GzDriver struct {
	// Open is a function to open
	// a Gazetteer.
	Open func(string) (Gazetteer, error)

	// About is a function that return
	// a short description of the driver.
	About func() string
}

var (
	gzDriversMu sync.RWMutex
	gzDrivers   = make(map[string]GzDriver)
)

// RegisterGz makes a Gazetteer driver
// available by the provided name.
// If Register is called twice with the same name
// or if the driver is nil,
// it panics.
func RegisterGz(name string, driver GzDriver) {
	gzDriversMu.Lock()
	defer gzDriversMu.Unlock()
	if driver.Open == nil {
		panic("biodv: GzDriver driver Open is nil")
	}
	if _, dup := gzDrivers[name]; dup {
		panic("biodv: GzDriver called twice for driver " + name)
	}
	gzDrivers[name] = driver
}

// GzDrivers returns a sorted list
// of names of the registered gazetteer drivers.
func GzDrivers() []string {
	gzDriversMu.Lock()
	defer gzDriversMu.Unlock()
	var ls []string
	for name := range gzDrivers {
		ls = append(ls, name)
	}
	sort.Strings(ls)
	return ls
}

// OpenGz opens a Gazetteer
// by its driver,
// and a driver specific parameter string.
func OpenGz(driver, param string) (Gazetteer, error) {
	if driver == "" {
		return nil, errors.New("biodv: empty gazetteer driver")
	}
	gzDriversMu.Lock()
	dr, ok := gzDrivers[driver]
	gzDriversMu.Unlock()
	if !ok {
		return nil, errors.Errorf("biodv: unknown gazetteer driver %q", driver)
	}
	return dr.Open(param)
}

// GzAbout returns a short message
// describing the driver.
func GzAbout(driver string) string {
	if driver == "" {
		return ""
	}
	gzDriversMu.Lock()
	dr, ok := gzDrivers[driver]
	gzDriversMu.Unlock()
	if !ok {
		return ""
	}
	if dr.About == nil {
		return ""
	}
	return dr.About()
}

// A GeoScan is a georeferenced position scanner
// to stream the results of a query
// that is expected to produce
// a list of georeferenced positions.
//
// Use Scan to advance the stream:
//
//	sc := gz.Locate(adm, "las pavas")
//	for sc.Scan() {
//		p := sc.Position()
//		...
//	}
//	if err := sc.Err(); err != nil {
//		...	// process the error
//	}
type GeoScan struct {
	// the position channel
	c chan geography.Position

	// an error encountered during iteration
	err error

	// closed is true if the scanner is closed
	closed bool

	// pos in the last read position
	pos geography.Position
}

// NewGeoScan creates a georeferenced position scanner,
// with a buffer of the indicated size.
func NewGeoScan(sz int) *GeoScan {
	if sz < 10 {
		sz = 10
	}
	return &GeoScan{
		c:   make(chan geography.Position, sz),
		pos: geography.NewPosition(),
	}
}

// Add adds a position or an error
// to a GeoScan.
// It should be used by clients that
// return the scanner.
//
// It returns true,
// if the element is added successfully.
func (gsc *GeoScan) Add(p geography.Position, err error) bool {
	if gsc.err != nil {
		return false
	}
	if gsc.closed {
		return false
	}
	if err != nil {
		gsc.c <- geography.NewPosition()
		close(gsc.c)
		gsc.err = err
		return true
	}
	gsc.c <- p
	return true
}

// Close closes the scanner.
// If Scan is called and returns false
// the scanner is closed automatically.
func (gsc *GeoScan) Close() {
	if gsc.closed {
		return
	}
	if gsc.err != nil {
		return
	}
	close(gsc.c)
	gsc.closed = true
}

// Err returns the error,
// if any,
// that was encountered during iteration.
func (gsc *GeoScan) Err() error {
	if !gsc.closed {
		return nil
	}
	if errors.Cause(gsc.err) == io.EOF {
		return nil
	}
	return gsc.err
}

// Position returns the last read position.
// Every call to Position must be preceded
// by a call to Scan.
func (gsc *GeoScan) Position() geography.Position {
	if gsc.closed {
		panic("biodv: accessing a closed position scanner")
	}
	p := gsc.pos
	gsc.pos = geography.Position{}
	if !p.IsValid() {
		panic("biodv: calling Position without an Scan call")
	}
	return p
}

// Scan advances the scanner to the next result.
// It returns false when there is no more positions,
// or an error happens when preparing it.
// Err should be consulted to distinguish
// between the two cases.
//
// Every call to Position,
// even the first one,
// must be precede by a call to Scan.
func (gsc *GeoScan) Scan() bool {
	if gsc.closed {
		return false
	}
	gsc.pos = <-gsc.c
	if !gsc.pos.IsValid() {
		gsc.closed = true
		if gsc.err == nil {
			gsc.err = io.EOF
			close(gsc.c)
		}
		return false
	}
	return true
}
