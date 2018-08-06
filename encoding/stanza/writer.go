// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package stanza

import (
	"bufio"
	"io"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

// A Writer writes records to a stanza encoded file.
//
// If fields are defined,
// only the indicated fields,
// in the given order will be written.
// This is important if the output file
// should have fields on a given order
// (e.g. for comparison with 'diff').
//
// By default only fields with some content will be printed.
// If ForceEmpty is true,
// then fields without content will be also printed.
type Writer struct {
	ForceEmpty bool // write empty fields
	fields     []string
	w          *bufio.Writer
	fc         int // field count (used in writing)
}

// NewWriter returns a new Writer that writes to w.
func NewWriter(w io.Writer) *Writer {
	return &Writer{w: bufio.NewWriter(w)}
}

// SetFields sets the fields to be written.
// The fields must be unique,
// in lower cases
// and without spaces.
// In the following records,
// the fields will be printed
// in the order of the input slice.
func (w *Writer) SetFields(fields []string) error {
	ok := make(map[string]bool)
	var nf []string
	for _, f := range fields {
		cp := strings.ToLower(strings.Join(strings.Fields(f), "-"))
		if len(cp) == 0 {
			continue
		}
		if cp != f {
			return errors.Errorf("stanza: SetFields: field %q is not valid", f)
		}
		if ok[f] {
			continue
		}
		ok[f] = true
		nf = append(nf, f)
	}
	w.fields = nf
	return nil
}

// Fields returns the fields to be written
// in the writing order.
// The caller should not modify this slice.
func (w *Writer) Fields() []string {
	return w.fields
}

// Flush writes any bufferend data to the underlying io.Writer.
func (w *Writer) Flush() error {
	w.w.Flush()
	if _, err := w.w.Write(nil); err != nil {
		return errors.Wrap(err, "stanza: Flush")
	}
	return nil
}

// Write writes a single record to w.
// A record is a map in which each entry
// represents the content of the field indicated by the key.
func (w *Writer) Write(record map[string]string) error {
	w.fc = 0
	fields := w.fields
	if len(fields) == 0 {
		fields = w.fieldList(record)
	}
	for _, f := range fields {
		if err := w.writeField(f, record[f]); err != nil {
			return errors.Wrap(err, "stanza: Write: writing record")
		}
	}
	if w.fc > 0 {
		if _, err := w.w.WriteString("%%\r\n"); err != nil {
			return errors.Wrap(err, "stanza: Write: writing end-of-record")
		}
	}
	return nil
}

// FieldList returns a list of fields
// in alphabetical order.
func (w *Writer) fieldList(rec map[string]string) []string {
	var fields []string
	ok := make(map[string]bool)
	for f := range rec {
		f = strings.ToLower(strings.Join(strings.Fields(f), "-"))
		if len(f) == 0 {
			continue
		}
		if ok[f] {
			continue
		}
		ok[f] = true
		fields = append(fields, f)
	}
	sort.Strings(fields)
	return fields
}

// WriteField writes a field into a file.
func (w *Writer) writeField(f, v string) (err error) {
	v = strings.TrimSpace(v)
	if len(v) == 0 {
		if !w.ForceEmpty {
			return nil
		}
		if _, err = w.w.WriteString(f + "\r\n"); err != nil {
			return err
		}
		w.fc++
		return nil
	}
	if len(f) < 6 {
		_, err = w.w.WriteString(f + ":\t")
	} else {
		_, err = w.w.WriteString(f + ": ")
	}
	if err != nil {
		return err
	}

	for _, r1 := range v {
		switch r1 {
		case '\r':
		case '\n':
			_, err = w.w.WriteString("\r\n\t")
		default:
			_, err = w.w.WriteRune(r1)
		}
	}
	if _, err = w.w.WriteString("\r\n"); err != nil {
		return err
	}
	w.fc++
	return nil
}
