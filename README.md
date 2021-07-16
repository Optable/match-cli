# match-cli

An open-source Command Line Interface (CLI) tool written in golang to allow non Optable customers to interact with the Optable Data Connectiveity Nodes (DCN). 

The match-cli tool enables non DCN users to create and run a secure private set intersection (PSI) match with a DCN customer using the open-source [match](https://github.com/Optable/match) libary. Both parties will run the [DHPSI](https://github.com/Optable/match/blob/main/pkg/dhpsi/README.md) protocol by default to ensure that no private data is leaked during the match. 

## Example use cases
To perform a secure match with a DCN customer, user needs to first obtain an invite code from the DCN customer, and partner up with the DCN. `<partner-name>` below is used to identify and manage partners.
```bash
$ optable-match-cli partner connect <partner-name> "<invite-code>"
```

After successful partnering, a match can be created.
```bash
$ optable-match-cli match create <partner-name> <match-name>
```
On success, a `match_uuid` will be printed, and we will use this token to perform a match.

```bash
$ optable-match-cli match run <partner-name> <match_uuid> <path-to-file>
```
Upon successful execution of the match, the size of the intersected identifiers will be returned in a json formatted string.

For each subcommands `partner` and `match`, you can use the flag `--help` to see detailed help messages.

Full documentation is available [here](https://app.gitbook.com/@optable/s/optable-documentation/guides/match-cli).