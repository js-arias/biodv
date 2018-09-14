// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package catalog implements the tax.catalog command,
// i.e. print a taxonomic catalog.
package catalog

import (
	"fmt"
	"html"
	"os"
	"strings"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/cmdapp"

	"github.com/pkg/errors"
)

var cmd = &cmdapp.Command{
	UsageLine: `tax.catalog [--db <database>] [--id]
	[-f|--format <value>] <taxon>`,
	Short: "print a taxonomic catalog",
	Long: `
Command tax.catalog prints the taxonomy of the indicated taxon in the
format of a simple taxonomic catalog.

Options are:
    -db <database>
    --db <database>
      If set, the indicated database will be used to extract the
      taxonomic information.
      To see the available databases use the command ‘db.drivers’.
      The default biodv database on the current directory.

    -id
    --id
      If set, the search of the taxon will be based on the taxon ID,
      instead of the taxon name.

    -f <value>
    --format <value>
      Sets the output format, by default it will use txt format.
      Valid format are:
          txt	text format
          html	html format

    <taxon>
      A required parameter. Indicates the taxon for which the taxonomy
      catalog of the taxon will be printed, if the name is ambiguous,
      the ID of the ambiguous taxa will be printed. If the option --id
      is set, it must be a taxon ID instead of a taxon name.
	`,
	Run:           run,
	RegisterFlags: register,
}

func init() {
	cmdapp.Add(cmd)
}

var dbName string
var id bool
var format string

func register(c *cmdapp.Command) {
	c.Flag.StringVar(&dbName, "db", "biodv", "")
	c.Flag.BoolVar(&id, "id", false, "")
	c.Flag.StringVar(&format, "format", "txt", "")
	c.Flag.StringVar(&format, "f", "txt", "")
}

func run(c *cmdapp.Command, args []string) error {
	if dbName == "" {
		dbName = "biodv"
	}
	var param string
	dbName, param = biodv.ParseDriverString(dbName)

	nm := strings.Join(args, " ")
	if nm == "" {
		return errors.Errorf("%s: a taxon name or ID, should be given", c.Name())
	}

	format = strings.ToLower(format)
	switch format {
	case "txt":
	case "html":
		fmt.Printf("<html>\n")
		fmt.Printf("<head><meta http-equiv=\"Content-Type\" content=\"text/html\" charset=utf-8\" /></head>\n")
		fmt.Printf("<body bgcolor=\"white\">\n<font face=\"sans-serif\"><pre>\n")
	default:
		return errors.Errorf("%s: unknown format %s", c.Name(), format)
	}

	db, err := biodv.OpenTax(dbName, param)
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	tax, err := getTaxon(db, nm)
	if err != nil {
		return errors.Wrap(err, c.Name())
	}
	if tax == nil {
		return nil
	}

	if err := navigate(db, tax, biodv.Unranked); err != nil {
		return errors.Wrap(err, c.Name())
	}
	return nil
}

// GetTaxon returns a taxon from the options.
func getTaxon(db biodv.Taxonomy, nm string) (biodv.Taxon, error) {
	if id {
		return db.TaxID(nm)
	}
	ls, err := biodv.TaxList(db.Taxon(nm))
	if err != nil {
		return nil, err
	}
	if len(ls) == 0 {
		return nil, nil
	}
	if len(ls) > 1 {
		fmt.Fprintf(os.Stderr, "ambiguous name:\n")
		for _, tx := range ls {
			fmt.Fprintf(os.Stderr, "id:%s\t%s %s\t", tx.ID(), tx.Name(), tx.Value(biodv.TaxAuthor))
			if tx.IsCorrect() {
				fmt.Fprintf(os.Stderr, "correct name\n")
			} else {
				fmt.Fprintf(os.Stderr, "synonym\n")
			}
		}
		return nil, nil
	}
	return ls[0], nil
}

// Navigate follows the taxonomy.
func navigate(db biodv.Taxonomy, tax biodv.Taxon, prev biodv.Rank) error {
	if err := printTaxon(db, tax, prev); err != nil {
		return err
	}
	r := tax.Rank()
	if r == biodv.Unranked {
		r = prev
	}

	ls, err := biodv.TaxList(db.Children(tax.ID()))
	if err != nil {
		return err
	}
	for _, c := range ls {
		if err := navigate(db, c, r); err != nil {
			return err
		}
	}
	return nil
}

// PrintTaxon prints taxon catalog information.
func printTaxon(db biodv.Taxonomy, tax biodv.Taxon, prev biodv.Rank) error {
	r := tax.Rank()
	if r == biodv.Unranked {
		r = prev
	}

	ls, err := biodv.TaxList(db.Synonyms(tax.ID()))
	if err != nil {
		return err
	}

	if r < biodv.Species {
		printSupra(tax, ls)
		return nil
	}
	printSpecies(tax, ls)
	return nil
}

func printSupra(tax biodv.Taxon, syns []biodv.Taxon) {
	nm := strings.ToTitle(tax.Name())
	switch format {
	case "html":
		nm = html.EscapeString(nm)
		ids := getIDsString(tax)
		rk := tax.Rank()

		if rk != biodv.Unranked {
			if rk == biodv.Genus {
				fmt.Printf("\n%s <strong><i>%s</i></strong> %s", strings.Title(rk.String()), nm, html.EscapeString(tax.Value(biodv.TaxAuthor)))
			} else {
				fmt.Printf("\n%s <strong>%s</strong> %s", strings.Title(rk.String()), nm, html.EscapeString(tax.Value(biodv.TaxAuthor)))
			}
		} else {
			fmt.Printf("\n<strong>%s</strong> %s", nm, html.EscapeString(tax.Value(biodv.TaxAuthor)))
		}
		fmt.Printf(" <font size=-1>[%s]</font>\n", ids)
		for _, s := range syns {
			sid := getIDsString(s)
			if s.Rank() == biodv.Genus {
				fmt.Printf("\t<font color=\"gray\"><i>%s</i> %s <font size=-1>[%s]</font></font>\n", html.EscapeString(s.Name()), html.EscapeString(s.Value(biodv.TaxAuthor)), sid)
				continue
			}
			fmt.Printf("\t<font color=\"gray\">%s %s <font size=-1>[%s]</font></font>\n", html.EscapeString(s.Name()), html.EscapeString(s.Value(biodv.TaxAuthor)), sid)
		}
	case "txt":
		ids := getIDsString(tax)
		if tax.Rank() != biodv.Unranked {
			fmt.Printf("\n%s %s %s\n", strings.ToTitle(tax.Rank().String()), nm, tax.Value(biodv.TaxAuthor))
		} else {
			fmt.Printf("\n%s %s\n", nm, tax.Value(biodv.TaxAuthor))
		}
		fmt.Printf("\t\t[%s]\n", ids)
		for _, s := range syns {
			sid := html.EscapeString(getIDsString(s))
			fmt.Printf("\t%s %s [%s]\n", s.Name(), s.Value(biodv.TaxAuthor), sid)
		}
		fmt.Printf("\n")
	}
}

func printSpecies(tax biodv.Taxon, syns []biodv.Taxon) {
	switch format {
	case "html":
		nm := html.EscapeString(tax.Name())
		ids := getIDsString(tax)
		if tax.Rank() != biodv.Species {
			fmt.Printf("\t\t<i>%s</i> %s <font size=-1>[%s]</font>\n", nm, html.EscapeString(tax.Value(biodv.TaxAuthor)), ids)
		} else {
			fmt.Printf("\t<i>%s</i> %s", nm, html.EscapeString(tax.Value(biodv.TaxAuthor)))
			fmt.Printf(" <font size=-1>[%s]</font>\n", ids)
		}
		for _, s := range syns {
			sid := getIDsString(s)
			if tax.Rank() != biodv.Species {
				fmt.Printf("\t\t\t<font color=\"gray\"><i>%s</i> %s <font size=-1>[%s]</font></font>\n", html.EscapeString(s.Name()), html.EscapeString(s.Value(biodv.TaxAuthor)), sid)
				continue
			}
			fmt.Printf("\t\t<font color=\"gray\"><i>%s</i> %s <font size=-1>[%s]</font></font>\n", html.EscapeString(s.Name()), html.EscapeString(s.Value(biodv.TaxAuthor)), sid)
		}
	case "txt":
		ids := getIDsString(tax)
		if tax.Rank() != biodv.Species {
			fmt.Printf("\t\t%s %s [%s]\n", tax.Name(), tax.Value(biodv.TaxAuthor), ids)
		} else {
			fmt.Printf("\t%s %s\n", tax.Name(), tax.Value(biodv.TaxAuthor))
			fmt.Printf("\t\t\t[%s]\n", ids)
		}
		for _, s := range syns {
			sid := html.EscapeString(getIDsString(s))
			if tax.Rank() != biodv.Species {
				fmt.Printf("\t\t\t%s %s [%s]\n", s.Name(), s.Value(biodv.TaxAuthor), sid)
				continue
			}
			fmt.Printf("\t\t%s %s [%s]\n", s.Name(), s.Value(biodv.TaxAuthor), sid)
		}
	}
}

func getIDsString(tax biodv.Taxon) string {
	ext := tax.Value(biodv.TaxExtern)
	ls := append([]string{dbName + ":" + tax.ID()}, strings.Fields(ext)...)

	switch format {
	case "html":
		var b strings.Builder
		for i, id := range ls {
			u := getGetTaxonURL(id)
			if i == 0 {
				b.WriteString(u)
				continue
			}
			b.WriteString(" " + u)
		}
		return b.String()
	case "txt":
		return strings.Join(ls, " ")
	}
	return ""
}

func getGetTaxonURL(id string) string {
	i := strings.Index(id, ":")
	if i <= 0 {
		return html.EscapeString(id)
	}
	srv, eid := id[:i], id[i+1:]
	u := biodv.TaxURL(srv, eid)
	if u == "" {
		return html.EscapeString(id)
	}
	return fmt.Sprintf("<a href=%s>%s</a>", u, html.EscapeString(id))
}
