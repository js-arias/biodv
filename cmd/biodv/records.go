// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package main

import (
	"github.com/js-arias/biodv/cmdapp"

	// initialize records sub-commands
	_ "github.com/js-arias/biodv/cmd/biodv/internal/records/add"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/records/assign"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/records/dbadd"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/records/del"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/records/ed"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/records/georef"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/records/gzgeoref"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/records/info"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/records/mapcmd"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/records/set"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/records/table"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/records/validate"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/records/value"
)

var recHelp = &cmdapp.Command{
	UsageLine: "records",
	Short:     "specimen records database",
	Long: `
In biodv the specimen records database is stored in the records
sub-directory. It contains a file called taxons.lst that is a text file
with the name of the taxons with records (one taxon per line), and one
or more stanza-encoded files with the taxon’s name that contain the
specimen record data for that taxon.

When assigned with a biodv command, specimen records are stored in the
most exclusive taxon available. For example, records assigned to Puma
concolor, will be stored on file 'Puma-concolor.stz', whereas recods
assigned to Felis concolor (a synonym of Puma concolor) will be stored
on file 'Felis-concolor.stz', and records assigned to Puma concolor
cougar will be stored on file 'Puma-concolor-cougar.stz'.

The following fields are recognized by the biodv records commands:

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
    z            in flying or oceanic specimens, the distance to groud
                 (depth as negative) when the sampling was made.
    latitude     geographic latitude of the record.
    longitude    geographic longitude of the record.
    geosource    source of the georeference.
    validation   validation of the georeference.
    uncertainty  georeference uncertainty in meters.
    elevation    elevation over sea level, in meters.
    reference    a bibliographic reference.
    dataset      source of the specimen record information.
    determiner   the person who identified the specimen.
    organism     the organism ID.
    stage        the growth stage of the organism.
    sex          sex of the organism.
    altitude     in flying specimens, the altitude above ground when
                 the observation was made.

Most biodv commands assume that the specimen records datafiles are well
formatted. In the case of an untrusted database, it can be validated with
the command rec.validate.
	`,
}

func init() {
	cmdapp.Add(recHelp)
}
