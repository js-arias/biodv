// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package info implements the rec.info command,
// i.e. print record information.
package info

import (
	"fmt"
	"strings"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/cmdapp"

	"github.com/pkg/errors"
)

var cmd = &cmdapp.Command{
	UsageLine: "rec.info [-db <database>] <value>",
	Short:     "print record information",
	Long: `
Command rec.info prints the information data available for an specimen
record, in a given database.

Options are:

    -db <database>
    --db <database>
      If set, the indicated database will be used to extract the
      specimen record information.
      To see the available databases use the command ‘db.drivers’.

    <value>
      The ID of the specimen record.
	`,
	Run:           run,
	RegisterFlags: register,
}

func init() {
	cmdapp.Add(cmd)
}

var dbName string
var param string

func register(c *cmdapp.Command) {
	c.Flag.StringVar(&dbName, "db", "biodv", "")
}

func run(c *cmdapp.Command, args []string) error {
	if dbName == "" {
		dbName = "biodv"
	}
	dbName, param = biodv.ParseDriverString(dbName)

	if len(args) != 1 {
		return errors.Errorf("%s: a record ID should be given", c.Name())
	}
	id := args[0]

	txm, err := biodv.OpenTax(dbName, param)
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	recs, err := biodv.OpenRec(dbName, param)
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	sp, err := recs.RecID(id)
	if err != nil {
		return errors.Wrap(err, c.Name())
	}
	if err := print(txm, sp); err != nil {
		return errors.Wrap(err, c.Name())
	}
	return nil
}

func print(txm biodv.Taxonomy, rc biodv.Record) error {
	if err := printTaxon(txm, rc); err != nil {
		return err
	}

	ids(rc)

	if org := rc.Value(biodv.RecOrganism); org != "" {
		fmt.Printf("Organism ID: %s\n", org)
	}
	if stg := rc.Value(biodv.RecStage); stg != "" {
		fmt.Printf("Stage: %s\n", stg)
	}
	if sex := rc.Value(biodv.RecSex); sex != "" {
		fmt.Printf("Sex: %s\n", sex)
	}
	if det := rc.Value(biodv.RecDeterm); det != "" {
		fmt.Printf("Determiner: %s\n", det)
	}
	fmt.Printf("Basis of record: %s\n", rc.Basis())
	if set := rc.Value(biodv.RecDataset); set != "" {
		fmt.Printf("Dataset: %s\n", set)
	}
	if ref := rc.Value(biodv.RecRef); ref != "" {
		fmt.Printf("Reference: %s\n", ref)
	}

	printCollEvent(rc)
	if c := rc.Value(biodv.RecComment); c != "" {
		fmt.Printf("Comments:\n%s\n", c)
	}

	printValues(rc)
	return dataset(rc.Value(biodv.RecDataset))
}

func printTaxon(txm biodv.Taxonomy, rc biodv.Record) error {
	tax, err := txm.TaxID(rc.Taxon())
	if err != nil {
		return err
	}
	fmt.Printf("%s %s [%s:%s]\n", tax.Name(), tax.Value(biodv.TaxAuthor), dbName, tax.ID())
	if !tax.IsCorrect() {
		pID := tax.Parent()
		if pID == "" {
			fmt.Printf("\tUnknown synonym\n")
			return nil
		}
		p, err := txm.TaxID(tax.Parent())
		if err != nil {
			return err
		}
		fmt.Printf("\tSynonym of %s %s [%s:%s]\n", p.Name(), p.Value(biodv.TaxAuthor), dbName, p.ID())
	}
	return nil
}

func printCollEvent(rc biodv.Record) {
	ev := rc.CollEvent()
	fmt.Printf("Collection event:\n")
	if !ev.Date.IsZero() {
		fmt.Printf("\tDate: %v\n", ev.Date)
	}
	if ev.Collector != "" {
		fmt.Printf("\tCollector: %s\n", ev.Collector)
	}
	if c := ev.Country(); c != "" {
		fmt.Printf("\tCountry: %s\n", c)
		if s := ev.State(); s != "" {
			fmt.Printf("\tState: %s\n", s)
		}
		if t := ev.County(); t != "" {
			fmt.Printf("\tCounty: %s\n", t)
		}
	}
	if ev.Locality != "" {
		fmt.Printf("\tLocality: %s\n", ev.Locality)
	}
	if ev.Z != 0 {
		fmt.Printf("\tZ-val: %d\n", ev.Z)
	}

	geo := rc.GeoRef()
	if !geo.IsValid() {
		return
	}
	fmt.Printf("\tLatitude: %.6f\n", geo.Lat)
	fmt.Printf("\tLongitude: %.6f\n", geo.Lon)
	if geo.Source != "" {
		fmt.Printf("\tGeoreference source: %s\n", geo.Source)
	}
	if geo.Validation != "" {
		fmt.Printf("\tGeoreference validation: %s\n", geo.Validation)
	}
}

func ids(rc biodv.Record) {
	fmt.Printf("Record-ID: %s:%s\n", dbName, rc.ID())
	v := rc.Value(biodv.RecExtern)
	for _, e := range strings.Fields(v) {
		fmt.Printf("\t\t%s\n", e)
	}
	if cat := rc.Value(biodv.RecCatalog); cat != "" {
		fmt.Printf("Catalogue-ID: %s\n", cat)
	}
}

func printValues(rc biodv.Record) {
	for _, k := range rc.Keys() {
		if k == biodv.RecOrganism || k == biodv.RecStage || k == biodv.RecSex {
			continue
		}
		if k == biodv.RecDeterm || k == biodv.RecDataset || k == biodv.RecRef {
			continue
		}
		if k == biodv.RecComment {
			continue
		}
		if k == biodv.RecExtern || k == biodv.RecCatalog {
			continue
		}
		if k == biodv.RecDataset {
			continue
		}
		v := rc.Value(k)
		if v == "" {
			continue
		}
		fmt.Printf("%s: %s\n", k, v)
	}
}

func dataset(id string) error {
	if id == "" {
		return nil
	}
	var setDB biodv.SetDB
	ls := biodv.SetDrivers()
	for _, s := range ls {
		if s == dbName {
			var err error
			setDB, err = biodv.OpenSet(dbName, param)
			if err != nil {
				return err
			}
			break
		}
	}
	if setDB == nil {
		return nil
	}

	set, err := setDB.SetID(id)
	if err != nil {
		return err
	}
	fmt.Printf("Source:\t%s\n", set.Title())
	return nil
}
