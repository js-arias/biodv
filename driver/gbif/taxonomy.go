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
	biodv.RegisterTax("gbif", biodv.TaxDriver{OpenTax, TaxURL, aboutTaxGBIF})
}

// TaxURL returns the url of a GBIF taxon.
func TaxURL(id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return ""
	}
	return "https://www.gbif.org/species/" + id
}

// TaxNoNub0 is a parameter used to open a taxonomy
// that skip taxons with nubKey = 0,
// by default,
// when nub is 0,
// it will use the key field instead of the nubKey.
const TaxNoNub0 = "no-nub-0"

// OpenTax returns the GBIF
// taxonomy handler,
// that implements the biodv.Taxonomy interface.
//
// If the param is equal to TaxNoNub0
// it will skip taxons with a nubKey = 0.
func OpenTax(param string) (biodv.Taxonomy, error) {
	if reqChan == nil {
		initReqs()
	}
	db := taxDB{useNub0: true}
	if param == TaxNoNub0 {
		db.useNub0 = false
	}
	return db, nil
}

// TaxDB is the handler of GBIF taxonomy DB.
type taxDB struct {
	useNub0 bool
}

// AboutTaxGBIF returns a simple statement of the purpose of the driver.
func aboutTaxGBIF() string {
	return "a driver for the GBIF taxonomy DB"
}

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

	Issues []string
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

// NoMatch search for the not match error
// on a taxon name.
func (sp *species) noMatch() bool {
	for _, i := range sp.Issues {
		if i == "BACKBONE_MATCH_NONE" {
			return true
		}
	}
	return false
}

func (db taxDB) Taxon(name string) *biodv.TaxScan {
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
	go db.taxonList(sc, req, param)
	return sc
}

func (db taxDB) Children(id string) *biodv.TaxScan {
	sc := biodv.NewTaxScan(300)
	id = strings.TrimSpace(id)
	if id == "" || id == "0" {
		go db.rootTaxons(sc)
		return sc
	}
	req := "species/" + id + "/children?"
	param := url.Values{}
	param.Add("limit", "300")
	go db.taxonList(sc, req, param)
	return sc
}

func (db taxDB) Synonyms(id string) *biodv.TaxScan {
	sc := biodv.NewTaxScan(300)
	id = strings.TrimSpace(id)
	if id == "" || id == "0" {
		sc.Add(nil, errors.Errorf("gbif: taxonomy: invalid ID for synonyms"))
		return sc
	}
	req := "species/" + id + "/synonyms?"
	param := url.Values{}
	param.Add("limit", "300")
	go db.taxonList(sc, req, param)
	return sc
}

// TaxonList returns an specific list of taxons
// with a given set of parameters.
func (db taxDB) taxonList(sc *biodv.TaxScan, reqstr string, param url.Values) {
	var err error
	// nubs store all the nubs found
	nubs := make(map[int64]bool)

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
					// skip taxons with no match
					// to the GBIF backbone
					if sp.noMatch() {
						continue
					}
					nub := sp.NubKey
					if nub == 0 && db.useNub0 {
						nub = sp.Key
					}
					if sp.Key != nub {
						if !nubs[nub] {
							nubs[nub] = false
						}
						continue
					}
					nubs[nub] = true
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

		// Add taxons if the nun taxon was found,
		// but never included
		// (for example,
		// when searching an orthographic variant)
		for id, ok := range nubs {
			if ok {
				continue
			}
			sp, err := db.TaxID(strconv.FormatInt(id, 10))
			if err != nil {
				sc.Add(nil, err)
				return
			}
			sc.Add(sp, nil)
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

func (db taxDB) rootTaxons(sc *biodv.TaxScan) {
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

func (db taxDB) TaxID(id string) (biodv.Taxon, error) {
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
