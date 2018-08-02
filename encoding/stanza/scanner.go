// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package stanza reads and writes records
// in a list ('stanza') format.
//
// Stanza files have the following format:
//	1- Each line containing a field
//	   must starts with the field name
//	   and separated from its content by ':' character.
//	   If the field name
//	   ends with a new line rather than ':',
//	   the field is considered as empty.
//	2- Field names are case insensitive
//	   (always read as lower caps),
//	   without spaces,
//	   and should be unique.
//	3- A field ends with a new line.
//	   If the content of the field
//	   extends more than one line,
//	   the next line should start
//	   with at least one space
//	   or tab character.
//	4- A record ends with a line that start with '%' character.
//	   Any character after '%' will be ignored
//	   (Usually "%%" is used to increase visibility of end-of-record).
//	5- Lines starting with '#' are taken as comments.
//	6- Empty lines are ignored.
//
// An example of a stanza list is:
//
//	# Country data facts
//	name:	República Argentina
//	common:	Argentina
//	iso3166: AR
//	capital: Buenos Aires
//	population: 42669500
//	anthem:	Ya su trono dignísimo abrieron
//		las Provincias Unidas del Sud
//		y los libres del mundo responden:
//		"¡Al gran pueblo argentino, salud!"
//	%%
//	name:	대한민국
//	common:	South Korea
//	iso3166: KR
//	capital: Seoul
//	population: 51302044
//	anthem:	무궁화 삼천리 화려강산
//		대한 사람, 대한으로 길이 보전하세
//	%%
//	name:	中华人民共和国
//	common:	China
//	iso3166: CN
//	capital: Beijing
//	population: 1339724852
//	%%
//	name:	Росси́я
//	common:	Russia
//	iso3166: RU
//	capital: Moscow
//	population: 144192450
//	anthem: Славься, Отечество наше свободное,
//		Братских народов союз вековой,
//		Предками данная мудрость народная!
//		Славься, страна! Мы гордимся тобой!
//	%%
//
// Stanza file format is inspired
// by the record-jar/stanza format described
// by E. Raymond "The Art of UNIX programming"
// (2003) Addison-Wesley
// (<http://www.catb.org/esr/writings/taoup/html/ch05s02.html#id2906931>),
// and C. Strozzi NoSQL list format
// (2007)
// (<http://www.strozzi.it/cgi-bin/CSA/tw7/I/en_US/NoSQL/Table%20structure>).
package stanza

import (
	"bufio"
	"io"
	"strings"
	"unicode"

	"github.com/pkg/errors"
)

// A Scanner reads records
// from a stanza-encoded file.
type Scanner struct {
	line   int
	fields []string        // sorted list of fields
	fok    map[string]bool // list of present fields
	r      *bufio.Reader
	b      *strings.Builder
	err    error
	rec    map[string]string
}

// NewScanner returns a new Scanner that reads from r.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		line: 1,
		fok:  make(map[string]bool),
		r:    bufio.NewReader(r),
		b:    &strings.Builder{},
	}
}

// Err returns the error,
// if any,
// that was encountered during iteration.
func (sc *Scanner) Err() error {
	if errors.Cause(sc.err) == io.EOF {
		return nil
	}
	if sc.err != nil {
		return errors.Wrap(sc.err, "stanza")
	}
	return nil
}

// Fields returns a list of all the fields read
// until the last read call.
// The list is sorted by the order of appearance
// in the read file.
// The caller should not modify this slice.
func (sc *Scanner) Fields() []string {
	return sc.fields
}

// Record returns the last read record.
// Every call to Record must be preceded
// by a call to Scan.
func (sc *Scanner) Record() map[string]string {
	if errors.Cause(sc.err) == io.EOF && sc.rec != nil {
		rec := sc.rec
		sc.rec = nil
		return rec
	}
	if sc.err != nil {
		panic("stanza: accessing a closed scanner")
	}
	if sc.rec == nil {
		panic("stanza: calling Record without a Scan call")
	}
	rec := sc.rec
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
// even the first one,
// must be precede by a call to Scan.
func (sc *Scanner) Scan() bool {
	if sc.err != nil {
		return false
	}
	for {
		record, err := sc.parseRecord()
		if err != nil {
			sc.err = errors.Wrap(err, "stanza: Scan")
			return false
		}
		if record == nil {
			continue
		}
		sc.rec = record
		return true
	}
}

// parseRecord parses a single record.
func (sc *Scanner) parseRecord() (record map[string]string, err error) {
	record = make(map[string]string)
	for {
		f, delim, err := sc.parseFieldName()
		if err != nil {
			if len(record) == 0 {
				return nil, err
			}
			break
		}
		if delim == '\n' {
			continue
		}
		if delim == '%' {
			break
		}
		v, end := sc.parseFieldValue()
		if len(f) > 0 && len(v) > 0 {
			if _, dup := record[f]; dup {
				return nil, errors.Errorf("line: %d: duplicated field %q", sc.line, f)
			}
			record[f] = v
			if !sc.fok[f] {
				sc.fok[f] = true
				sc.fields = append(sc.fields, f)
			}
		}
		if end {
			break
		}
	}
	if len(record) == 0 {
		return nil, nil
	}
	return record, nil
}

// parseFieldName parses a field name.
// Delim indicates the character at the
// end of the field name.
func (sc *Scanner) parseFieldName() (field string, delim rune, err error) {
	// setup the reading of a field line:
	// ignores lines starting with comments,
	// and finish if on an end-of-record.
	for {
		r1, err := readRune(sc.r)
		if err != nil {
			return "", 0, err
		}
		if !unicode.IsSpace(r1) {
			if r1 == '#' {
				skip(sc.r, '\n') // skip comments
				sc.line++
				continue
			}
			if r1 == '%' {
				skip(sc.r, '\n') // end-of-record
				sc.line++
				return "", '%', nil
			}
			sc.r.UnreadRune()
			break
		}
		if r1 == '\n' {
			sc.line++
		}
	}

	// reads the field name,
	// stop at a colon (:),
	// or with a new line
	// (interpreted as an empty field).
	sc.b.Reset()
	space := false
	for {
		r1, err := readRune(sc.r)
		if err != nil {
			return "", 0, err
		}
		if r1 == ':' || r1 == '\n' {
			if r1 == '\n' {
				sc.line++
			}
			delim = r1
			break
		}
		if unicode.IsSpace(r1) {
			space = true
			continue
		}
		// replace spaces with '-' character
		if space {
			space = false
			sc.b.WriteRune('-')
		}
		sc.b.WriteRune(r1)
	}
	return strings.ToLower(sc.b.String()), delim, nil
}

// parseFieldValue parses the value of a field.
// End indicates that the end-of-record was found,
// this can be either an explicit end of record
// ('%' character)
// or an error.
func (sc *Scanner) parseFieldValue() (value string, end bool) {
	sc.b.Reset()
	space, first, line := false, true, false
	for {
		r1, err := readRune(sc.r)
		if err != nil {
			end = true
			break
		}

		// check the next line
		if r1 == '\n' {
			sc.line++
			space, line = false, true
			r1, err = readRune(sc.r)
			if err != nil {
				end = true
				break
			}
			if r1 == '#' {
				skip(sc.r, '\n') // skip comments
				sc.line++
				continue
			}
			if r1 == '%' {
				end = true
				skip(sc.r, '\n') // end-of-record
				sc.line++
				break
			}
			if r1 == '\n' {
				sc.r.UnreadRune() // make decision on next loop
				continue
			}
			if unicode.IsSpace(r1) {
				continue // multiline field
			}
			sc.r.UnreadRune() // end-of-field
			break
		}
		if unicode.IsSpace(r1) {
			space = true
			continue
		}
		if line {
			sc.b.WriteRune('\n')
			space = false
			line = false
		}
		if space {
			if !first {
				sc.b.WriteRune(' ')
			}
			space = false
		}
		sc.b.WriteRune(r1)
		first = false
	}
	return sc.b.String(), end
}

// readRune reads a rune,
// folding \r\n to \n.
func readRune(r *bufio.Reader) (rune, error) {
	r1, _, err := r.ReadRune()

	// handle \r\n
	if r1 == '\r' {
		r1, _, err = r.ReadRune()
		if err != nil {
			if r1 != '\n' {
				r.UnreadRune()
				r1 = '\r'
			}
		}
	}
	return r1, err
}

// skip read runes up to and including the rune delim
// or until error.
func skip(r *bufio.Reader, delim rune) error {
	for {
		r1, err := readRune(r)
		if err != nil {
			return err
		}
		if r1 == delim {
			return nil
		}
	}
}
