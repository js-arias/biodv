# Biodv

Biodv is a collection of packages
and tools
for management
and analysis of biodiversity data.

## Database organization

Biodv code and tools,
requires the data to be organized in an specific directory
and file hierarchy
within the current working path
(which is defined hera as a “project”).
Each sub-directory stores the data
in one or more files using the stanza format.

Here is an example of a project sub-directories:

~~~
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
~~~

Not all sub-directories are required,
as they will be created given
the needs of each project.

Both tools,
and biodv packages look for the database files automatically,
so most of the time,
a user,
or a developer,
does not need to know how each file is called,
or where is stored.

## Stanza format

In biodv data is stored using the stanza format.
The stanza format is an structured UTF-8 text file
with the following rules:

1. Each line containing a field must starts with the field name
   and separated from its content by a colon ( : ) character. 
   If the field name ends with a new line rather than a colon,
   the field is considered as empty.
2. Field names are case insensitive
   (always read as lower caps),
   without spaces,
   and should be unique.
3. A field ends with a new line.
   If the content of a field extends more than one line,
   the next line should start with at least one space
   or tab character.
4. A record ends with a line that starts with percent sign ( % ) character.
   Any character after the percent sign will be ignored
   (usually %% is used to increase visibility of the end-of-record).
5. Lines starting with the sharp symbol ( # ) character
   are taken as comments.
6. Empty lines are ignored.

Here is a simple example of an stanza file containing taxonomic data:

~~~
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
~~~

Stanza file format was inspired by the
[record-jar/stanza format](http://www.catb.org/esr/writings/taoup/html/ch05s02.html#id2906931)
described by E. Raymond "*The Art of UNIX programming*"
(2003)
Addison-Wesley,
and C. Strozzi
"*[NoSQL list format](http://www.strozzi.it/cgi-bin/CSA/tw7/I/en_US/NoSQL/Table%20structure)*"
(2007).

## Authorship and license

Biodv is distributed under a BSD2 license
that can be found in the LICENSE file.
For a list of authors,
see the AUTHORS file.

