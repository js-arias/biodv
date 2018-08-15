// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Biodv is a tool for management and analysis of biodiveristy data.
package main

import (
	"github.com/js-arias/biodv/cmdapp"

	// load drivers
	_ "github.com/js-arias/biodv/driver/gbif"
)

func main() {
	cmdapp.Short = "Biodv is a tool for management and analysis of biodiveristy data."
	cmdapp.Main()
}

var dbHelp = &cmdapp.Command{
	UsageLine: "database",
	Short:     "biodv database organization",
	Long: `
In biodv, the database is organized in sub-directories within the
current working path (which is defined here as a “project”). Each
sub-directory stores the data in one or more files using the stanza
format.

Here is an example of a project sub-directories:

	my-project/
		taxonomy/
			taxonomy.stz
		specimens/
			taxon-a.stz
			taxon-b.stz
			synonym-a.stz
			...
		sources/
			biblio.stz
			datasets.stz
			sources.stz
		geography/
			geography.stz

Not all sub-directories are required, as they will be created given
the needs of each project.

Commands in biodv look out for the database files automatically, so
most of the time, the user does not need to know how each file is
called, or where it can be found.

Each data guide include more details on how each subdirectory is
organized, and the particular constrains of the stored data in that
subdirectories.
	`,
}

func init() {
	cmdapp.Add(dbHelp)
}

var stanzaHelp = &cmdapp.Command{
	UsageLine: "stanza",
	Short:     "stanza file format",
	Long: `
In biodv data is stored using the stanza format. The stanza format is an
structured UTF-8 text file with the following rules:

	1. Each line containing a field must starts with the field name
	   and separated from its content by a colon ( : ) character. If
	   the field name ends with a new line rather than a colon, the
	   field is considered as empty.
	2. Field names are case insensitive (always read as lower caps),
	   without spaces, and should be unique.
	3. A field ends with a new line. If the content of a field
	   extends more than one line, the next line should start with at
	   least one space or tab character.
	4. A record ends with a line that starts with percent sign ( % )
	   character. Any character after the percent sign will be
	   ignored (usually %% is used to increase visibility of the
	   end-of-record).
	5. Lines starting with the sharp symbol ( # ) character are
	   taken as comments.
	6. Empty lines are ignored.

Here is a simple example of an stanza file containing taxonomic data:

	# example dataset
	name:	Homo
	rank:	genus
	correct: true
	%%
	name:	Pan
	rank:	genus
	correct: true
	%%
	name:	Homo sapiens
	parent:	Homo
	rank:	species
	correct: true
	%%
	name:	Homo erectus
	parent:	Homo
	rank:	species
	correct: true
	%%
	name:	Pithecanthropus
	parent:	Homo
	rank:	genus
	correct: false
	%%
	name:	Sinanthropus pekinensis
	parent:	Homo erectus
	rank:	species
	correct: false
	%%

Stanza file format was inspired by the record-jar/stanza format
described by E. Raymond "The Art of UNIX programming"
<http://www.catb.org/esr/writings/taoup/html/ch05s02.html#id2906931>
(2003) Addison-Wesley, and C. Strozzi "NoSQL list format"
<http://www.strozzi.it/cgi-bin/CSA/tw7/I/en_US/NoSQL/Table%20structure>
(2007).
	`,
}

func init() {
	cmdapp.Add(stanzaHelp)
}
