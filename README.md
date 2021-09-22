# match-cli
[![CircleCI](https://circleci.com/gh/Optable/match-cli/tree/main.svg?style=svg)](https://circleci.com/gh/Optable/match-cli/tree/main)
[![Go Report Card](https://goreportcard.com/badge/github.com/optable/match-cli)](https://goreportcard.com/report/github.com/optable/match-cli)
[![GoDoc](https://godoc.org/github.com/optable/match-cli?status.svg)](https://godoc.org/github.com/optable/match-cli)

An open-source Command Line Interface (CLI) utility written in Golang to allow any *partner* of an Optable Data Connectivity Node (DCN) user to perform secure matches with the DCN. 

The match-cli tool enables anyone without access to an Optable DCN (_external partners_) to create and run a secure private set intersection (PSI) match with a DCN customer using the open-source [match](https://github.com/Optable/match) libary. Both parties will run the [DHPSI](https://github.com/Optable/match/blob/main/pkg/dhpsi/README.md) protocol by default to ensure that no private data is leaked during the match. 

## Build
You can build the latest `match-cli` binary by running the following comamnd:
```bash
# clone the repo and go to the directory
git clone https://github.com/Optable/match-cli.git && cd match-cli

# compile
make
# or more specifically
make build
```
The compiled binary will be located in `bin/match-cli`.

## Commands
The `match-cli` tool provides two subcommands. The `partner` subcommand connects to a DCN to match with and identifies the sender (`match-cli` operator) as an external partner. The `match` subcommand creates a match attempt and performs the secure intersection protocol. For each subcommand, use the `--help` flag to see detailed help messages and available options.

Note that it's not currently possible to be a secure match *receiver* using the `match-cli` utility. To receive secure matches you currently must have access to an Optable DCN.

Additional documentation is available [here](https://app.gitbook.com/@optable/s/optable-documentation/guides/match-cli).

## Example
To perform a secure PSI match with a DCN, you must first obtain an `<invite-code>` from the DCN's operator. The `<partner-name>` below is used to identify the DCN you are connecting with for subsequent match operations.
```bash
$ bin/match-cli partner connect <partner-name> "<invite-code>"
```

After successful partnering, a match can be created. You can use `<match-name>` to identify and manage matches. A `match_uuid` will be displayed in a JSON-formatted output once the match is succefully created.
```bash
$ bin/match-cli match create <partner-name> <match-name>
$ {"match_uuid":"UUID"}
```

Note that you are not required to save the `<match_uuid>`, you can run the following command to retrieve it later:
```bash
$ bin/match-cli match list <partner-name>
$ {"match_uuid":"UUID","name":"<match-name>"}
```
You can then run a match with an input file that contains matchable identifiers as follows:
```bash
$ bin/match-cli match run <partner-name> <match_uuid> <path-to-file>
```
Upon successful execution of the match, the number of the matching identifiers will be returned by the remote DCN in a JSON-formatted string.
```bash
{"time":"YYYY-MM-DDTHH:MM:SS.000000Z","id":"UUID","state":"completed","results":{"emails":<intersection-size>}}
```
