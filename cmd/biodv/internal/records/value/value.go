// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package value implements the rec.value command,
// i.e. get an specimen record value.
package value

import (
	"fmt"
	"strings"
	"time"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/cmdapp"

	"github.com/pkg/errors"
)

var cmd = &cmdapp.Command{
	UsageLine: "rec.value [-db <database>] [-k|--key <key>] <record>",
	Short:     "get an specimen record value",
	Long: `
Command rec.value prints the value of a given key for the indicated
specimen record. If no key is given, a list of available keys for the
indicated record will be given.

Options are:

    -db <database>
    --db <database>
      If set, the indicated database will be used to search the
      indicated taxon.
      To see the available databases use the command ‘db.drivers’.
      The default biodv database on the current directory.

    -k <key>
    --key <key>
      If set, the value of the indicated key will be printed.

    <record>
      The ID of the record that will be searched. This parameter is
      required.
	`,
	Run:           run,
	RegisterFlags: register,
}

func init() {
	cmdapp.Add(cmd)
}

var dbName string
var key string

func register(c *cmdapp.Command) {
	c.Flag.StringVar(&dbName, "db", "biodv", "")
	c.Flag.StringVar(&key, "key", "", "")
	c.Flag.StringVar(&key, "k", "", "")
}

func run(c *cmdapp.Command, args []string) error {
	if dbName == "" {
		dbName = "biodv"
	}

	rec, err := getRecord(strings.Join(args, " "))
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	if key == "" {
		ls := []string{
			"taxon",
			"id",
			"basis",
			"date",
			"country",
			"state",
			"county",
			"locality",
			"collector",
			"z",
			"latlon",
			"uncertainty",
			"elevation",
			"geosource",
			"validation",
		}
		ls = append(ls, rec.Keys()...)
		for _, k := range ls {
			fmt.Printf("%s\n", k)
		}
		return nil
	}

	key = strings.ToLower(key)
	geo := rec.GeoRef()
	ev := rec.CollEvent()
	switch key {
	case "taxon":
		fmt.Printf("%s\n", rec.Taxon())
	case "id":
		fmt.Printf("%s\n", rec.ID())
	case "basis":
		fmt.Printf("%s\n", rec.Basis())
	case "date":
		fmt.Printf("%v\n", ev.Date.Format(time.RFC3339))
	case "country":
		fmt.Printf("%s\n", ev.Country())
	case "state":
		fmt.Printf("%s\n", ev.State())
	case "county":
		fmt.Printf("%s\n", ev.County())
	case "locality":
		fmt.Printf("%s\n", ev.Locality)
	case "collector":
		fmt.Printf("%s\n", ev.Collector)
	case "z":
		fmt.Printf("%d\n", ev.Z)
	case "latlon":
		if geo.IsValid() {
			fmt.Printf("%f %f\n", geo.Lat, geo.Lon)
		} else {
			fmt.Printf("NA NA\n")
		}
	case "uncertainty":
		fmt.Printf("%d\n", geo.Uncertainty)
	case "elevation":
		fmt.Printf("%d\n", geo.Elevation)
	case "geosource":
		fmt.Printf("%s\n", geo.Source)
	case "validation":
		fmt.Printf("%s\n", geo.Validation)
	case biodv.RecExtern:
		ext := rec.Value(key)
		for _, e := range strings.Fields(ext) {
			fmt.Printf("%s\n", e)
		}
	default:
		fmt.Printf("%s\n", rec.Value(key))
	}
	return nil
}

// GetRecord returns a record from the options.
func getRecord(id string) (biodv.Record, error) {
	if id == "" {
		return nil, errors.New("a record ID should be given")
	}

	var param string
	dbName, param = biodv.ParseDriverString(dbName)
	recs, err := biodv.OpenRec(dbName, param)
	if err != nil {
		return nil, err
	}
	return recs.RecID(id)
}
