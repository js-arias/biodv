// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package taxonomy

import (
	"io"
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
	sc  *stanza.Scanner
	rec map[string]string
	err error
}

// Record is an stanza record
// that fullfills the biodv.Taxon interface.
type record map[string]string

func (r record) Name() string {
	return r["name"]
}

func (r record) ID() string {
	return r["name"]
}

func (r record) Parent() string {
	return r["parent"]
}

func (r record) Rank() biodv.Rank {
	return biodv.GetRank(r["rank"])
}

func (r record) IsCorrect() bool {
	if r["correct"] == "false" {
		return false
	}
	return true
}

func (r record) Keys() []string {
	var ls []string
	mp := make(map[string]bool)
	mp["name"] = true
	mp["parent"] = true
	mp["rank"] = true
	mp["correct"] = true
	for k := range r {
		if mp[k] {
			continue
		}
		ls = append(ls, k)
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
		rec["name"] = biodv.TaxCanon(rec["name"])
		if rec["name"] == "" {
			sc.err = errors.Errorf("taxonomy: scanner: empty taxon name")
			return false
		}
		rec["parent"] = biodv.TaxCanon(rec["parent"])
		rec["rank"] = biodv.GetRank(rec["rank"]).String()
		rec["correct"] = strings.ToLower(strings.TrimSpace(rec["correct"]))
		if rec["correct"] != "false" {
			rec["correct"] = "true"
		}
		if rec["parent"] == "" && rec["correct"] == "false" {
			sc.err = errors.Errorf("taxonomy: scanner: while parsing %q: synonym without a parent", rec["name"])
			return false
		}
		sc.rec = rec
		return true
	}
	if err := sc.sc.Err(); err != nil {
		sc.err = errors.Wrap(err, "taxonomy: scanner")
		return false
	}
	sc.err = io.EOF
	return false
}
