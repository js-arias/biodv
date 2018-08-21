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

// RecUseObs is is a parameter used to open a recDB
// that returns all records
// based on observations.
const RecUseObs = "use-obs"

// RecSetDataset is is a parameter used to open a recDB
// that returns records
// only from the given dataset
// irrespectibely of the basis of each record.
//
// The dataset should be given after the constant:
//
// 	param = gbif.SetDataset + "83e20573"
//	// param is equal to: "dataset:83e20573"
const RecSetDataset = "dataset:"

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
	if i := strings.Index(param, ":"); i == len(RecSetDataset)-1 {
		if param[:i] != RecSetDataset {
			db.param.Add("basisOfRecord", "PRESERVED_SPECIMEN")
			db.param.Add("basisOfRecord", "FOSSIL_SPECIMEN")
			return db, nil
		}
		db.param.Add("dataset_key", param[i+1:])
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
	CollectorName    string
	OccurrenceDate   string
	CountryCode      string
	StateProvince    string
	County           string
	Locality         string
	VerbatimLocality string

	// Georeference
	DecimalLongitude                    float64
	DecimalLatitude                     float64
	MinimumDistanceAboveSurfaceInMeters string
	GeoreferenceSources                 string

	// comments
	FieldNotes        string
	OccurrenceRemarks string

	// additional biologicald data
	OrganismID string
	Sex        string
	LifeStage  string
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
	t, _ := time.Parse("2006-01-02T15:04:05.000-0700", occ.OccurrenceDate)
	cl := biodv.CollectionEvent{
		Date:      t,
		Country:   occ.CountryCode,
		State:     strings.Join(strings.Fields(occ.StateProvince), " "),
		County:    strings.Join(strings.Fields(occ.County), " "),
		Locality:  strings.Join(strings.Fields(occ.Locality), " "),
		Collector: strings.Join(strings.Fields(occ.CollectorName), " "),
	}
	if cl.State == "" && cl.County == "" {
		cl.Locality = strings.Join(strings.Fields(occ.VerbatimLocality), " ")
	}
	return cl
}

func (occ *occurrence) GeoRef() biodv.Point {
	lon, lat := float64(360), float64(360)

	// I put 0 as invalid georef
	// as some systems put undefined georeferences
	// as 0, 0.
	if occ.DecimalLongitude != 0 && occ.DecimalLongitude > biodv.MinLon && occ.DecimalLongitude <= biodv.MaxLon {
		lon = occ.DecimalLongitude
	}

	if occ.DecimalLatitude != 0 && occ.DecimalLatitude > biodv.MinLat && occ.DecimalLatitude <= biodv.MaxLat {
		lat = occ.DecimalLatitude
	}

	alt, _ := strconv.ParseInt(occ.MinimumDistanceAboveSurfaceInMeters, 10, 64)
	return biodv.Point{
		Lon:      lon,
		Lat:      lat,
		Altitude: float64(alt),
		Source:   occ.GeoreferenceSources,
	}
}

func (occ *occurrence) Keys() []string {
	return []string{
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
	case biodv.RecDataset:
		return occ.DatasetKey
	case biodv.RecCatalog:
		return occ.InstitutionCode + ":" + occ.CollectionCode + ":" + occ.CatalogNumber
	case biodv.RecDeterm:
		return occ.IdentifierName
	case biodv.RecComment:
		return occ.FieldNotes + "\n" + occ.OccurrenceRemarks
	case biodv.RecOrganism:
		return occ.OrganismID
	case biodv.RecSex:
		return strings.ToLower(occ.Sex)
	case biodv.RecStage:
		return strings.ToLower(occ.LifeStage)
	}
	return ""
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
