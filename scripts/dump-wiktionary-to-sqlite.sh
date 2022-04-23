#! /bin/bash

#
# dump-wiktionary-to-sqlite.sh
#
# Using the wdictosqlite tool, fetch and convert the current dump of Wiktionary.
# Creates and potentially overwrites a file enwiktionary-latest-pages-articles.sqlite3
# in the current working directory.
#
# You will need curl, bunzip, gzip and wdictosqlite in your PATH.
#

set -euo pipefail

script_dir="$(dirname "$(readlink -f "$0")")"

source_addr=https://dumps.wikimedia.org/enwiktionary/latest/enwiktionary-latest-pages-articles.xml.bz2
copying_file="$script_dir/default-copying.txt"
out_file="$script_dir/enwiktionary-latest-pages-articles.sqlite3"

curl --silent "$source_addr" | bunzip2 | wdictosqlite -copying "$copying_file" -outfile "$out_file"
gzip "$out_file"
