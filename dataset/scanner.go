// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package dataset

import (
	"io"
	"sort"
	"strings"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/encoding/stanza"

	"github.com/pkg/errors"
)

// Scanner scans datasets
// from a dataset file,
// in stanza format.
type Scanner struct {
	sc  *stanza.Scanner
	rec map[string]string
	err error
}

// Dataset is an stanza record
// that fullfills the biodv.Dataset interface.
type dataset map[string]string

func (d dataset) ID() string {
	return d[titleKey]
}

func (d dataset) Title() string {
	return d[titleKey]
}

func (d dataset) Keys() []string {
	var ls []string
	mp := make(map[string]bool)
	mp[titleKey] = true
	for k := range d {
		if mp[k] {
			continue
		}
		ls = append(ls, k)
		mp[k] = true
	}
	sort.Strings(ls)
	return ls
}

func (d dataset) Value(key string) string {
	key = strings.ToLower(strings.TrimSpace(key))
	if key == "" {
		return ""
	}
	return d[key]
}

// NewScanner returns a scanner
// that reads from r.
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

// Dataset returns the last read dataset.
// Every call to Dataset must be preceded
// by a call to Scan.
func (sc *Scanner) Dataset() biodv.Dataset {
	if sc.rec == nil {
		panic("dataset: callid Dataset without a Scan call")
	}
	rec := dataset(sc.rec)
	sc.rec = nil
	return rec
}

// Scan advances the scanner to the next dataset.
// It returns false when there is no more datasets,
// or an error happens when preparing it.
// Err should be consulted to distinguish
// between the two cases.
//
// Every call to Dataset,
// even the first one,
// must be preceded a call to Scan.
func (sc *Scanner) Scan() bool {
	if sc.err != nil {
		return false
	}
	for {
		if !sc.sc.Scan() {
			break
		}
		rec := sc.sc.Record()
		if rec[titleKey] == "" {
			sc.err = errors.Errorf("dataset: scanner: empty dataset title")
			return false
		}
		sc.rec = rec
		return true
	}
	if err := sc.sc.Err(); err != nil {
		sc.err = errors.Wrap(err, "dataset: scanner")
		return false
	}
	sc.err = io.EOF
	return false
}
