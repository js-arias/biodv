// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package georef implements the rec.georef command,
// i.e. set the georeference of an specimen record.
package georef

import (
	"strings"

	"github.com/js-arias/biodv/cmdapp"
	"github.com/js-arias/biodv/geography"
	"github.com/js-arias/biodv/records"

	"github.com/pkg/errors"
)

var cmd = &cmdapp.Command{
	UsageLine: `rec.georef [-lat|--latitude <value>]
		[-lon|--longitude <value>] [-u|--uncertainty <value>]
		[-e|--elevation <value>] [-s|--source <value>]
		[-v|--validation <value>] [-r|--remove] <record>`,
	Short: "set the georeference of an specimen record",
	Long: `
Command rec.georef sets the georeference of the specified specimen
record. Options are used to set particular values. If they left empty,
they are ignored. -lat or --latitude option and -lon or--longitude
options should be always defined as a pair.

To eliminate a geographic georeference use the option -r or --remove
option.

To eliminate an string value (-s, or --source, and -v, or --validation,
options, use '-' as value. 

Latitude and longitude should be defined using decimal points, and signs
to indicate the hemisphere (negatives for southern and western
hemispheres).

Options are:

    -lat <value>
    --latitude <value>
      Set the latitude of the record, with decimal points. If defined,
      it should be paired with -lon or --longitude option.

    -lon <value>
    --longitude <value>
      Set the longitude of the record, with decimal points. If defined,
      it should be paired with -lat or --latitude option.

    -e <value>
    --elevation <value>
      Elevation above sea level, in meters.

    -u <value>
    --uncertainty <value>
      The uncertainty of the georeference, in meters.

    -s <value>
    --source <value>
      An ID or a description of the source of the georeference, for
      example a GPS device, or a gazetteer service.

    -v <value>
    --validation <value>
      An ID or a description of a validation of the georeference, if
      any.

    -r
    --remove
      If set, the latitude and longitude pair of the record will be removed.

    <record>
      The record to be set.
	`,
	Run:           run,
	RegisterFlags: register,
}

func init() {
	cmdapp.Add(cmd)
}

var lat float64
var lon float64
var elev int
var uncert int
var source string
var valid string
var remov bool

func register(c *cmdapp.Command) {
	c.Flag.Float64Var(&lat, "latitude", geography.MaxLat*2, "")
	c.Flag.Float64Var(&lat, "lat", geography.MaxLat*2, "")
	c.Flag.Float64Var(&lon, "longitude", geography.MaxLat*2, "")
	c.Flag.Float64Var(&lon, "lon", geography.MaxLat*2, "")
	c.Flag.IntVar(&elev, "elevation", -1, "")
	c.Flag.IntVar(&elev, "e", -1, "")
	c.Flag.IntVar(&uncert, "uncertatinty", -1, "")
	c.Flag.IntVar(&uncert, "u", -1, "")
	c.Flag.StringVar(&source, "source", "", "")
	c.Flag.StringVar(&source, "s", "", "")
	c.Flag.StringVar(&valid, "validation", "", "")
	c.Flag.StringVar(&valid, "v", "", "")
	c.Flag.BoolVar(&remov, "remove", false, "")
	c.Flag.BoolVar(&remov, "r", false, "")
}

func run(c *cmdapp.Command, args []string) error {
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
	setRec(rec)
	if err := recs.Commit(); err != nil {
		return errors.Wrap(err, c.Name())
	}
	return nil
}

func setRec(rec *records.Record) {
	if remov {
		geo := geography.NewPosition()
		rec.SetGeoRef(geo)
		return
	}
	geo := rec.GeoRef()
	if geography.IsValidCoord(lat, lon) {
		geo.Lat = lat
		geo.Lon = lon
	}
	if elev >= 0 {
		geo.Elevation = uint(elev)
	}
	if uncert >= 0 {
		geo.Uncertainty = uint(uncert)
	}
	if source == "-" {
		geo.Source = ""
	} else if source != "" {
		geo.Source = source
	}
	if valid == "-" {
		geo.Validation = ""
	} else if valid != "" {
		geo.Validation = valid
	}
	rec.SetGeoRef(geo)
}
