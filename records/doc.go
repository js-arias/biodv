// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

/*
Package records implements
a records database.

In biodv the specimen records database
is stored in the records sub-directory.
It contains a file called taxons.lst
that is a text file with the name of the taxons with records
(one taxon per line),
and one or more stanza-encoded files
with the taxon’s name that contain
the specimen record data for that taxon.

When assigned with records package,
speciemen records are stored
in the most exclusive taxon available.
For example,
records assigned to Puma concolor,
will be stored on file 'Puma-concolor.stz',
whereas recods assigned to Felis concolor
(a synonym of Puma concolor)
will be stored on file 'Felis-concolor.stz',
and records assigned to Puma concolor cougar
will be stored on file 'Puma-concolor-cougar.stz'.

Instead of accessing record files directly,
use one of these mechanisms:

If you want to read the file sequentially
(for example,
just load the fields that you need
on a particular taxon)
use the Scanner type,
creating it with function OpenScanner
and using the name of the taxon that
you will search.

If you want to use a read-write records database
use DB type
creating it with the Open function.
To access records,
use Record method,
and to access a list of records
use RecList method.
Changes on the Record database
will be kept in memory,
and are only copied to hard disk with an explicit
call to Commit method.

If you want to use a biodv.RecDB,
you can open it using “biodv” as the driver name,
and creating it with the biodv.OpenRec function.

A Scanner can be used also
to read specimen records data from an arbitrary stanza file,
it can be done by creating the Scanner
with the function NewScanner.

The following fields are recognized by the record scanner:

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

If the database is accessed by package records,
it will kept it well formatted.
In case of an untrusted database,
it can be validated by open it with DB type,
that report any violation found while reading the data files.
*/
package records
