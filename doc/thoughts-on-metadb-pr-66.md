# Thoughts on MetaDB pull-request 66

Mike Taylor, Index Data <mike@indexdata.com>


<!-- md2toc -l 2 thoughts-on-metadb-pr-66.md -->
* [Introduction](#introduction)
* [Minor changes in various files](#minor-changes-in-various-files)
* [Major changes in `poll.go`](#major-changes-in-pollgo)
* [Appendix A. Things to do in the MetaDB source code](#appendix-a-things-to-do-in-the-metadb-source-code)
* [Appendix B. Changes to make to user documentation](#appendix-b-changes-to-make-to-user-documentation)


## Introduction

Pull-request 66 on MetaDB is
[Add parallel read from kafka](https://github.com/metadb-project/metadb/pull/66/files),
submitted by [nazgaret](https://github.com/nazgaret)
on behalf of EPAM,
working for EBSCO,
supporting a university library customer (Cornell?).
Its purpose is to increase throughput of Kafka events to address performance problems with very slow update of MetaDB.

My thoughts on the individual proposed changes follow.


## Minor changes in various files

* **cmd/metadb/dbx/dbx.go line 167**.
It's probably a mistake that the `NewPool` function sets a hardwired `MaxConns` value of 20, and that mistake is not really addressed by changing the hardwired value to 40. This should be configurable. However, this doesn't seem like a radical change that could break anything.
  * Learning point: how is configuration done? I see from other code that there is a configuration file, but I don't know if there is also a set of environment variables, or a known location for configuration information within the database.

* **cmd/metadb/util/util.go line 221**.
Similar comments apply to the old and new hardwired values for `checkpointSegmentSize`. Since this value can already be overridden in configuration, changing the default value seems unnecessary.

* **doc/serveradmin.adoc lines 9-10**.
I don't understand on what basis these hardware requirements have been bumped, but maybe they just reflect real-life experience.
  * Learning point: why does the MetaDB documentation use AsciiDoc rather than (like all our other software) Markdown?

* **doc/serveradmin.adoc line 22**.
There seems to be no reason to bump the required version of Go up to 1.22. The top-level `go.mod` file is definitive, and that says 1.21.

* **go.mod** and **go.sum**.
One new dependency is added (`golang.org/x/sync`), which is used by the new code. So that's reasonable. But upgrades of several other modules are also included, and it's not clear to me whether they are necessary. They are all minor version upgrades, though, so probably harmless. Nassib may have a view.

This leaves the more substantial changes in **cmd/metadb/server/poll.go**


## Major changes in `poll.go`

* It is really not on that `createConsumers` hardwires the number of consumers to 40 (or any specific value).


## Appendix A. Things to do in the MetaDB source code

As I've been reading the code to understand the implications of the pull request, other issues are becoming apparent. Here I am making very brief notes so I can return to them when propitious.

* Progressively add developer documentation as I start to understand the source code -- the documentatiin that I wish I'd had available.
* Reconsider auto-formatting of `.adoc` files, in which line breaks are currently hard. Somthing about this [touches on one-space-between-sentences](https://indexdata.slack.com/archives/D0HP2MQ11/p1712755034675339).


## Appendix B. Changes to make to user documentation

Again, these observations arise from my reading of the user documentation, coming to it as a newbie. Some of the issues raised here are merely requests for clarification, and should be addressed just in the documentation; but some may turn out to be questions about implementation decisions.

* In section 1.1L "The data contained in the Metadb database originally come from another place: a *data source*." Multiple sources can contribute to a single MetaDB store. We don't presently have real-world configurations that do this, but a theoretical example where this would be useful might be if a FOLIO instance has their module storage backends split across different database instances.

* Leaping straight from "1.1. Getting started" to "1.2. Main tables" is bewildering. We really need more discussion of the concepts in between these sections. For example, a MetaDB database _is_ a Postgres database that is expected to be used in the same ways as other Postgres databases; that the same `psql` client is used to access both Postgres and MetaDB, but the latter does not fully proxy to the former; that certain GUI tools can be used for querying; that records have start and end datetimes which bound the period of their relevance in historic data.

* The "Main tables", "Current tables" and "Transformed tables" sections should be subsections of a higher-level section that describes the concepts in a high-level way and outlines the naming conventions (including transformed main tables such as `patrongroup__t__`).

* The example in section 1.3 shows that the `__id`, `__start` and `__origin` columns exist in current tables (as in main tables). But what about `__end` and `__current`?

* In section 1.4, "In the current version of Metadb, only top-level, scalar JSON fields are extracted into transformed tables." We need to explain what non-top-level and non-scalar fields are. Section 4.1.4.4 implicitly explains this when it discusses extracting array values from JSONB fields.

* "Note that JSON data are treated as "schemaless," and fields are inferred from their presence in the data rather than read from a JSON schema.  As a result, a column is only created from a JSON field if the field is present in at least one JSON record." More precisely, the table definition is created dynamically continually: _each_ record added will result in the addition of a new column if a JSON field in that record contains a field not previously seen.

* I wondered if there there was any particular reason why the names of special tables _end_ with `__`, as in `patrongroup__`, while the names of special fields _begin_ with `__`, as in `__start`? The reason is that all of the related tables are grouped together when sorting alphabetically.

* In section 1.7: "For monetary amounts, `numeric(19, 4)` is usually a good choice.  For exchange rates, `numeric(19, 14)` may be used." This is standard practice to avoid round-off errors in financial accounting databases.

* Section 1.8. on Creating reports should include information on publishing them to GitHub for use in the FOLIO Reporting app.

* Section 1.10 explains how a running MetaDB instance can provide information about itself, such as its version number, when the various tables were last updated, system logging messages and query status. It would be helpful to expose at least some of this in the Reporting app. I have filed [a UILDP issue in Jira](https://folio-org.atlassian.net/browse/UILDP-148) and [a mod-reporting issue in the GitHub tracker](https://github.com/indexdata/mod-reporting/issues/66) so we don't forget about this.

* Section 2.1. Data type conversion refers to on-the-fly adjustments to the definitions of existing tables based on ingesting new data. The table in section 2.1.1 is about how existing columns get their types “promoted” to enable them to encompass values of both the old and new records’ types. MetaDB decides which conversion to apply on the basis of what type can support the types of all values provided in records.

* In section 2.4, "Metadb allows scheduling external SQL files to run on a regular basis." That is all hard-coded at the moment. The word "scheduling" there is misleading because at the moment there are no configuration knobs. When  a data source is configured, the `module` setting optionally specifies `folio` or `reshare`.  If one of those is specified, it enables the external SQL with specific hardcoded values.  That is the only configuration for external SQL at the moment. A relatively high priority feature would be to add configuration knobs to that, including the Git repository, tag, path, etc., and time of day and frequency when the SQL should run.  Ideally, this will work best if fully generalized to allow arbitrary scheduling of SQL.  The best way to describe this is a "job scheduler" and there is [a placeholder issue for it](https://github.com/metadb-project/metadb/issues/43).

* "Any tables created should not specify a schema name." It seems that this is friendly advice rather than a prohibition. At the moment running external SQL is narrowly focused on specific community practices in FOLIO and ReShare, and it isn't as generalized as some other features. So for example it points to a specific repository that is managed by the reporting community in those projects.  So there may be an assumption here or there that you wouldn't want to make if you allow it to be pointed at any SQL. Current FOLIO/ReShare customers are not the only potential users of MetaDB. We contemplate using it in other applications, for example to analyse collections in [Reservoir](https://github.com/folio-org/mod-reservoir). Such new applications could well be the driver for prioritizing condifurability of scheduled SQL queries.

* In section 2.4.1, "The --metadb:table directive declares that the SQL file updates a specific table.  This allows Metadb to report on the status of the table." It does this by maintaining the the `metadb.table_update` table described in section 2.3.3.

* In section 2.5, "These statements are only available when connecting to the Metadb server (not the database)." i.e. when `psql` or a similar program is targeted that the Metadb server rather than the underlying Postgres server.

* Section 2.5.2. ALTER TABLE: how is this different from regular `ALTER TABLE`? We know that in general Postgres-level DDL commands should not be run on MetaDB's tables. The MetaDB-level version runs extra checks.

* The section on "CREATE DATA ORIGIN" does not have a section number, and so appears as part of section 2.5.3 on AUTHORIZE.

* Data origins form essentially a controlled vocabulary for the `__origin` column in tables. They are used primarily in ReShare to indicate which tenant in a consortium a given row was taken from. (When using Metadb for FOLIO, each tenant’s data goes into a separate database; but when using it for ReShare, all the tenants of a consortium go into a single shared database. ReShare internally uses FOLIO-style tenants to represent the tenants in a consortium, and Metadb combines the tables from all the tenants into a single table and fills in `__origin` to tell users which tenant a record belongs to.)

* More detail on how multiple ReShare tenants are handled:
> When configured for ReShare, Metadb typically uses a single data source for the whole consortium, but it combines data from mutiple tenants into aggregated tables. Each record in these tables has an `__origin` field set to the name of the tenant from which it was derived. The ReShare data source does this by matching the start of the schema names of incoming records against the set of origins that have been defined by `CREATE DATA ORIGIN`. (This is conceptually defined in a `reshare` module, but the code has not at this point been split out into a modular system.)
>
> For example if the source ReShare database has `east_mod_rs.table1` and `west_mod_rs.table1`, the data in Metadb will be in `reshare_rs.table1` and each row's `__origin` will contain the value `east` or `west` to indicate which tenant the data came from.
>
> (The Kafka stream feeding a ReShare data source is configured to return records from all tenants. So configuring that streams is a whole nother aspect that’s different for FOLIO and ReShare.)
>
> The `__origin` field could be used for other things in non-ReShare systems. More generally, origin is intended as an alternative category to source.  Source may be just an accident of how data are collected, e.g. geographically in a sensor network. Origin is a way of tagging related data that come from different sources. It's an opaque string but it could potentially have structure like `a.b.c`.
>
> One application of this is that it’s perfectly possible for a ReShare consortium’s tenants to be hosted by multiple FOLIO instances. We would then set up MetaDB to use multiple sources, one per FOLIO instance (or, more precisely, one per Kafka that is listening to a FOLIO instance), and each of those sources would contribute records to the same MetaDB ReShare schema but provide different sets of values for those records’ `__origin` fields.

* "When the initial snapshot has finished streaming, the message "source snapshot complete (deadline exceeded)" will be written to the log." (The "deadline exceeded" exceeded here refers to a timeout when waiting for further change events. So it is communicating that the update has completed, not that it has been abandoned due to running out of allocated time.)

* Reading section 2.5.4. CREATE DATA SOURCE, it's quickly apparent that we need to write a narrative guide on setting up Kafka for a FOLIO or ReShare installation, then creating MetaDB data sources for it.

* In section 2.5.5, "CREATE USER defines a new database user that will be managed by Metadb." At the moment it doesn't mean much for MetaDB to manage the user, but the intent is that some database objects, such as user accounts, should not be modified directly by admins. Metadb-created tables fall into that category, and so do users. In principle there is a clear line between objects managed by Metadb and other objects.

* In section 3.1.1 (Hardware requirements), “Architecture: x86-64 (AMD64)“. Why does MetaDB care what architecture CPU it’s compiled for? Metadb is tested primarily on x86-64.  DevOps will want to try running it on ARM64, which is relatively new and not fully supported (or was not fully supported at last check) in some of the libraries.  Things like atomics used for spinlocks and also the Kafka client library.  This should be revisited once in a while to see if we can test on and support ARM64.

* In section 3.1.2 (Software requirements), "Operating system: Debian 12 or later". This doesn't seem to be saying that Debian is _required_ (i.e. I won't be able to run MetaDB on my MacBook), just that it's the primary development OS. I will confirm this when I try to build and run my own MetaDB service.

* Section 3.3. Server configuration should be preceded by a section on the different ways to invoke the `metadb` binary as a command-line tool or server.

* In section 3.3, "superuser = postgres // superuser_password = zpreCaWS7S79dt82zgvD". Why does MetaDB need superuser access to Postgres? It escalates to superuser for a very few things. One of them is creating users.

* In section 3.4 (Backups): "In general persistent data are stored in the database, and so the database should be backed up often." This should have a link to how this is best done for Postgres.

* In section 3.7, it is implied that a MetaDB server _is_ a Postgres server, but with extensions. That's not right, though. It's close enough that we can and do use the `psql` command-line client to communicate with MetaDB instead of the bespoke legacy client `mdb`, but MetaDB does not proxy all unrecognized commands through to the underlying Postgres.

* In section 3.8.1, "Metadb currently supports reading Kafka messages in the format produced by the Debezium PostgreSQL connector for Kafka Connect." Links are needed here ... and I have a lot of learning to do.

* This Section should open with a diagram showing the flow of data from the FOLIO UI via Okapi into the underlying Postgres database, through Debezium (via Kafka Connect?) to Kafka and on from there into the MetaDB database as a source where it is analyzed by queries from a command-line or GUI client.

* In that section "A source PostgreSQL database" is presumably a FOLIO or ReShare database in our primary use cases.

* In section 3.8.2 (Creating a connector), "... by setting `wal_level = logical` in `postgresql.conf`." This is in the source (FOLIO) Postgres database, not Metadb's database.

* Sections 3.8.3 (Monitoring replication) and 3.8.6 (Deleting a connection) are completely opaque to me. When I understand them, I will expand them.

* In section 4.1.1, "Metadb transforms MARC records from the tables `marc_records_lb` and `records_lb` in schema `folio_source_record` to a tabular form which is stored in a new table, `folio_source_record.marc__t`." This set of schema and table names is hardwired and generally tailored for FOLIO and its reporting community. They share queries and wouldn't want variance in the table names.

* "Only records considered to be current are transformed, where current is defined as having `state = 'ACTUAL'` and an identifier present in `999 ff $i`." This is an established FOLIO-specific convention for marking a MARC record as “current”. The FOLIO backing store record needs to have a `state` field with value `ACTUAL` and the MARC record that is carried by that FOLIO record needs to have a 999 field with both indicators set to `f` and whose `$i` subfield is non-empty.

* "The MARC transform stores partition tables in the schema `marctab`." This refers to so-called "horizontal partitioning", which is used by almost all tables in MetaDB, but which we can usually ignore -- as we can in this case, using instead the virtual union table `folio_source_record.marc__t`.

* In section 4.1.2, "FOLIO "derived tables" are automatically updated once per day, usually at about 3:00 UTC by default." The document does not introduce the term "derived table". These are helper tables for commonly used joins etc. They are created and updated by the Scheduled Query facility discussed above, executing SQL written by the community and kept in [the `sql_metadb/derived_tables/` area of the `folio-analytics` repository](https://github.com/folio-org/folio-analytics/tree/main/sql_metadb/derived_tables). the list of queries to run is held in [`runlist.txt`](https://github.com/folio-org/folio-analytics/blob/main/sql_metadb/derived_tables/runlist.txt).

* Section 4.1.3. (Data model) should talk more about what can be known about the FOLIO-derived tables, as well as noting what is not documented. And maybe also list some specific important tables.

* In section 4.1.4.1. "Table names have changed and now are derived from FOLIO internal table names". The bigger change here seems to be that the tables are spread across many different schemas. The schema names come from the source database. LDP Classic was intended to be part of FOLIO and to provide reporting features for FOLIO, so it mapped FOLIO's _schema_._table_ pairs to its own favoured names. But Metadb is general-purpose and works more or less with any source database, so it has dropped the FOLIO-specific mapping step.

* In section 4.1.4.5, "Note that JSON data contained in the imported records are not transformed into columns." (There is no way to trigger this transformation after the import is complete.)

* Section 4.1.5 (Configuring Metadb for FOLIO) should come much earlier in the document. It raises several questions:
  * "use the `module 'folio'` option" -- Section 2.5.4 on CREATE DATA SOURCE says that `module` specifies "Name of pre-defined configuration", but this is not merely a shortcut. It's doing specific things to do with how FOLIO and ReShare use tenants, as well as specifying what scheduled queries to run. For FOLIO, it also runs a MARC transform that is not used for ReShare. In the future there will likely be other differences.
  * "Set `trimschemaprefix` to remove the tenant from schema names": We do this because when we're interested in multiple FOLIO tenants, each tenant's data goes into an entirely separate MetaDB database.
  * "Set [...] `addschemaprefix` to add a `folio_` prefix to the schema names." This is done because we generally want the option of using MetaDB for _all_ a library's data analysis needs, not justthose related to FOLIO. For example, you might also make a data-source for (say) VuFind’s Solr database, and use it to maintain tables with names like `vufind_inventory.items` and `vufind_requests.requests`.
  * There is nothing in the `CREATE DATA SOURCE` statement to indicate what tenant's data we want. That information has to be included in the configuration of the nominated Kafka topic.
  * XXX "In the Debezium PostgreSQL connector configuration, the following exclusions are suggested [list]". Various tables are excluded for different reasons. Most of them are omitted just because they are not of interest (e.g. pubsub state) but data from some modules, e.g. `mod_login`, is omitted for security reasons. It is up to individual libraries to tailor this exclusion list to their requirements.

* In section 4.2.2 (Configuring Metadb for ReShare), "Before defining a ReShare data source, create a data origin for each consortial tenant". This means each tenant _in_ a consortium, not a tenant _that represents_ a consortium?

* We do not use `trimschemaprefix` when ingesting from ReShare, because the `reshare` module uses the tentant names in the prefixes to choose between the configured data origins.

* At the end of section 4.2.2, some backquote slippage results in a hunk of text being in code font.

* Section 4.3 (MARC transform for LDP) makes it clear that MARC transformation
for LDP Classic is done by an external program, `marct` (formerly known as `ldpmarc`). For MetaDB, though, the MARC transformation is integrated.

* We should link to https://github.com/folio-org/folio-analytics/wiki/Cookbook:-User-Defined-Function

