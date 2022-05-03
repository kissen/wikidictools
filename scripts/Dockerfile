#
# This Dockerfile pull the latest Wiktionary dump and converts it to
# a sqlite dump.
#

FROM index.docker.io/library/golang:1.18-alpine

RUN mkdir -p /build
WORKDIR /build

RUN apk update && apk add bash build-base bzip2 curl git

# Install the command line utility used for generating the dictionary. This
# will already install into PATH.
RUN go install github.com/kissen/wikidictools/wdictosqlite@latest

# Now we clone the Git repository and run the script.
RUN git clone https://github.com/kissen/wikidictools.git /build/wikidictools
RUN /build/wikidictools/scripts/dump-wiktionary-to-sqlite.sh

ECHO Created /build/wikidictools/scripts/enwiktionary-latest-pages-articles.sqlite3. Run
ECHO   cp <CONTAINER_ID>:/build/wikidictools/scripts/enwiktionary-latest-pages-articles.sqlite3 .
ECHO to extract the created dictionary file.
