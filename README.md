# mod-reporting

Copyright (C) 2023 Index Data

This software is distributed under the terms of the Apache License,
Version 2.0. See the file "[LICENSE](LICENSE)" for more information.


## Overview

`mod-reporting` is a plug-compatible replacement for [`mod-ldp`](https://github.com/folio-org/mod-ldp), using the same API specification (FOLIO module descriptor, RAML file and associated JSON Schemas and examples). It provides the same interface (`ldp-query`) with the same semantics.

`mod-reporting` is written in Go, and compilation is controlled by a good old-fashioned [`Makefile`](Makefile). To build, just run `make`; to start the module running locally, use `make run`.

For the time being, this project's GitHub issue tracker will be used; once the software is ready for prime time, we will likely switch to Jira.

For now, see [`mod-ldp`'s README](https://github.com/folio-org/mod-ldp#readme) for further details.


## Author

Mike Taylor, [Index Data ApS](https://www.indexdata.com/).
mike@indexdata.com

