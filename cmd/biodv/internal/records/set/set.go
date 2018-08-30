// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package set implements the rec.set command,
// i.e. set an specimen record value.
package set

import (
	"strconv"
	"strings"
	"time"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/cmdapp"
	"github.com/js-arias/biodv/geography"
	"github.com/js-arias/biodv/records"

	"github.com/pkg/errors"
)

var cmd = &cmdapp.Command{
	UsageLine: "rec.set -k|--key <key> [-v|--value <value>] <record>",
	Short:     "set an specimen record value",
	Long: `
Command rec.set sets the value of a given key for the indicated record,
overwritting any previous value. If the value is empty, the content of
the key will be eliminated.

Command rec.set can be used to set almost any key value, except taxon
assignation, and the geographic point of the record,

Except for some standard keys, no content of the values will be evaluated
by the program or the database.

Options are:

    - k <key>
    --key <key>
      A key, a required parameter. Keys must be in lower case and
      without spaces  (it will be reformatted to lower case, and spaces
      between words replaced by the dash ‘-’ character). Any key can
      be stored, but the recognized keys (and their expected values)
      are:
        catalog     a catalog code, usually in the form
                    <institution code>:<collection code>:<catalog number>.
        basis       basis of record, it can be:
                      unknown      if the basis is unknown
                      preserved    if it is a preserved (museum)
                                   specimen
                      fossil       if it is a fossil (museum) specimen
                      observation  if the record is based on a human
                                   observation
                      machine      if the record is based on a machine
                                   sensor reading
        date        the sampling date, it must be in the RFC3339 format,
                    e.g. '2006-01-02T15:04:05Z07:00'
        country     the country of the sample, a two letter ISO 3166-1
                    alpha-2 code.
        state       the state, province, or a similar principal country
                    subdivision.
        county      a secondary country subdivision.
        locality    the locality of the sampling.
        collector   the person who collect the sample.
        z           in flying or oceanic specimens, the distance to
                    groud (depth as negative) when the sampling was
                    made.
        elevation   elevation over sea level, in meters
        depth       depth below sea level, in meters
        reference   a bibliographic reference
        dataset     source of the specimen record information
        determiner  the person who identified the specimen
        organism    the organism ID
        stage       the growth stage of the organism
        sex         sex of the organism
        altitude    in flying specimens, the altitude above ground when
                    the observation was made
      For a set of available keys of a given specimen, use rec.value.

    -v <value>
    --value <value>
      The value to set. If no value is defined, or an empty string is
      given, the value on that key will be deleted.

    <record>
      The record to be set.
	`,
	Run:           run,
	RegisterFlags: register,
}

func init() {
	cmdapp.Add(cmd)
}

var key string
var value string

func register(c *cmdapp.Command) {
	c.Flag.StringVar(&key, "key", "", "")
	c.Flag.StringVar(&key, "k", "", "")
	c.Flag.StringVar(&value, "value", "", "")
	c.Flag.StringVar(&value, "v", "", "")
}

func run(c *cmdapp.Command, args []string) error {
	if key == "" {
		return errors.Errorf("%s: a key should be defined", c.Name())
	}
	key = strings.ToLower(key)

	id := strings.Join(args, " ")
	if id == "" {
		return errors.Errorf("%s: a record should be defined", c.Name())
	}

	recs, err := records.Open("")
	if err != nil {
		return errors.Wrap(err, c.Name())
	}
	rec := recs.Record(id)
	if rec == nil {
		return nil
	}
	if err := setRec(rec); err != nil {
		return errors.Wrap(err, c.Name())
	}
	if err := recs.Commit(); err != nil {
		return errors.Wrap(err, c.Name())
	}
	return nil
}

func setRec(rec *records.Record) error {
	geo := rec.GeoRef()
	ev := rec.CollEvent()
	switch key {
	case "date":
		ev.Date = time.Time{}
		if value != "" {
			t, err := time.Parse(time.RFC3339, value)
			if err != nil {
				return errors.Wrap(err, "invalid time value")
			}
			ev.Date = t
		}
		rec.SetCollEvent(ev)
	case "country":
		ev.Admin.Country = ""
		if value != "" {
			if geography.Country(value) == "" {
				return errors.New("invalid country code")
			}
			ev.Admin.Country = strings.ToUpper(value)
		}
		rec.SetCollEvent(ev)
	case "state":
		ev.Admin.State = value
		rec.SetCollEvent(ev)
	case "county":
		ev.Admin.County = value
		rec.SetCollEvent(ev)
	case "locality":
		ev.Locality = value
		rec.SetCollEvent(ev)
	case "collector":
		ev.Collector = value
		rec.SetCollEvent(ev)
	case "z":
		ev.Z = 0
		if value != "" {
			z, err := strconv.Atoi(value)
			if err != nil {
				return errors.Wrap(err, "invalid z value")
			}
			ev.Z = z
		}
		rec.SetCollEvent(ev)
	case "elevation":
		geo.Elevation = 0
		if value != "" {
			elv, err := strconv.Atoi(value)
			if err != nil {
				return errors.Wrap(err, "invalid elevation")
			}
			geo.Elevation = uint(elv)
		}
		rec.SetGeoRef(geo, biodv.GeoPrecision)
	case "depth":
		geo.Depth = 0
		if value != "" {
			depth, err := strconv.Atoi(value)
			if err != nil {
				return errors.Wrap(err, "invalid depth")
			}
			geo.Depth = uint(depth)
		}
		rec.SetGeoRef(geo, biodv.GeoPrecision)
	default:
		return rec.Set(key, value)
	}
	return nil
}
