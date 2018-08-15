// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package main

import (
	"github.com/js-arias/biodv/cmdapp"

	// initialize taxonomy sub-commands
	_ "github.com/js-arias/biodv/cmd/biodv/internal/taxonomy/add"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/taxonomy/catalog"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/taxonomy/dbadd"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/taxonomy/dbfill"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/taxonomy/dbsync"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/taxonomy/dbupdate"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/taxonomy/del"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/taxonomy/format"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/taxonomy/info"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/taxonomy/list"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/taxonomy/move"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/taxonomy/rank"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/taxonomy/set"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/taxonomy/validate"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/taxonomy/value"
)

var taxHelp = &cmdapp.Command{
	UsageLine: "taxonomy",
	Short:     "taxonomy database",
	Long: `
In biodv the taxonomy database is stored in taxonomy sub-directory in
the file taxonomy.stz. The file is an stanza-encoded file.

The following fields are recognized by biodv taxonomy commands:

	name       cannonical name of the taxon.
	author     author of the taxon's name.
	rank       rank of the taxon.
	correct    if the taxon is not correct (a synonym) is set to
	           "false", any other value is interpreted as "true"
	           (i.e. a correct name).
	parent     name of the taxon's parent.
	extern     a list (separated by spaces) of external IDs of the
	           taxon, it is expected to be in the form
	           <service>:<key>.
	reference  a bibliographic reference (or an ID to that
	           reference).
	source	   the database (or its ID) of the source of the
	           taxonomic information.

There are some constrains in the file:
	
	(i)     if a taxon record has a parent, this parent should be
	        defined previously, and be a correct name.
	(ii)    if a taxon is not a correct name (i.e. a synonym), it should
		have a parent.
	(iii)	ranks should be consistent within the taxonomy.

Usually this constrains are assumed by most biodv commands, as it is expected
that the file conforms to this constrains. In case of a untrusted database,
it can be validated with the command tax.validate.
	`,
}

func init() {
	cmdapp.Add(taxHelp)
}
