// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package records

import (
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/encoding/stanza"

	"github.com/pkg/errors"
)

// Scanner scans records
// from a records file,
// in stanza format.
type Scanner struct {
	f   *os.File
	sc  *stanza.Scanner
	rec map[string]string
	err error
}

// Recmap is an stanza record
// that fullfills the biodv.Record interface.
type recmap map[string]string

func (r recmap) Taxon() string {
	return r[taxonKey]
}

func (r recmap) ID() string {
	return r[idKey]
}

func (r recmap) Basis() biodv.BasisOfRecord {
	return biodv.GetBasis(r[basisKey])
}

func (r recmap) CollEvent() biodv.CollectionEvent {
	t, _ := time.Parse(time.RFC3339, r[dateKey])
	return biodv.CollectionEvent{
		Date:      t,
		Country:   r[countryKey],
		State:     r[stateKey],
		County:    r[countyKey],
		Locality:  r[localityKey],
		Collector: r[collectorKey],
	}
}

func (r recmap) GeoRef() biodv.Point {
	p := biodv.InvalidPoint()
	alt, _ := strconv.ParseFloat(r[altitudeKey], 64)
	dep, _ := strconv.ParseFloat(r[depthKey], 64)
	p.Altitude = alt
	p.Depth = dep

	v := strings.Fields(r[latlonKey])
	if len(v) != 2 {
		return p
	}
	lat, _ := strconv.ParseFloat(v[0], 64)
	lon, _ := strconv.ParseFloat(v[1], 64)
	p.Lat = lat
	p.Lon = lon
	p.Source = r[geosourceKey]
	p.Validation = r[validationKey]

	un, _ := strconv.Atoi(r[uncertaintyKey])
	p.Uncertainty = uint(un)
	return p
}

func (r recmap) Keys() []string {
	var ls []string
	mp := map[string]bool{
		taxonKey:       true,
		idKey:          true,
		basisKey:       true,
		dateKey:        true,
		countryKey:     true,
		stateKey:       true,
		countyKey:      true,
		localityKey:    true,
		collectorKey:   true,
		latlonKey:      true,
		uncertaintyKey: true,
		altitudeKey:    true,
		depthKey:       true,
		geosourceKey:   true,
		validationKey:  true,
	}
	for k := range r {
		if mp[k] {
			continue
		}
		ls = append(ls, k)
		mp[k] = true
	}
	sort.Strings(ls)
	return ls
}

func (r recmap) Value(key string) string {
	key = strings.ToLower(strings.TrimSpace(key))
	if key == "" {
		return ""
	}
	return r[key]
}

// NewScanner returns a scanner
// that read records from r.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{sc: stanza.NewScanner(r)}
}

// OpenScanner returns a scanner
// that reads from a taxon-records file
// on a given path.
func OpenScanner(path, taxon string) *Scanner {
	taxFile := taxFileName(taxon)
	if taxFile == "" {
		return &Scanner{err: errors.New("records: scanner: undefined taxon name")}
	}
	file := filepath.Join(path, recDir, taxFile)
	f, err := os.Open(file)
	if err != nil {
		return &Scanner{err: io.EOF}
	}
	return &Scanner{
		sc: stanza.NewScanner(f),
		f:  f,
	}
}

// Close closes the scanner,
// preventing furher enumeration.
//
// If Scan returns false,
// the scanner is closed automatically
// and it will suffice to check the result of Err.
func (sc *Scanner) Close() {
	if sc.err == io.EOF {
		return
	}
	if sc.f != nil {
		sc.f.Close()
	}
	sc.err = io.EOF
}

// Err returns the error,
// if any,
// that was encountered during iteration.
func (sc *Scanner) Err() error {
	if sc.err == io.EOF {
		return nil
	}
	return sc.err
}

// Record returns the last read record.
// Every call to Record must be preceded
// by a call to Scan.
func (sc *Scanner) Record() biodv.Record {
	if sc.err != nil {
		panic("records: scanner: accessing a closed scanner")
	}
	if sc.rec == nil {
		panic("records: scanner: calling Record without a Scan call")
	}
	rec := recmap(sc.rec)
	sc.rec = nil
	return rec
}

// Scan advances the scanner to the next record.
// It returns false when there is no more records,
// or an error happens when preparing it.
// Err should be consulted to distinguish
// between the two cases.
//
// Every call to Record,
// event the first one,
// must be preceded by a call to Scan.
func (sc *Scanner) Scan() bool {
	if sc.err != nil {
		return false
	}
	for {
		if !sc.sc.Scan() {
			break
		}
		sc.rec = sc.sc.Record()
		return true
	}
	if err := sc.sc.Err(); err != nil {
		sc.Close()
		sc.err = errors.Wrap(err, "records: scanner")
		return false
	}
	sc.Close()
	return false
}
