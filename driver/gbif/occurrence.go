// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package gbif

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/js-arias/biodv"
)

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
	IdentifierName string // indetifiedby

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

func decodeRecordList(b *bytes.Buffer) (*occAnswer, error) {
	d := json.NewDecoder(b)
	resp := &occAnswer{}
	err := d.Decode(resp)
	return resp, err
}
