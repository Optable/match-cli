# match-cli

An open-source Command Line Interface (CLI) tool written in golang to allow non Optable customers to interact with the Optable Data Connectiveity Nodes (DCN) or a sandbox. 

The match-cli tool enables non DCN users (_adhoc partners_) to create and run a secure private set intersection (PSI) match with a DCN customer using the open-source [match](https://github.com/Optable/match) libary. Both parties will run the [DHPSI](https://github.com/Optable/match/blob/main/pkg/dhpsi/README.md) protocol by default to ensure that no private data is leaked during the match. 

## Example use cases
To perform a secure match with a DCN customer, user needs to first obtain an `<invite-code>` from a DCN customer, and then connect to the DCN partner.  `<partner-name>` below is used to identify and manage partners.
```bash
$ go run cmd/cli/main.go partner connect <partner-name> "<invite-code>"
```

After successful partnering, a match can be created. You can use `<match-name>` to identify and manage matches.
```bash
$ go run cmd/cli/main.go match create <partner-name> <match-name>
```
On success, a `match_uuid` will be displayed in a json formatted output.
```bash
{"match_uuid":"UUID"}
```
Note that you are not required to save the `<match_uuid>`, you can run the following command to retrieve it back.
```bash
$ go run cmd/cli/main.go match list <partner-name>
$ {"match_uuid":"UUID","name":"<match-name>"}
```
We can then run a match with an input file that contains matchable identifiers.
```bash
$ go run cmd/cli/main.go match run <partner-name> <match_uuid> <path-to-file>
```
Upon successful execution of the match, the size of the intersected identifiers will be returned in a json formatted string.
```bash
{"time":"YYYY-MM-DDTHH:MM:SS.000000Z","id":"UUID","state":"completed","results":{"emails":<intersection-size>}}
```

For each subcommands `partner` and `match`, you can use the flag `--help` to see detailed help messages.

Full documentation is available [here](https://app.gitbook.com/@optable/s/optable-documentation/guides/match-cli).