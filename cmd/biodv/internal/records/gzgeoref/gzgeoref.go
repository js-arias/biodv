// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package gzgeoref implements the rec.gz.georef command,
// i.e. georeference specimen records.
package gzgeoref

import (
	"fmt"
	"os"
	"strings"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/cmdapp"
	"github.com/js-arias/biodv/geography"
	"github.com/js-arias/biodv/records"

	"github.com/pkg/errors"
)

var cmd = &cmdapp.Command{
	UsageLine: `rec.gz.georef -s|--service <service>
		[-u|--uncertainty <number>] [<name>]`,
	Short: "georeference specimen records",
	Long: `
Command rec.gz.georef sets georeference values from a given gazetteer
service. The -u or --uncertainty option sets the maximum uncertainty in
meters, the default value is 50000 (50 km). If set to 0, any uncertainty
is accepted.

When records are ambiguous, and uncertainty is set to 0, the record will
be left ungeoreferenced. In other cases, the first returned record will
be set as the point, and if distance to all other records in the set fall
inside the uncertainty level, then the point will be assigned.

Options are:

    -s <service>
    --service <service>
      A required parameter. The gazetteer service to be used.

    -u <number>
    --uncertainty <number>
      Sets the maximum uncertainty in meters. Default value: 50000
      (i.e. 50 km).

    <name>
      If set, only the records for the indicated taxon (and its
      descendants) will be georeferenced.
	`,
	Run:           run,
	RegisterFlags: register,
}

func init() {
	cmdapp.Add(cmd)
}

var service string
var uncertainty uint

func register(c *cmdapp.Command) {
	c.Flag.StringVar(&service, "service", "", "")
	c.Flag.StringVar(&service, "s", "", "")
	c.Flag.UintVar(&uncertainty, "uncertainty", 50000, "")
	c.Flag.UintVar(&uncertainty, "u", 50000, "")
}

func run(c *cmdapp.Command, args []string) error {
	if service == "" {
		return errors.Errorf("%s: a gazetteer serive must be defined", c.Name())
	}
	var param string
	service, param = biodv.ParseDriverString(service)
	gz, err := biodv.OpenGz(service, param)
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	txm, err := biodv.OpenTax("biodv", "")
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	recs, err := records.Open("")
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	if len(args) > 0 {
		nm := strings.Join(args, " ")
		tax, _ := txm.TaxID(nm)
		if tax == nil {
			return nil
		}
		procTaxon(txm, gz, recs, tax)
		if err := recs.Commit(); err != nil {
			return errors.Wrap(err, c.Name())
		}
		return nil
	}

	ls, err := biodv.TaxList(txm.Children(""))
	if err != nil {
		return errors.Wrap(err, c.Name())
	}
	for _, tax := range ls {
		procTaxon(txm, gz, recs, tax)
	}
	if err := recs.Commit(); err != nil {
		return errors.Wrap(err, c.Name())
	}
	return nil
}

// ProcTaxon georeference the records of a given taxon.
func procTaxon(txm biodv.Taxonomy, gz biodv.Gazetteer, recs *records.DB, tax biodv.Taxon) {
	if getRank(txm, tax) < biodv.Species {
		procChildren(txm, gz, recs, tax)
		return
	}

	ls := recs.RecList(tax.ID())
	if len(ls) == 0 {
		procChildren(txm, gz, recs, tax)
		return
	}

	for _, r := range ls {
		geo := r.GeoRef()
		if geo.IsValid() {
			continue
		}

		ev := r.CollEvent()
		if !geography.IsValidCode(ev.CountryCode()) {
			continue
		}
		if ev.Locality == "" {
			ev.Locality = ev.County()
		}
		if ev.Locality == "" {
			ev.Locality = ev.State()
		}
		if ev.Locality == "" {
			continue
		}
		sg := gz.Locate(ev.Admin, ev.Locality)
		i := 0
		var max uint
		for sg.Scan() {
			p := sg.Position()
			i++
			if i == 1 {
				geo = p
				continue
			}
			if uncertainty == 0 {
				fmt.Fprintf(os.Stderr, "warning: %s [tax: %s] ambiguous\n", r.ID(), tax.Name())
				geo = geography.NewPosition()
				sg.Close()
				break
			}
			d := geo.MaxDist(p)
			if d > uncertainty {
				fmt.Fprintf(os.Stderr, "warning: %s [tax: %s] ambiguous\n", r.ID(), tax.Name())
				geo = geography.NewPosition()
				sg.Close()
				break
			}
			if d > max {
				max = d
			}
		}
		if err := sg.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: %s [tax: %s]: %v\n", r.ID(), tax.Name(), err)
			continue
		}
		if !geo.IsValid() {
			if i == 0 {
				fmt.Fprintf(os.Stderr, "warning: %s [tax: %s] unable to locate %s: %s\n", r.ID(), tax.Name(), ev.Country(), ev.Locality)
			}
			continue
		}

		if max > geo.Uncertainty {
			geo.Source += fmt.Sprintf(" (average of %d locations)", i)
			geo.Uncertainty = max
		}
		r.SetGeoRef(geo)
	}
	procChildren(txm, gz, recs, tax)
}

func procChildren(txm biodv.Taxonomy, gz biodv.Gazetteer, recs *records.DB, tax biodv.Taxon) {
	children, _ := biodv.TaxList(txm.Children(tax.ID()))
	syns, _ := biodv.TaxList(txm.Synonyms(tax.ID()))
	children = append(children, syns...)

	for _, c := range children {
		procTaxon(txm, gz, recs, c)
	}
}

// GetRank returns the rank of a taxon,
// or the rank of a ranked parent
// if the taxon is unranked.
func getRank(txm biodv.Taxonomy, tax biodv.Taxon) biodv.Rank {
	for p := tax; p != nil; p, _ = txm.TaxID(p.Parent()) {
		if p.Rank() != biodv.Unranked {
			return p.Rank()
		}
	}
	return biodv.Unranked
}
