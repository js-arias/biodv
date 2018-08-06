// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package gbif

import (
	"bytes"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"github.com/js-arias/biodv"

	"github.com/pkg/errors"
)

func init() {
	biodv.RegisterTax("gbif", OpenTax)
}

// NubO is a parameter used to open a taxonomy
// that accepts taxons with nubKey = 0,
// in such cases it will use the key field
// instead of the nubKey.
const Nub0 = "nub-0"

// OpenTax returns the GBIF
// taxonomy handler,
// that implements the biodv.Taxonomy interface.
//
// If the param is equal to Nub0
// it will accept taxons with a nubKey = 0.
func OpenTax(param string) (biodv.Taxonomy, error) {
	if reqChan == nil {
		initReqs()
	}
	if param == Nub0 {
		useNub0 = true
	}
	return database{}, nil
}

var useNub0 = false

// SpAnswer is the answer for the species request.
type spAnswer struct {
	Offset, Limit int64
	EndOfRecords  bool
	Results       []*species
}

// Species stores the GBIF taxonomic information.
// It implements the biodv.Taxonomy interface.
type species struct {
	Key, NubKey, AcceptedKey int64  // id
	CanonicalName            string // name
	ScientificName           string // for cases in which canonical is not defined
	Authorship               string // author
	RankStr                  string `json:"rank"`
	Synonym                  bool   // correct
	ConstituentKey           string // source
	ParentKey                int64  // parent
	PublishedIn              string // reference

	//parents
	KingdomKey int64
	PhylumKey  int64
	ClassKey   int64
	OrderKey   int64
	FamilyKey  int64
	GenusKey   int64

	Kingdom string
	Phylum  string
	Clazz   string
	Order   string
	Family  string
	Genus   string
}

func (sp *species) Name() string {
	if sp.CanonicalName == "" {
		return sp.ScientificName
	}
	return sp.CanonicalName
}

func (sp *species) ID() string {
	if sp.NubKey == 0 {
		return strconv.FormatInt(sp.Key, 10)
	}
	return strconv.FormatInt(sp.NubKey, 10)
}

func (sp *species) Parent() string {
	if sp.Synonym {
		return strconv.FormatInt(sp.AcceptedKey, 10)
	}
	if sp.ParentKey == 0 {
		return ""
	}
	return strconv.FormatInt(sp.ParentKey, 10)
}

func (sp *species) Rank() biodv.Rank {
	return biodv.GetRank(sp.RankStr)
}

func (sp *species) IsCorrect() bool {
	return !sp.Synonym
}

func (sp *species) Keys() []string {
	return []string{
		biodv.TaxAuthor,
		biodv.TaxRef,
		biodv.TaxSource,
	}
}

func (sp *species) Value(key string) string {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case biodv.TaxAuthor:
		return sp.Authorship
	case biodv.TaxRef:
		return sp.PublishedIn
	case biodv.TaxSource:
		return sp.ConstituentKey
	}
	return ""
}

func (db database) Taxon(name string) *biodv.TaxScan {
	sc := biodv.NewTaxScan(300)
	name = strings.Join(strings.Fields(name), " ")
	if name == "" {
		sc.Add(nil, errors.Errorf("gbif: taxonomy: empty taxon name"))
		return sc
	}
	req := "species?"
	param := url.Values{}
	param.Add("name", name)
	param.Add("limit", "300")
	go taxonList(sc, req, param)
	return sc
}

func (db database) Children(id string) *biodv.TaxScan {
	sc := biodv.NewTaxScan(300)
	id = strings.TrimSpace(id)
	if id == "" || id == "0" {
		go db.rootTaxons(sc)
		return sc
	}
	req := "species/" + id + "/children?"
	param := url.Values{}
	param.Add("limit", "300")
	go taxonList(sc, req, param)
	return sc
}

func (db database) Synonyms(id string) *biodv.TaxScan {
	sc := biodv.NewTaxScan(300)
	id = strings.TrimSpace(id)
	if id == "" || id == "0" {
		sc.Add(nil, errors.Errorf("gbif: taxonomy: invalid ID for synonyms"))
		return sc
	}
	req := "species/" + id + "/synonyms?"
	param := url.Values{}
	param.Add("limit", "300")
	go taxonList(sc, req, param)
	return sc
}

// TaxonList returns an specific list of taxons
// with a given set of parameters.
func taxonList(sc *biodv.TaxScan, reqstr string, param url.Values) {
	var err error
	end := false
	for off := int64(0); !end; {
		if off > 0 {
			param.Set("offset", strconv.FormatInt(off, 10))
		}
		retryErr := true
		for r := 0; r < Retry; r++ {
			req := newRequest(reqstr + param.Encode())
			select {
			case err = <-req.err:
				continue
			case a := <-req.ans:
				var resp *spAnswer
				resp, err = decodeTaxonList(&a)
				if err != nil {
					continue
				}

				for _, sp := range resp.Results {
					nub := sp.NubKey
					if nub == 0 && useNub0 {
						nub = sp.Key
					}
					if sp.Key != nub {
						continue
					}
					if !sc.Add(sp, nil) {
						return
					}
				}

				// end retry loop
				r = Retry
				retryErr = false
				if resp.EndOfRecords {
					end = true
				}
				off += resp.Limit
			}
		}

		if retryErr {
			if err == nil {
				sc.Add(nil, errors.Errorf("gbif: taxonomy: no answer after %d retries", Retry))
			}
			sc.Add(nil, errors.Wrap(err, "gbif: taxonomy"))
			return
		}
	}
	sc.Add(nil, nil)
}

func decodeTaxonList(b *bytes.Buffer) (*spAnswer, error) {
	d := json.NewDecoder(b)
	resp := &spAnswer{}
	err := d.Decode(resp)
	return resp, err
}

func (db database) rootTaxons(sc *biodv.TaxScan) {
	var kingdoms = []string{
		"1", // Animalia
		"2", // Archaea
		"3", // Bacteria
		"4", // Chromista
		"5", // Fungi
		"6", // Plantae
		"7", // Protozoa
		"8", // Viruses
	}
	for _, id := range kingdoms {
		tax, err := db.TaxID(id)
		if err != nil {
			sc.Add(nil, err)
			break
		}
		sc.Add(tax, nil)
	}
	sc.Add(nil, nil)
}

func (db database) TaxID(id string) (biodv.Taxon, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.Errorf("gbif: taxonomy: empty taxon ID")
	}
	var err error
	for r := 0; r < Retry; r++ {
		req := newRequest("species/" + id)
		select {
		case err = <-req.err:
			continue
		case a := <-req.ans:
			d := json.NewDecoder(&a)
			sp := &species{}
			err = d.Decode(sp)
			if err != nil {
				continue
			}
			return sp, nil
		}
	}
	if err == nil {
		return nil, errors.Errorf("gbif: taxonomy: no answer after %d retries", Retry)
	}
	return nil, errors.Wrap(err, "gbif: taxonomy")
}
