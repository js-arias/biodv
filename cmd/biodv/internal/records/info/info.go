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

func register(c *cmdapp.Command) {
	c.Flag.StringVar(&dbName, "db", "", "")
}

func run(c *cmdapp.Command, args []string) error {
	if dbName == "" {
		return errors.Errorf("%s: no database defined", c.Name())
	}
	var param string
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
	fmt.Printf("Record: %s:%s\n", dbName, rc.ID())
	if cat := rc.Value(biodv.RecCatalog); cat != "" {
		fmt.Printf("Catalogue ID: %s\n", cat)
	}
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
	return nil
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
	if ev.Country != "" {
		fmt.Printf("\tCountry: %s\n", ev.Country)
		if ev.State != "" {
			fmt.Printf("\tState: %s\n", ev.State)
		}
		if ev.County != "" {
			fmt.Printf("\tCounty: %s\n", ev.County)
		}
	}
	if ev.Locality != "" {
		fmt.Printf("\tLocality: %s\n", ev.Locality)
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
