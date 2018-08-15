// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

/*
Package taxonomy implements
a hierarchical,
linnean ranked taxonomy.

In biodv the taxonomy database
is stored in taxonomy sub-directory
in the file taxonomy.stz.
The file is an stanza-encoded file.

Instead of accessing the file directly,
use one of these mechanisms:

If you want to read the file sequentially
(for example,
to just load the fields you need)
use the Scanner type,
creating it with function OpenScanner.

If you want to use a read-write taxonomy
use DB type
creating it with Open function.
To access an editable Taxon,
use TaxEd method,
and to access list of taxons
use TaxList method.
Changes on the taxonomic data
will be kept in memory,
and are only copied to hard disk with an explicit
call to Commit method.

If you want to use a biodv.Taxonomy,
you can open it using “biodv” as the driver name,
and creating it with the biodv.OpenTax function.

A Scanner can be used also
to read taxon data from an arbitrary stanza file,
this can be done by creating the Scanner
with the function NewScanner.

The stanza format understood by a taxonomy scanner
should have the following fields,
that should be accessed by a Taxon method:

	name       cannonical name of the taxon.
	rank       rank of the taxon.
	correct    if the taxon is not correct
	           (a synonym) is set to "false",
	           any other value is interpreted as "true"
	           (i.e. a correct name).
	parent     name of the taxon's parent,
	           if not present,
	           the taxon will be interpreted as attached
	           to the root of the taxonomy.

Other fields,
should be accessed by the Keys and Value methods of a biodv.Taxon.
The ones interpreted are:

	author     author of the taxon's name.
	extern     a list
	           (separated by spaces)
	           of external IDs
	           of the taxon,
	           it is expected to be in the form
	           <service>:<key>.
	reference  a bibliographic reference
	           (or an ID to that reference).
	source	   the database
	           (or its ID)
	           of the source of the taxonomic information.

Other fields,
if any,
can be defined by the user.

There are some constrains in the taxonomy file:

	(i)     if a taxon record has a parent,
	        this parent should be defined previously,
	        and be a correct name.
	(ii)    if a taxon is not a correct name
	        (i.e. a synonym),
	        it should have a parent.
	(iii)	ranks should be consistent within the taxonomy.

This constrains are assumed when using the Scanner,
as it is expected that the file conforms to this constrains.
In case of a untrusted database,
it can be validated by open it with DB type,
that report any violation found while reading the file.
*/
package taxonomy
