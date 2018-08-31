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
	"time"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/geography"

	"github.com/pkg/errors"
)

func init() {
	biodv.RegisterRec("gbif", biodv.RecDriver{OpenRec, RecURL, aboutRecGBIF})
}

// RecURL returns the url of a GBIF taxon.
func RecURL(id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return ""
	}
	return "https://www.gbif.org/occurrence/" + id
}

// AboutRecGBIF returns a simple statement of the purpose of the driver.
func aboutRecGBIF() string {
	return "a driver for the GBIF records DB"
}

// RecUseAll is a parameter used to open a recDB
// that return all records for a given taxon
// including records with unknown source
// and human observations.
const RecUseAll = "use-all"

// RecUseObs is a parameter used to open a recDB
// that returns all records
// based on observations.
const RecUseObs = "use-obs"

// RecSetDataset is a parameter used to open a recDB
// that returns records
// only from the given dataset
// irrespectibely of the basis of each record.
//
// The dataset ID should be given after the constant:
//
// 	param = gbif.RecSetDataset + "83e20573"
//	// param is equal to: "dataset:83e20573"
const RecSetDataset = "dataset:"

// RecOrganism is a parameter used to open a recDB
// that returns records
// of a given specimen.
//
// The specimen ID should be given after the constant:
//
//	param = gbif.RecOrganism + "H903607"
//	// param is equal to "organism:H903607"
const RecOrganism = "organism:"

// OpenRec returns the GBIF
// records handler,
// that implements the biodv.RecDB interface.
//
// By default,
// it will search only preserved specimens.
func OpenRec(param string) (biodv.RecDB, error) {
	if reqChan == nil {
		initReqs()
	}
	db := recDB{param: url.Values{}}
	if i := strings.Index(param, ":"); i > 1 {
		if strings.HasPrefix(param, RecSetDataset) {
			db.param.Add("dataset_key", param[i+1:])
			return db, nil
		}
		if strings.HasPrefix(param, RecOrganism) {
			db.param.Add("organism_id", param[i+1:])
			return db, nil
		}
		db.param.Add("basisOfRecord", "PRESERVED_SPECIMEN")
		db.param.Add("basisOfRecord", "FOSSIL_SPECIMEN")
		return db, nil
	}

	if param == RecUseObs {
		db.param.Add("basisOfRecord", "OBSERVATION")
		db.param.Add("basisOfRecord", "MACHINE_OBSERVATION")
		return db, nil
	}

	if param == RecUseAll {
		return db, nil
	}

	db.param.Add("basisOfRecord", "PRESERVED_SPECIMEN")
	db.param.Add("basisOfRecord", "FOSSIL_SPECIMEN")
	return db, nil
}

// RecDB is the handler of GBIF records DB.
type recDB struct {
	param url.Values
}

// OccAnswer is the answer of the occurrence request.
type occAnswer struct {
	Offset, Limit int64
	EndOfRecords  bool
	Results       []*occurrence
}

// Occurrence stores the GBIF specimen-occurrence information.
// It implements the biodv.Record interface.
type occurrence struct {
	Key            int64 // id
	TaxonKey       int64 // taxon ID
	BasisOfRecord  string
	DatasetKey     string
	IdentifierName string // identified by

	// Catalog
	InstitutionCode string
	CollectionCode  string
	CatalogNumber   string

	// Collection event
	CollectorName                       string
	EventDate                           string
	OccurrenceDate                      string
	CountryCode                         string
	StateProvince                       string
	County                              string
	Locality                            string
	VerbatimLocality                    string
	MinimumDistanceAboveSurfaceInMeters string

	// Georeference
	DecimalLongitude    float64
	DecimalLatitude     float64
	Elevation           float64
	Depth               float64
	GeoreferenceSources string

	// additional data
	OrganismID            string
	Sex                   string
	LifeStage             string
	BibliographicCitation string

	// comments
	FieldNotes        string
	OccurrenceRemarks string

	Issues []string
}

func (occ *occurrence) Taxon() string {
	return strconv.FormatInt(occ.TaxonKey, 10)
}

func (occ *occurrence) ID() string {
	return strconv.FormatInt(occ.Key, 10)
}

func (occ *occurrence) Basis() biodv.BasisOfRecord {
	b := strings.Split(occ.BasisOfRecord, "_")
	return biodv.GetBasis(b[0])
}

func (occ *occurrence) CollEvent() biodv.CollectionEvent {
	t, _ := time.Parse("2006-01-02T15:04:05.000-0700", occ.EventDate)
	if t.IsZero() {
		t, _ = time.Parse("2006-01-02T15:04:05.000-0700", occ.OccurrenceDate)
	}
	adm := geography.Admin{
		Country: occ.CountryCode,
		State:   strings.Join(strings.Fields(occ.StateProvince), " "),
		County:  strings.Join(strings.Fields(occ.County), " "),
	}
	if geography.Country(adm.Country) == "" {
		adm = geography.Admin{}
	}

	cl := biodv.CollectionEvent{
		Date:      t,
		Admin:     adm,
		Locality:  strings.Join(strings.Fields(occ.Locality), " "),
		Collector: strings.Join(strings.Fields(occ.CollectorName), " "),
	}
	if cl.State() == "" && cl.County() == "" {
		cl.Locality = strings.Join(strings.Fields(occ.VerbatimLocality), " ")
	}

	dpt := -int(occ.Depth)
	alt, _ := strconv.Atoi(occ.MinimumDistanceAboveSurfaceInMeters)
	if dpt < 0 && alt == 0 {
		cl.Z = dpt
	}
	if alt > 0 && dpt == 0 {
		cl.Z = alt
	}
	return cl
}

func (occ *occurrence) GeoRef() biodv.Point {
	pt := biodv.InvalidPoint()

	if occ.Elevation > 0 && occ.Depth == 0 {
		pt.Elevation = uint(occ.Elevation)
	}
	if occ.Depth < 0 && occ.Elevation == 0 {
		pt.Depth = uint(occ.Depth)
	}

	if occ.isZero() {
		return pt
	}
	if geography.IsValidCoord(occ.DecimalLatitude, occ.DecimalLongitude) {
		pt.Lat = occ.DecimalLatitude
		pt.Lon = occ.DecimalLongitude
	}
	pt.Source = occ.GeoreferenceSources
	return pt
}

func (occ *occurrence) Keys() []string {
	return []string{
		biodv.RecRef,
		biodv.RecDataset,
		biodv.RecCatalog,
		biodv.RecDeterm,
		biodv.RecComment,
		biodv.RecOrganism,
		biodv.RecSex,
		biodv.RecStage,
	}
}

func (occ *occurrence) Value(key string) string {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case biodv.RecRef:
		return occ.BibliographicCitation
	case biodv.RecDataset:
		return occ.DatasetKey
	case biodv.RecCatalog:
		return occ.catalog()
	case biodv.RecDeterm:
		return occ.IdentifierName
	case biodv.RecComment:
		return strings.TrimSpace(occ.FieldNotes + "\n" + occ.OccurrenceRemarks)
	case biodv.RecOrganism:
		return occ.OrganismID
	case biodv.RecSex:
		return strings.ToLower(occ.Sex)
	case biodv.RecStage:
		return strings.ToLower(occ.LifeStage)
	}
	return ""
}

func (occ *occurrence) catalog() string {
	if occ.CatalogNumber == "" {
		return ""
	}
	if occ.CatalogNumber == "NO DISPONIBLE" {
		return ""
	}
	if occ.InstitutionCode == "" {
		return occ.CatalogNumber
	}
	cat := occ.CatalogNumber
	coll := occ.CollectionCode
	if x, ok := musAcronyms[occ.InstitutionCode]; ok {
		occ.InstitutionCode = x
	}
	inst := occ.InstitutionCode

	if inst == coll {
		coll = ""
	}
	if strings.HasPrefix(cat, occ.CollectionCode) {
		coll = ""
	}
	if strings.HasPrefix(cat, occ.InstitutionCode) {
		inst = ""
		coll = ""
	}
	if coll != "" && strings.HasPrefix(occ.CollectionCode, occ.InstitutionCode) {
		inst = ""
	}

	v := cat
	if coll != "" {
		v = coll + ":" + cat
	}
	if inst != "" {
		v = inst + ":" + v
	}
	return v
}

// IsZero returns true if the "zero coordinate" issue
// is marked for the occurrence,
// or if any of latitude or longitude
// is zero.
func (occ *occurrence) isZero() bool {
	for _, i := range occ.Issues {
		if i == "ZERO_COORDINATE" {
			return true
		}
	}
	if occ.DecimalLatitude == 0 {
		return true
	}
	if occ.DecimalLongitude == 0 {
		return true
	}
	return false
}

func (db recDB) TaxRecs(id string) *biodv.RecScan {
	sc := biodv.NewRecScan(300)
	id = strings.TrimSpace(id)
	if id == "" || id == "0" {
		sc.Add(nil, errors.Errorf("gbif: recDB: invalid taxon ID"))
		return sc
	}
	req := "occurrence/search?"
	param := url.Values{}
	for key, val := range db.param {
		for _, s := range val {
			param.Add(key, s)
		}
	}
	param.Add("taxonKey", id)
	param.Add("limit", "300")
	go db.recordList(sc, req, param)
	return sc
}

// RecordList returns an specific list of records
// with a given set of parameters.
func (db recDB) recordList(sc *biodv.RecScan, reqstr string, param url.Values) {
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
				var resp *occAnswer
				resp, err = decodeRecordList(&a)
				if err != nil {
					continue
				}

				for _, occ := range resp.Results {
					if !sc.Add(occ, nil) {
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
				sc.Add(nil, errors.Errorf("gbif: recDB: no answer after %d retries", Retry))
			}
			sc.Add(nil, errors.Wrap(err, "gbif: recDB"))
			return
		}
	}
	sc.Add(nil, nil)
}

func decodeRecordList(b *bytes.Buffer) (*occAnswer, error) {
	d := json.NewDecoder(b)
	resp := &occAnswer{}
	err := d.Decode(resp)
	return resp, err
}

func (db recDB) RecID(id string) (biodv.Record, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.Errorf("gbif: recDB: empty record ID")
	}
	var err error
	for r := 0; r < Retry; r++ {
		req := newRequest("occurrence/" + id)
		select {
		case err = <-req.err:
			continue
		case a := <-req.ans:
			d := json.NewDecoder(&a)
			occ := &occurrence{}
			err = d.Decode(occ)
			if err != nil {
				continue
			}
			return occ, nil
		}
	}
	if err == nil {
		return nil, errors.Errorf("gbif: recDB: no answer after %d retries", Retry)
	}
	return nil, errors.Wrap(err, "gbif: recDB")
}
