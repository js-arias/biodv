// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package add implements the rec.add command,
// i.e. add specimen records.
package add

import (
	"encoding/csv"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/cmdapp"
	"github.com/js-arias/biodv/records"

	"github.com/pkg/errors"
)

var cmd = &cmdapp.Command{
	UsageLine: "rec.add [<file>..,]",
	Short:     "add specimen records",
	Long: `
Command rec.add adds one or more records from the indicated files, or
the standard input (if no file is defined) to the specimen records
database. It assumes that the input file is a table with tab-delimited
values. Recoginzed column names (and their accepted values) are:

    id           the ID of the record.
    taxon        name (or ID) of the taxon assigned to the specimen.
    catalog      a catalog code, usually in the form
                 <institution code>:<collection code>:<catalog number>.
    basis        basis of record, it can be:
                    unknown      if the basis is unknown
                    preserved    if it is a preserved (museum)
                                 specimen
                    fossil       if it is a fossil (museum) specimen
                    observation  if the record is based on a human
                                 observation
                    machine      if the record is based on a machine
                                 sensor reading
    date         the sampling date, it must be in the RFC3339 format,
                 e.g. '2006-01-02T15:04:05Z07:00'
    country      the country of the sample, a two letter ISO 3166-1
                 alpha-2 code.
    state        the state, province, or a similar principal country
                 subdivision.
    county       a secondary country subdivision.
    locality     the locality of the sampling.
    collector    the person who collect the sample.
    latitude     geographic latitude of the record.
    longitude    geographic longitude of the record.
    geosource    source of the georeference.
    validation   validation of the georeference.
    uncertainty  georeference uncertainty in meters.
    elevation    elevation over sea level, in meters.
    depth        depth below sea level, in meters.
    reference    a bibliographic reference.
    dataset      source of the specimen record information.
    determiner   the person who identified the specimen.
    organism     the organism ID.
    stage        the growth stage of the organism.
    sex          sex of the organism.
    altitude     in flying specimens, the altitude above ground when
                 the observation was made.

If no ID is defined, but a catalog code is given, then the catalog code
will be used as the record ID.

Other values are accepted and stored as given.

Options are:

    <file>
      One or more files to be processed by rec.add. If no file is given,
      the data will be read from the standard input.
	`,
	Run: run,
}

func init() {
	cmdapp.Add(cmd)
}

func run(c *cmdapp.Command, args []string) error {
	recs, err := records.Open("")
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	if len(args) == 0 {
		args = append(args, "-")
	}
	for _, a := range args {
		if a == "-" {
			if err := read(recs, os.Stdin); err != nil {
				return errors.Wrapf(err, "%s: while reading from stdin", c.Name())
			}
			continue
		}
		f, err := os.Open(a)
		if err != nil {
			return errors.Wrapf(err, "%s: unable to open %s", c.Name(), a)
		}
		err = read(recs, f)
		f.Close()
		if err != nil {
			return errors.Wrapf(err, "%s: while reading from %s", c.Name(), a)
		}
	}

	if err := recs.Commit(); err != nil {
		return errors.Wrap(err, c.Name())
	}
	return nil
}

func read(recs *records.DB, in io.Reader) error {
	r := csv.NewReader(in)
	r.Comma = '\t'
	r.Comment = '#'

	// reads the header
	cols := make(map[string]int)

	head, err := r.Read()
	if err != nil {
		return errors.Wrap(err, "while reading header")
	}
	for i, h := range head {
		h = strings.ToLower(h)
		if _, dup := cols[h]; dup {
			return errors.Errorf("column name %q repeated", h)
		}
		cols[h] = i
	}
	if _, ok := cols["taxon"]; !ok {
		return errors.New("a column called 'taxon' must be defined")
	}

	// reads the records
	for i := 1; ; i++ {
		row, err := r.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return errors.Wrapf(err, "on row %d", i)
		}

		pt := biodv.InvalidPoint()
		nm := biodv.TaxCanon(row[cols["taxon"]])
		if nm == "" {
			continue
		}
		id := ""
		if c, ok := cols["id"]; ok {
			id = row[c]
		}
		basis := biodv.UnknownBasis
		if c, ok := cols["basis"]; ok {
			basis = biodv.GetBasis(row[c])
		}
		lt := ""
		if c, ok := cols["latitude"]; ok {
			lt = row[c]
		}
		ln := ""
		if c, ok := cols["longitude"]; ok {
			ln = row[c]
		}
		if lt != "" && ln != "" {
			lat, err := strconv.ParseFloat(lt, 64)
			if err != nil {
				return errors.Wrapf(err, "on row %d, col 'latitude'", i)
			}
			lon, err := strconv.ParseFloat(ln, 64)
			if err != nil {
				return errors.Wrapf(err, "on row %d, col 'longitude'", i)
			}
			pt.Lat = lat
			pt.Lon = lon
		}
		cat := ""
		if c, ok := cols["catalog"]; ok {
			cat = row[c]
		}

		rec, err := recs.Add(nm, id, cat, basis, pt.Lat, pt.Lon)
		if err != nil {
			return errors.Wrapf(err, "on row %d", i)
		}
		ev := biodv.CollectionEvent{}
		for h, c := range cols {
			switch h {
			case "taxon", "id", "catalog", "basis":
			case "date":
				if row[c] != "" {
					ev.Date, _ = time.Parse(time.RFC3339, row[c])
				}
			case "country":
				ev.Country = row[c]
			case "state":
				ev.State = row[c]
			case "county":
				ev.County = row[c]
			case "locality":
				ev.Locality = row[c]
			case "collector":
				ev.Collector = row[c]
			case "latitude", "longitude":
			case "elevation":
				elv, _ := strconv.Atoi(row[c])
				pt.Elevation = uint(elv)
			case "depth":
				dp, _ := strconv.Atoi(row[c])
				pt.Depth = uint(dp)
			case "uncertainty":
				un, _ := strconv.Atoi(row[c])
				pt.Uncertainty = uint(un)
			case "geosource":
				pt.Source = row[c]
			case "validation":
				pt.Validation = row[c]
			default:
				if err := rec.Set(h, row[c]); err != nil {
					return errors.Wrapf(err, "on row %d, col '%s'", i, h)
				}
			}
		}
		rec.SetCollEvent(ev)
		rec.SetGeoRef(pt, biodv.GeoPrecision)
	}
}
