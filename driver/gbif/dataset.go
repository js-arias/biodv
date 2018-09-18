// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package gbif

import (
	"encoding/json"
	"strings"

	"github.com/js-arias/biodv"

	"github.com/pkg/errors"
)

func init() {
	biodv.RegisterSet("gbif", biodv.SetDriver{OpenSet, SetURL, aboutSetGBIF})
}

// SetURL returns the url of a GBIF dataset.
func SetURL(id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return ""
	}
	return "https://www.gbif.org/dataset/" + id
}

// AboutSetGBIF retuns a simple statement of the purpose of the driver.
func aboutSetGBIF() string {
	return "a driver for the GBIF dataset DB"
}

// OpenSet returns the GBIF
// dataset handler,
// that implements the biodv.SetDB interface.
func OpenSet(param string) (biodv.SetDB, error) {
	if reqChan == nil {
		initReqs()
	}
	return setDB{}, nil
}

// SetDB is the handler of GBIF dataset DB.
type setDB struct{}

// Dataset stores the GBIF dataset information.
// It implements the biodv.Dataset interface.
type dataset struct {
	Key         string // id
	TitleStr    string `json:"title"`
	Description string
	Citation    citation
	Homepage    string
	License     string
	Contacts    []contact
}

type citation struct {
	Text string
}

type contact struct {
	Type         string
	Homepage     []string
	Organization string
}

func (ds *dataset) ID() string {
	return ds.Key
}

func (ds *dataset) Title() string {
	return ds.TitleStr
}

func (ds *dataset) Keys() []string {
	return []string{
		biodv.SetAboutKey,
		biodv.SetRef,
		biodv.SetLicense,
		biodv.SetURLKey,
		biodv.SetPublisher,
	}
}

func (ds *dataset) Value(key string) string {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case biodv.SetAboutKey:
		return ds.Description
	case biodv.SetRef:
		return ds.Citation.Text
	case biodv.SetLicense:
		switch ds.License {
		case "http://creativecommons.org/publicdomain/zero/1.0/legalcode":
			return "CC0-1.0"
		case "http://creativecommons.org/licenses/by/4.0/legalcode":
			return "CC BY 4.0"
		case "http://creativecommons.org/licenses/by-nc/4.0/legalcode":
			return "CC BY-NC 4.0"
		default:
			return "unknown"
		}
	case biodv.SetURLKey:
		if ds.Homepage != "" {
			return ds.Homepage
		}
		if c := ds.bestContact(); len(c.Homepage) > 0 {
			return c.Homepage[0]
		}
		return SetURL(ds.Key)
	case biodv.SetPublisher:
		c := ds.bestContact()
		return c.Organization
	}
	return ""
}

func (ds *dataset) bestContact() contact {
	r := contact{}
	for _, c := range ds.Contacts {
		if c.Type == "ORIGINATOR" {
			r = c
			continue
		}
		if c.Type == "METADATA_AUTHOR" {
			if r.Type == "" {
				r = c
			}
			continue
		}
	}
	return r
}

func (db setDB) SetID(id string) (biodv.Dataset, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.Errorf("gbif: setDB: empty dataset DB")
	}
	var err error
	for r := 0; r < Retry; r++ {
		req := newRequest("dataset/" + id)
		select {
		case err = <-req.err:
			continue
		case a := <-req.ans:
			d := json.NewDecoder(&a)
			ds := &dataset{}
			err = d.Decode(ds)
			if err != nil {
				continue
			}
			return ds, nil
		}
	}
	if err == nil {
		return nil, errors.Errorf("gbif: setDB: no answer after %d retries", Retry)
	}
	return nil, errors.Wrap(err, "gbif: setDB")
}
