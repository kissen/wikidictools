#! /bin/sh

#
# dump-wiktionary-to-sqlite.sh
#
# Using the wdictosqlite tool, fetch and convert the current dump of Wiktionary.
# Creates and potentially overwrites a file enwiktionary-latest-pages-articles.sqlite3
# in the current working directory.
#
# You will need curl, bunzip and wdictosqlite in your PATH.
#

set -eu

source_addr=https://dumps.wikimedia.org/enwiktionary/latest/enwiktionary-latest-pages-articles.xml.bz2
outfile=enwiktionary-latest-pages-articles.sqlite3

curl --silent $source_addr | bunzip2 | wdictosqlite -outfile $outfile
