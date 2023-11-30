# mod-reporting

Copyright (C) 2023 Index Data

This software is distributed under the terms of the Apache License,
Version 2.0. See the file "[LICENSE](LICENSE)" for more information.


## Overview

`mod-reporting` is a plug-compatible replacement for [`mod-ldp`](https://github.com/folio-org/mod-ldp), using the same API specification (FOLIO module descriptor, RAML file and associated JSON Schemas and examples). It provides the same interface (`ldp-query`) with the same semantics.

**Bulding and running.**
`mod-reporting` is written in Go, and compilation is controlled by a good old-fashioned [`Makefile`](Makefile). To build, just run `make`; to start the module running locally, use `make run`.

**Personal data.**
This module does not store any personal data. See the file [`PERSONAL_DATA_DISCLOSURE.md`](PERSONAL_DATA_DISCLOSURE.md) for details.

**Contributing.**
See the file [`CONTRIBUTING.md`](CONTRIBUTING.md) for details.

**Issues.**
For the time being, [this project's GitHub issue tracker](https://github.com/indexdata/mod-reporting/issues) will be used; once the software is ready for prime time, we will likely switch to Jira.

For now, see [`mod-ldp`'s README](https://github.com/folio-org/mod-ldp#readme) for further details.


## API note

In the response from `/ldp/db/reports`, there is a numeric element `totalRecords`. Note that this is a count of the number of records included in the `records` array -- _not_ the total number of hits in the database. (That information is not available from PostgreSQL; the provided field is redundant, and would have been better omitted, but we retain it for backwards compatibility.)


## Environment

In normal operation, mod-reporting determines which underlying reporting database to connect to on the basis of the information configured in the mod-settings record with scope `ui-ldp.config` and key `dbinfo`, as managed by the "Database configuration" settings page of the Reporting app. However, these configured settings can be overridden if mod-reporting is run with all three of the following environment variables set:

* `REPORTING_DB_URL` -- The full URL of the PostgreSQL database to connect to, e.g. `postgres://localhost:5432/metadb`
* `REPORTING_DB_USER` -- The name of the PostgreSQL user to act as when accessing this database
* `REPORTING_DB_PASS` -- The password to use for nominated user


## See also

* [The `ldp-config-tool` utility](config-tool)


## Author

Mike Taylor, [Index Data ApS](https://www.indexdata.com/).
mike@indexdata.com


