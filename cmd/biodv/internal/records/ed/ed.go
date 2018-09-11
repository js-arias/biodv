// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package ed implements the rec.ed command,
// i.e. edit records interactively.
package ed

import (
	"fmt"
	"os"
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
	UsageLine: "rec.ed",
	Short:     "edit records interactively",
	Long: `
Command rec.ed implements a simple interactive specimen record editor.

The commands understood by rec.ed are:

    a [<catalog>]
    add [<catalog>]
      Add an specimen record.

    c [<taxon>]
    count [<taxon>]
      Print the number of specimen records of a given taxon.

    del [<record>]
    delete [<record>]
      Delete an specimen record.

    d [<taxon>]
    desc [<taxon>]
      List descendants of a taxon.

    e
    exit
      Shorthand for 'write' and 'quit'.

    h [<command>]
    help [<command>]
      Print command help.

    l [<taxon>]
    list [<taxon>]
      List the IDs of the specimen records of a given taxon.

    m <taxon>
    move <taxon>
      Move a specimen record to a taxon.

    n
    next
      Move to the next specimen record.

    nv
      Shorthand for 'next' and 'view'.

    p
    prev
      Move to the previous specimen record.

    pv
      Shorthand for 'prev' and 'view'.

    q
    quit
      Quit the program, without making any change.

    rk [<taxon>]
    rank [<taxon>]
      Print the rank of a taxon.

    r [<record>]
    record [<record>]
      Move to the indicated specimen record.

    s <key> <value>
    set <key> <value>
      Set a value of an specimen record.

    sv <key> <value>
      Shorthand for 'set' and 'view'.

    t <taxon>
    taxon <taxon>
      Move to the indicated taxon.

    v [<record>]
    view [<record>]
      Show specimen record data.

    w
    write
      Write the database on the hard disk.
	`,
	Run: run,
}

func init() {
	cmdapp.Add(cmd)
}

var txm biodv.Taxonomy
var tax biodv.Taxon
var recs *records.DB
var recLs []*records.Record
var curRec int

func run(c *cmdapp.Command, args []string) error {
	var err error
	txm, err = biodv.OpenTax("biodv", "")
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	recs, err = records.Open("")
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	i := cmdapp.NewInter(os.Stdin, prompt)
	addCommands(i)

	i.Loop()
	return nil
}

func prompt() string {
	var p = "root:"
	if tax != nil {
		p = fmt.Sprintf("%s:", tax.Name())
	}
	if len(recLs) > 0 {
		p += fmt.Sprintf("%s:", recLs[curRec].ID())
	}
	return p
}

func addCommands(i *cmdapp.Inter) {
	i.Add(&cmdapp.Cmd{"a", "add", "add an specimen records", addHelp, addCmd})
	i.Add(&cmdapp.Cmd{"c", "count", "number of specimen records", countHelp, countCmd})
	i.Add(&cmdapp.Cmd{"del", "delete", "delete an specimen record", deleteHelp, deleteCmd})
	i.Add(&cmdapp.Cmd{"d", "desc", "list descendant taxons", descHelp, descCmd})
	i.Add(&cmdapp.Cmd{"e", "exit", "shorthand for 'write' and 'quit'", exitHelp, exitCmd})
	i.Add(&cmdapp.Cmd{"l", "list", "list specimen records", listHelp, listCmd})
	i.Add(&cmdapp.Cmd{"m", "move", "move a specimen record to a taxon", moveHelp, moveCmd})
	i.Add(&cmdapp.Cmd{"n", "next", "move to next specimen record", nextHelp, nextCmd})
	i.Add(&cmdapp.Cmd{"", "nv", "shorthand for 'next' and 'view'", nvHelp, nvCmd})
	i.Add(&cmdapp.Cmd{"p", "prev", "move to previous specimen record", prevHelp, prevCmd})
	i.Add(&cmdapp.Cmd{"", "pv", "shorthand for 'prev' and 'view'", pvHelp, pvCmd})
	i.Add(&cmdapp.Cmd{"q", "quit", "quit the program", quitHelp, func([]string) (bool, error) { return true, nil }})
	i.Add(&cmdapp.Cmd{"rk", "rank", "print taxon rank", rankHelp, rankCmd})
	i.Add(&cmdapp.Cmd{"r", "record", "move to specimen record", recordHelp, recordCmd})
	i.Add(&cmdapp.Cmd{"s", "set", "set a value of an specimen record", setHelp, setCmd})
	i.Add(&cmdapp.Cmd{"", "sv", "shorthand for 'set' and 'view'", svHelp, svCmd})
	i.Add(&cmdapp.Cmd{"t", "taxon", "move to taxon", taxonHelp, taxonCmd})
	i.Add(&cmdapp.Cmd{"v", "view", "print specimen record data", viewHelp, viewCmd})
	i.Add(&cmdapp.Cmd{"w", "write", "write the database on the hard disk", writeHelp, writeCmd})
}

var addHelp = `
Usage:
    a [<catalog>]
    add [<catalog>]
Add a new record for the current taxon, and move the current position
to the new record. If no catalog number is given, an interal ID will
be created for the new record, and the basis of the record will be
set as unknown. If a catalog is given, the record will be set as an
preserved specimen.
`

func addCmd(args []string) (bool, error) {
	if tax == nil {
		return false, errors.New("specimen records can not be added on the root")
	}
	basis := biodv.UnknownBasis
	id := strings.Join(args, " ")
	if id != "" {
		if recs.Record(id) != nil {
			return false, errors.Errorf("catalog number '%s' already in database", id)
		}
		basis = biodv.Preserved
	}
	rec, err := recs.Add(tax.ID(), "", id, basis, 360, 360)
	if err != nil {
		return false, err
	}
	return recordCmd([]string{rec.ID()})
}

var countHelp = `
Usage:
    c [<taxon>]
    count [<taxon>]
Indicates the number of specimen records attached to the indiated
taxon (not including descendants). If no taxon is given, it will use
the current taxon.
`

func countCmd(args []string) (bool, error) {
	nm := strings.Join(args, " ")
	switch nm {
	case "", ".":
		if tax == nil {
			return false, nil
		}
		nm = tax.ID()
	case "/":
		return false, nil
	case "..":
		if tax == nil {
			return false, nil
		}
		nm = tax.Parent()
		if nm == "" {
			return false, nil
		}
	default:
		nt, _ := txm.TaxID(nm)
		if nt == nil {
			return false, nil
		}
		nm = nt.ID()
	}

	ls := recs.RecList(nm)
	fmt.Printf("%d\n", len(ls))
	return false, nil
}

var deleteHelp = `
Usage:
    del [<record>]
    delete [<record>]
Removes the specified record from the database. If no record is given
it will remove the current record.
`

func deleteCmd(args []string) (bool, error) {
	id := strings.Join(args, " ")
	if recLs == nil {
		if id == "" {
			return false, nil
		}
		if recs.Record(id) == nil {
			return false, nil
		}
		recs.Delete(id)
		return false, nil
	}

	if id != "" {
		rec := recs.Record(id)
		if rec.Taxon() != tax.ID() {
			recs.Delete(id)
			return false, nil
		}
		curID := recLs[curRec].ID()
		recs.Delete(id)
		return recordCmd([]string{curID})
	}

	id = recLs[curRec].ID()
	recs.Delete(id)
	ls := recs.RecList(tax.ID())
	if len(ls) == 0 || curRec >= len(ls) {
		recLs = nil
		curRec = 0
		return false, nil
	}
	recLs = ls
	return false, nil
}

var descHelp = `
Usage:
    d [<taxon>]
    desc [<taxon>]
Without parameters shows the list of descendants of the current taxon.
If a taxon is given, it will show the descendants of the indicated
taxon.
`

func descCmd(args []string) (bool, error) {
	nm := strings.Join(args, " ")
	switch nm {
	case "", ".":
		if tax != nil {
			nm = tax.ID()
		}
	case "/":
		nm = ""
	case "..":
		if tax == nil {
			return false, nil
		}
		nm = tax.Parent()
	default:
		nt, _ := txm.TaxID(nm)
		if nt == nil {
			return false, nil
		}
		nm = nt.ID()
	}
	ls, _ := biodv.TaxList(txm.Children(nm))
	if nm != "" {
		syns, _ := biodv.TaxList(txm.Synonyms(nm))
		ls = append(ls, syns...)
	}
	for _, c := range ls {
		fmt.Printf("%s\n", c.Name())
	}
	return false, nil
}

var exitHelp = `
Usage:
    e
    exit
Perform 'write' and then 'quit' commands.
`

func exitCmd(args []string) (bool, error) {
	if _, err := writeCmd(args); err != nil {
		return false, err
	}
	return true, nil
}

var listHelp = `
Usage:
    l [<taxon>]
    list [<taxon>]
List the IDs records of the records of the given taxon. If no taxon
is given, it will list the records of the current taxon.
`

func listCmd(args []string) (bool, error) {
	var ls []*records.Record
	nm := strings.Join(args, " ")
	switch nm {
	case "", ".":
		if tax == nil {
			return false, nil
		}
		ls = recLs
		if len(recLs) == 0 {
			ls = recs.RecList(tax.ID())
		}
		if len(ls) == 0 {
			return false, nil
		}
	case "/":
		return false, nil
	case "..":
		if tax == nil {
			return false, nil
		}
		p := tax.Parent()
		if p == "" {
			return false, nil
		}
		ls = recs.RecList(p)
	default:
		ls = recs.RecList(nm)
	}
	for _, r := range ls {
		fmt.Printf("%s\n", r.ID())
	}
	return false, nil
}

var moveHelp = `
Usage:
    m <taxon>
    move <taxon>
Changes the assignation of the current specimen record to the
indicated taxon. The taxon must be exist in the taxonomy. The
current record will be set to the current specimen record,
in the new taxon assignation.
`

func moveCmd(args []string) (bool, error) {
	if recLs == nil {
		return false, errors.New("a record should be set")
	}
	if len(args) < 1 {
		return false, errors.New("expecing a taxon")
	}

	var nt biodv.Taxon

	nm := strings.Join(args, " ")
	switch nm {
	case "", ".":
		return false, nil
	case "/":
		return false, errors.New("specimen records can not be moved to the root")
	case "..":
		if tax == nil {
			return false, errors.New("specimen records can not be moved to the root")
		}
		nt, _ = txm.TaxID(tax.Parent())
	default:
		nt, _ = txm.TaxID(nm)
	}
	if nt == nil {
		return false, errors.Errorf("taxon '%s' not in database", nm)
	}

	id := recLs[curRec].ID()
	if err := recs.Move(id, nt.ID()); err != nil {
		return false, err
	}
	return recordCmd([]string{id})
}

var nextHelp = `
Usage:
    n
    next
Move the record to the next record of the list.
`

func nextCmd(args []string) (bool, error) {
	if tax == nil {
		return false, nil
	}
	if len(recLs) == 0 {
		ls := recs.RecList(tax.ID())
		if len(ls) == 0 {
			return false, nil
		}
		recLs = ls
		curRec = 0
	} else {
		curRec++
		if curRec >= len(recLs) {
			recLs = nil
			curRec = 0
		}
	}
	return false, nil
}

var nvHelp = `
Usage:
    nv
Perform 'next' and then 'view' commands.
`

func nvCmd(args []string) (bool, error) {
	nextCmd(nil)
	return viewCmd(nil)
}

var prevHelp = `
Usage:
    p
    prev
Move the record to the previous record of the list.
`

func prevCmd(args []string) (bool, error) {
	if tax == nil {
		return false, nil
	}
	if len(recLs) == 0 {
		ls := recs.RecList(tax.ID())
		if len(ls) == 0 {
			return false, nil
		}
		recLs = ls
		curRec = len(ls) - 1
	} else {
		curRec--
		if curRec < 0 {
			recLs = nil
			curRec = 0
		}
	}
	return false, nil
}

var pvHelp = `
Usage:
    pv
Perform 'prev' and the 'view' command.
`

func pvCmd(args []string) (bool, error) {
	prevCmd(nil)
	return viewCmd(nil)
}

var quitHelp = `
Usage:
    q
    quit
Ends the program without saving any change.
`

var rankHelp = `
Usage:
    rk [<taxon>]
    rank [<taxon>]
Print the rank of a taxon. If no taxon is given it will print the
rank of the current taxon. If the taxon is unranked, the rank of
the most inmediate ranked parent will be printed in parenthesis.
`

func rankCmd(args []string) (bool, error) {
	tx := tax
	nm := strings.Join(args, " ")
	switch nm {
	case "", ".":
		if tax == nil {
			return false, nil
		}
	case "/":
		return false, nil
	case "..":
		if tax == nil {
			return false, nil
		}
		tx, _ = txm.TaxID(tax.Parent())
	default:
		tx, _ = txm.TaxID(nm)
	}
	if tx == nil {
		return false, nil
	}
	r := tx.Rank()
	if r == biodv.Unranked {
		r = getRank(tx)
		if r == biodv.Unranked {
			fmt.Printf("%s\n", r)
			return false, nil
		}
		fmt.Printf("%s (%s)\n", biodv.Unranked, r)
		return false, nil
	}
	fmt.Printf("%s\n", r)
	return false, nil
}

func getRank(tx biodv.Taxon) biodv.Rank {
	for tx != nil {
		if tx.Rank() != biodv.Unranked {
			return tx.Rank()
		}
		tx, _ = txm.TaxID(tx.Parent())
	}
	return biodv.Unranked
}

var recordHelp = `
Usage:
    r [<record>]
    record [<record>]
Change the current record to the indicated record. If no record ID
is given, it will use the first record of the current
taxon.
`

func recordCmd(args []string) (bool, error) {
	id := strings.Join(args, " ")
	if id == "" {
		recLs = recs.RecList(tax.ID())
		curRec = 0
		if len(recLs) == 0 {
			return false, nil
		}
	} else {
		rec := recs.Record(id)
		if rec == nil {
			return false, nil
		}
		if rec.Taxon() != tax.ID() {
			nt, _ := txm.TaxID(rec.Taxon())
			if nt == nil {
				return false, nil
			}
			tax = nt
		}
		recLs = recs.RecList(tax.ID())
		for i, r := range recLs {
			if r.ID() == rec.ID() {
				curRec = i
				break
			}
		}
	}
	return false, nil
}

var setHelp = `
Usage:
    s <key> <value>
    set <key> <value>
Set a value of a field of the current specimen record. In general
the format value will depend on the field.

Any key can be stored, but the recognized keys (and their expected
values are):
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
        z            in flying or oceanic specimens, the distance to
                     groud (depth as negative) when the sampling was
                     made.
        latlon       the decimal latitude and longitude of the record,
                     separated by one or more spaces.
        elevation    elevation over sea level, in meters.
        uncertainty  georeference uncertainty in meters.
        geosource    source of the georeference.
        validation   validation of the georeference.
        reference    a bibliographic reference
        dataset      source of the specimen record information
        determiner   the person who identified the specimen
        organism     the organism ID
        stage        the growth stage of the organism
        sex          sex of the organism

If value is set to ‘-’ then it will remove any value from the given
key. If the value starts with a ‘+’ it will be append the value (in
the case that append is valid).
`

func setCmd(args []string) (bool, error) {
	if recLs == nil {
		return false, errors.New("a record should be set")
	}
	if len(args) < 2 {
		return false, errors.New("expecing <key> <value> parameters")
	}
	rec := recLs[curRec]
	key := strings.ToLower(args[0])
	value := strings.Join(args[1:], " ")
	if strings.HasPrefix(value, "+") {
		value = rec.Value(key) + "\n" + value[1:]
	}
	if value == "-" {
		value = ""
	}

	switch key {
	case "date":
		ev := rec.CollEvent()
		ev.Date = time.Time{}
		if value != "" {
			t, err := time.Parse(time.RFC3339, value)
			if err != nil {
				return false, errors.Wrap(err, "invalid time value")
			}
			ev.Date = t
		}
		rec.SetCollEvent(ev)
	case "country":
		ev := rec.CollEvent()
		ev.Admin.Country = ""
		if value != "" {
			value = strings.ToUpper(value)
			if !geography.IsValidCode(value) {
				return false, errors.New("invalid country code")
			}
			ev.Admin.Country = value
		}
		rec.SetCollEvent(ev)
	case "state":
		ev := rec.CollEvent()
		ev.Admin.State = value
		rec.SetCollEvent(ev)
	case "county":
		ev := rec.CollEvent()
		ev.Admin.County = value
		rec.SetCollEvent(ev)
	case "locality":
		ev := rec.CollEvent()
		ev.Locality = value
		rec.SetCollEvent(ev)
	case "collector":
		ev := rec.CollEvent()
		ev.Collector = value
		rec.SetCollEvent(ev)
	case "z":
		ev := rec.CollEvent()
		ev.Z = 0
		if value != "" {
			z, err := strconv.Atoi(value)
			if err != nil {
				return false, errors.Wrap(err, "invalid z value")
			}
			ev.Z = z
		}
		rec.SetCollEvent(ev)
	case "latlon":
		geo := rec.GeoRef()
		if value == "" {
			geo = geography.NewPosition()
			rec.SetGeoRef(geo)
			break
		}
		v := strings.Fields(value)
		if len(v) != 2 {
			return false, errors.New("latlon value should be a pair o values")
		}
		lat, err := strconv.ParseFloat(v[0], 64)
		if err != nil {
			return false, errors.Wrap(err, "invalid latitude value")
		}
		lon, err := strconv.ParseFloat(v[1], 64)
		if err != nil {
			return false, errors.Wrap(err, "invalid longitude value")
		}
		if !geography.IsValidCoord(lat, lon) {
			return false, errors.New("invalid latitude, longitude values")
		}
		geo.Lat = lat
		geo.Lon = lon
		rec.SetGeoRef(geo)
	case "elevation":
		geo := rec.GeoRef()
		geo.Elevation = 0
		if value != "" {
			e, err := strconv.Atoi(value)
			if err != nil {
				return false, errors.Wrap(err, "invalid elevation")
			}
			if e < 0 {
				return false, errors.New("invalid elevation: negative value")
			}
			geo.Elevation = uint(e)
		}
		rec.SetGeoRef(geo)
	case "uncertainty":
		geo := rec.GeoRef()
		geo.Uncertainty = 0
		if value != "" {
			u, err := strconv.Atoi(value)
			if err != nil {
				return false, errors.Wrap(err, "invalid uncertainty")
			}
			if u < 0 {
				return false, errors.New("invalid uncertainty: negative value")
			}
			geo.Uncertainty = uint(u)
		}
		rec.SetGeoRef(geo)
	case "geosource":
		geo := rec.GeoRef()
		geo.Source = value
		rec.SetGeoRef(geo)
	case "validation":
		geo := rec.GeoRef()
		geo.Validation = value
		rec.SetGeoRef(geo)
	default:
		if err := rec.Set(key, value); err != nil {
			return false, err
		}
	}
	return false, nil
}

var svHelp = `
Usage:
    sv <key> <value>
Perform 'set' and the 'view' command.
`

func svCmd(args []string) (bool, error) {
	if _, err := setCmd(args); err != nil {
		return false, err
	}
	return viewCmd(nil)
}

var taxonHelp = `
Usage:
    t <taxon>
    taxon <taxon>
Changes the current taxon to the indicated taxon. To move to a
parent use '..' to move to a parent, or use '/' to move to the
root of the taxonomy.
`

func taxonCmd(args []string) (bool, error) {
	nm := strings.Join(args, " ")
	switch nm {
	case "", ".":
		return false, nil
	case "/":
		tax = nil
	case "..":
		if tax == nil {
			return false, nil
		}
		tax, _ = txm.TaxID(tax.Parent())
	default:
		nt, _ := txm.TaxID(nm)
		if nt == nil {
			return false, nil
		}
		tax = nt
	}
	recLs = nil
	curRec = 0
	return false, nil
}

var viewHelp = `
Usage:
    v <record>
    view <record>
Shows the information stored on the indicated record. If no record
is given, it will show the current record.
`

func viewCmd(args []string) (bool, error) {
	var rec *records.Record
	id := strings.Join(args, " ")
	if id == "" {
		if tax == nil {
			return false, nil
		}
		if len(recLs) == 0 {
			return false, nil
		}
		rec = recLs[curRec]
	} else {
		rec = recs.Record(id)
		if rec == nil {
			return false, nil
		}
	}

	fmt.Printf("record:\t%s\n", rec.ID())
	fmt.Printf("taxon:\t%s\n", rec.Taxon())
	fmt.Printf("basis:\t%s\n", rec.Basis())
	ev := rec.CollEvent()
	fmt.Printf("date:\t%s\n", ev.Date.Format(time.RFC3339))
	fmt.Printf("country:\t%s\n", ev.Admin.Country)
	fmt.Printf("state:\t%s\n", ev.Admin.State)
	fmt.Printf("county:\t%s\n", ev.Admin.County)
	fmt.Printf("locality:\t%s\n", ev.Locality)
	fmt.Printf("collector:\t%s\n", ev.Collector)
	fmt.Printf("z\t%d\n", ev.Z)
	geo := rec.GeoRef()
	if !geo.IsValid() {
		fmt.Printf("latlon:\tNA\n")
		fmt.Printf("elevation:\t0\n")
		fmt.Printf("uncertainty:\t0\n")
		fmt.Printf("geosource:\n")
		fmt.Printf("validation:\n")
	} else {
		fmt.Printf("latlon:\t%f %f\n", geo.Lat, geo.Lon)
		fmt.Printf("elevation:\t%d\n", geo.Elevation)
		fmt.Printf("uncertainty:\t%d\n", geo.Uncertainty)
		fmt.Printf("geosource:\t%s\n", geo.Source)
		fmt.Printf("validation:\t%s\n", geo.Validation)
	}
	for _, k := range rec.Keys() {
		v := rec.Value(k)
		fmt.Printf("%s:\t%s\n", k, v)
	}
	return false, nil
}

var writeHelp = `
Usage:
    w
    write
Write all changes made to the database since the start of the
edition season, or the last writing.
`

func writeCmd(args []string) (bool, error) {
	if err := recs.Commit(); err != nil {
		return false, err
	}
	return false, nil
}
