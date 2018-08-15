// Authomatically generated doc.go file for use with godoc.

/*
Biodv is a tool for management and analysis of biodiveristy data.

Usage:

    biodv [help] <command> [<args>...]

The commands are:
    help             display help information about biodv
    tax.add          add taxon names
    tax.catalog      print a taxonomic catalog
    tax.db.add       add taxons validated on an external DB
    tax.db.fill      add taxons from an external DB
    tax.db.sync      synchronize the local DB to an external taxonomy
    tax.db.update    update taxon information from an external DB
    tax.del          eliminate a taxon from the database
    tax.format       synonymize rankless taxa
    tax.info         print taxon information
    tax.list         print a list of taxons
    tax.move         change a taxon parent
    tax.rank         change a taxon rank
    tax.set          set a taxon data value
    tax.validate     validate a taxonomy database
    tax.value        get a taxon data value

Use 'biodv help <command>' for more information about a command.

Additional help topics:

    database         biodv database organization
    stanza           stanza file format
    taxonomy         taxonomy database

Use 'biodv help <topic>' for more information about that topic.

Biodv database organization

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

Display help information about biodv

Usage:

	biodv help [<command>]

Command help displays help information for a command or a help topic.

With no arguments it prints the list of available commands and help topics to
the standard output.

Stanza file format

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

Add taxon names

Usage:

	biodv tax.add [-p|--parent <name>] [-r|--rank <rank>]
		[-s|--synonym] [<file>...]

Command tax.add adds one or more taxons from the indicated files, or the
standard input (if no file is defined) to the taxonomy database. It
assumes that each line contains a taxon name (empty lines, or lines
starting with the sharp symbol ( # ) or semicolon character ( ; ) will
be ignored).

By default, taxons will be added to the root of the taxonomy (without
parent), as correct names, and rankless.

Options are:

    -p <name>
    --parent <name>
      Sets the parents of the added taxons. This must be a correct name
      (i.e. not a synonym), and should be present in the database.

    -r <rank>
    --rank <rank>
      Sets the rank of the added taxons. If a parent is defined, the rank
      should be concordant with the parent's rank.
      Valid ranks are:
        unranked
        kingdom
        class
        order
        family
        genus
        species
      If the rank is set to species, it will automatically interpret the
      first part of the name as a genus, and adds the species as a child
      of that genus.

    -s
    --synonym
      If set, the added taxons will be set as synonyms (non correct names)
      of its parent. It requires that a parent will be defined.

    <file>
      One or more files to be processed by tax.add. If no file is given,
      the taxon names will be read from the standard input.

Print a taxonomic catalog

Usage:

	biodv tax.catalog [--db <database>] [-f|--format <value>]
		[--id <value>] [<name>]

Command tax.catalog prints the taxonomy of the indicated taxon in the
format of a simple taxonomic catalog.

Options are:
    -db <database>
    --db <database>
      If set, the indicated database will be used to extract the
      taxonomic information.
      Available databases are:
        biodv	default database (on current directory)
        gbif	GBIF webservice (requires internet connection)

    -id <value>
    --id <value>
      If set, the taxonomy catalog of the indicated taxon will be
      printed.


    -f <value>
    --format <value>
      Sets the output format, by default it will use txt format.
      Valid format are:
          txt	text format
          html	html format

    <name>
      If set, the taxonomy catalog of the taxon will be printed, if the
      name is ambiguous, the ID of the ambigous taxa will be printed.

Add taxons validated on an external DB

Usage:

	biodv tax.db.add -e|--extern <database> [-u|--uprank <rank>]
		[<file>...]

Command tax.db.add adds one or more taxons from the indicated files,
or the standard input (if no file is defined) to the taxonomy database.
Each name is validated against an external database, and only taxons
found in that DB will be added (with the additional information given
by the database.

If a taxon name can not be added, it will be printed in the standard
output.

In the input file, it is assumed that each line contains a taxon name
(empty lines, or lines starting with the sharp symbol ( # ) or
semicolon character ( ; ) will be ignored).

If the option -u or --uprank is given, it will add additional parents up
to the given rank.

Options are:

    -e <database>
    --extern <database>
      A required parameter. It will set the external database.
      Available databases are:
        gbif	GBIF webservice (requires internet connection)

    -u <rank>
    --uprank <rank>
      If set, parent taxons, up to the given rank, will be added to
      the database.
      Valid ranks are:
        kingdom
        class
        order
        family
        genus     [default value]
        species

Add taxons from an external DB

Usage:

	biodv tax.db.fill -e|--extern <database> [-u|--uprank <rank>]
		[<name>]

Command tax.db.fill adds additional taxons from an external DB to the
current database. By default only synonyms (of any rank), and children
(of taxa at or below species level) will be added.

The option -u or -uprank will add only parents up to the given rank.

Options are:

    -e <database>
    --extern <database>
      A required parameter. It will set the external database.
      Available databases are:
        gbif	GBIF webservice (requires internet connection)

    -u <rank>
    --uprank <rank>
      If set, parent taxons, up to the given rank, will be added to
      the database.
      Valid ranks are:
        kingdom
        class
        order
        family
        genus
        species

    <name>
      If set, only the indicated taxon, and its descendants (or its
      parents, in the case the -u or –unprank option is defined) will
      be filled.

Synchronize the local DB to an external taxonomy

Usage:

	biodv tax.db.sync -e|--extern <database> [<name>]

Command tax.db.sync synchronize two taxonomies (i.e. made it compatible),
one external and the local DB.

It require that the local DB has already assigned the external IDs. The
process will only add required taxons (for example, the parent of a
synonym). Taxons without an external ID will be left untouched.

Options are:

    -e <database>
    --extern <database>
      A required parameter. It will set the external database.
      Available databases are:
        gbif	GBIF webservice (requires internet connection)

    <name>
      If set, only the indicated taxon, and its descendants will
      be synchronized.

Update taxon information from an external DB

Usage:

	biodv tax.db.update -e|--extern <database> [-m|--match] [<name>]

Command tax.db.update reads an external database and update the
additional fields stored on the external database. Neither the name,
nor the rank, the parent of the correct-synonym status will be
modified.

If option -m, or --match is used, only the external ID will be set.

If a taxon is defined, then only that taxon and its descendants will
be updated.

Options are:

    -e <database>
    --extern <database>
      A requiered parameter. It will set the external database.
      Available databases are:
        gbif	GBIF webservice (requires internet connection)

    -m
    --match
      If set, only the external ID, and no other data, will be stored.

    <name>
      If set, only the indicated taxon, and its descendants, will be
      updated.

Eliminate a taxon from the database

Usage:

	biodv tax.del [-r|--recurse] <name>

Command tax.del removes a taxon from the database. By default only
the indicated taxon will be deleted, and both children and synonyms
will be moved to the parent of the taxon.

If the option -r or --recurse is used, then the taxon, and all of
its descendants will be deleted.

If the taxon is attached to the root, its synonyms will be also
deleted.

Options are:

    -r
    --recurse
      If set, the indicated taxon, as well of all of its
      descendants, will be eliminated from the database.

    <name>
      The taxon to be deleted.

Synonymize rankless taxa

Usage:

	biodv tax.format [<name>]

Command tax.format search for all unranked taxons (except from taxons
attached to the root), and make them synonyms, so the taxonomy will only
have as correct names, names with some rank. This is useful to, for
example, collapse subspecies as synonyms of a species.

Options are:

    <name>
      If defined, only the descendants of the indicated taxon will be
      searched and synonymized.

Print taxon information

Usage:

	biodv tax.info [--db <database>] [--id <value>] [<name>]

Command tax.info prints the information data available for a taxon name, in
a given database.

Either a taxon name, of a database id, should be used.

Options are:

    -db <database>
    --db <database>
      If set, the indicated database will be used to extract the taxon
      information.
      Available databases are:
        biodv	default database (on current directory)
        gbif	GBIF webservice (requires internet connection)

    -id <value>
    --id <value>
      If set, the information of the indicated taxon will be printed.

    <name>
      If set, the information taxon with the given name will be printed,
      if the name is ambiguous, the ID of the ambigous taxa will be
      printed.

Print a list of taxons

Usage:

	biodv tax.list [--db <database>] [--id <value>] [-m|--machine]
		[-p|--parents] [-s|--synonym] [-v|--verbose] [<name>]

Command tax.list prints a list of the contained taxa of a given taxon
in a given database.

If no taxon is defined, the list of taxons attached to the root of
the database will be given.

If the option synonym is defined, instead of the contained taxa, a
list of synonyms of the name will be given.

Only names will be printed, if the option machine is defined, only IDs
will be printed, and verbose option will print ID, taxon name and
taxon author (if available).

Options are:

    -db <database>
    --db <database>
      If set, the indicated database will be used to produce the taxon
      list.
      Available databases are:
        biodv	default database (on current directory)
        gbif	GBIF webservice (requires internet connection)

    -id <value>
    --id <value>
      If set, the list will be based on the indicated taxon.

    -m
    --machine
      If set, only the IDs of the taxons will be printed.

    -p
    --parents
      If set, a list of parents of the taxon will be produced.

    -s
    --synonyms
      If set, a list of synonyms of the taxon, instead of contained
      taxa, will be produced.

    -v
    --verbose
      If set, the list will produced indicating the ID, the taxon
      name, and the author of the taxon.

    <name>
      If set, the list will be based on the indicated taxon, if the
      name is ambiguous, the ID of the ambigous taxa will be printed.

Change a taxon parent

Usage:

	biodv tax.move [--to <name>] [-s|--status <value>] <name>

Command tax.move changes moves the taxon to a new parent. If the -s,
or --status option is not defined, it will move the taxon with their
current status. The movement must be compatible with the current
taxonomy, or an error will be produced.

If a taxon is set as a synonym (either with the –status option or
because is already a synonym), it should have a defined parent (with
the --to option), as a synonym can not be attached to the root of
the taxonomy. If the moved taxon have descendants, all of the
descendants will be moved as children of the new parent.

Only the correct/valid taxons can be attached to the root.

Only a correct/valid taxon can be a parent.

Options are:

    -to <name>
    --to <name>
      Sets the new parent of the taxon. If no parent is set, the
      taxons will be moved to the root of the taxonomy.

    -s <value>
    --status <value>
      Sets the new status of the taxon. If no status is given, the
      current taxon status will be used.
      Valid values are:
        correct   the taxon name is correct
        synonym   the taxon name is a synonym
        accepted  equivalent to correct
        valid     equivalent to correct
        true      equivalent to correct
        false     equivalent to synonym

    <name>
      The taxon to be moved. This parameter is required.

Change a taxon rank

Usage:

	biodv tax.rank [-r|--rank <rank>] <name>

Command tax.rank sets a new rank to a given taxon. If no rank is
defined, it will set the taxon as unranked. The new rank should be
compatible with the current taxonomy.

Options are:

    -r <rank>
    --rank <rank>
      Sets the new rank of the taxon.
      Valid ranks are:
        unranked  (default)
        kingdom
        class
        order
        family
        genus
        species

    <name>
      The taxon to be reranked. This parameter is required.

Set a taxon data value

Usage:

	biodv tax.set -k|--key <key> [-v|--value <value>] <name>

Command tax.set sets the value of a given key for the indicated taxon,
overwriting any previous value. If value is empty, the content of the
key will be eliminated.

Command tax.set can be used to set almost any key value, except the
taxon name (that can not be changed), the rank (use tax.rank), and the
parent and status (correct/valid or synonym) of the taxon
(use tax.move).

Except for some standard keys, no content of the values will be
validated by the program, or the database.

Options are:

    -k <key>
    --key <key>
      A key, a required parameter. Keys must be in lower case and
      without spaces (it will be reformatted to lower case, and spaces
      between words replaced by the dash ‘-’ character). Any key can
      be stored, but the recognized keys (and their expected values)
      are:
        author     to set the taxon’s author.
        extern     to set the ID on an external database, it will be on
                   the form "<service>:<id>", if only “<service>:” is
                   given, the indicated external ID will be eliminated.
        reference  to set a bibliographic reference.
        source     to set the ID of the source of the taxonomic data.
      For a set of available keys of a given taxon, use tax.value.

    -v <value>
    --value <value>
      The value to set. If no value is defined, or an empty string is
      given, the value on that key will be deleted.

    <name>
      The taxon to be set.

Validate a taxonomy database

Usage:

	biodv tax.validate

Command tax.validate validates a taxonomy database. It is useful to test
if a biodov database from a third party is correct.  If there are no
errors, it will finish silently.

Get a taxon data value

Usage:

	biodv tax.value [--db <database>] [--id <value>]
		[-k|--key <key>] <name>

Command tax.value prints the value of a given key for the indicated
taxon. If no key is given, a list of available keys for the indicated
taxon will be given.

Options are:

    -db <database>
    --db <database>
      If set, the indicated database will be used to search the
      indicated taxon.
      Available databases are:
        biodv	default database (on current directory)
        gbif	GBIF webservice (requires internet connection)

    -id <value>
    --id <value>
      If set, then it will search the indicated taxon.

    -k <key>
    --key <key>
      If set, the value of the indicated key will be printed.

    <name>
      If set, the indicated taxon will be searched. If the name is
      ambiguous, the ID of the ambiguous taxa will be printed.

Taxonomy database

In biodv the taxonomy database is stored in taxonomy sub-directory in
the file taxonomy.stz. The file is an stanza-encoded file.

The following fields are recognized by biodv taxonomy commands:

	name       canonical name of the taxon.
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


*/
package main
