// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package taxonomy

import (
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/encoding/stanza"

	"github.com/pkg/errors"
)

// Scanner scans taxons
// from a taxonomy file,
// in stanza format.
type Scanner struct {
	f   *os.File
	sc  *stanza.Scanner
	rec map[string]string
	err error
}

// Record is an stanza record
// that fullfills the biodv.Taxon interface.
type record map[string]string

func (r record) Name() string {
	return r[nameKey]
}

func (r record) ID() string {
	return r[nameKey]
}

func (r record) Parent() string {
	return r[parentKey]
}

func (r record) Rank() biodv.Rank {
	return biodv.GetRank(r[rankKey])
}

func (r record) IsCorrect() bool {
	if r[correctKey] == "false" {
		return false
	}
	return true
}

func (r record) Keys() []string {
	var ls []string
	mp := make(map[string]bool)
	mp[nameKey] = true
	mp[parentKey] = true
	mp[rankKey] = true
	mp[correctKey] = true
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

func (r record) Value(key string) string {
	key = strings.ToLower(strings.TrimSpace(key))
	if key == "" {
		return ""
	}
	return r[key]
}

// NewScanner returns a scanner
// that read taxons from r.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{sc: stanza.NewScanner(r)}
}

// OpenScanner returns a scanner
// that reads from a taxonomy file
// on a given path.
func OpenScanner(path string) *Scanner {
	file := filepath.Join(path, taxDir, taxFile)
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

// Taxon returns the last read taxon.
// Every call to Raxon must be preceded
// by a call to Scan.
func (sc *Scanner) Taxon() biodv.Taxon {
	if sc.err != nil {
		panic("taxonomy: scanner: accessing a closed scanner")
	}
	if sc.rec == nil {
		panic("taxonomy: scanner: calling Taxon without a Scan call")
	}
	rec := record(sc.rec)
	sc.rec = nil
	return rec
}

// Scan advances the scanner to the next taxon.
// It returns false when tere is no more taxons,
// or an error happens when preparing it.
// Err should be consulted to distinguish
// between the two cases.
//
// Every call to Taxon,
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
		rec := sc.sc.Record()
		rec[nameKey] = biodv.TaxCanon(rec[nameKey])
		if rec[nameKey] == "" {
			sc.Close()
			sc.err = errors.Errorf("taxonomy: scanner: empty taxon name")
			return false
		}
		rec[parentKey] = biodv.TaxCanon(rec[parentKey])
		rec[rankKey] = biodv.GetRank(rec[rankKey]).String()
		rec[correctKey] = strings.ToLower(strings.TrimSpace(rec[correctKey]))
		if rec[correctKey] != "false" {
			rec[correctKey] = "true"
		}
		if rec[parentKey] == "" && rec[correctKey] == "false" {
			sc.Close()
			sc.err = errors.Errorf("taxonomy: scanner: while parsing %q: synonym without a parent", rec[nameKey])
			return false
		}
		sc.rec = rec
		return true
	}
	if err := sc.sc.Err(); err != nil {
		sc.Close()
		sc.err = errors.Wrap(err, "taxonomy: scanner")
		return false
	}
	sc.Close()
	return false
}
