# Thoughts on MetaDB pull-request 66

Mike Taylor, Index Data <mike@indexdata.com>


<!-- md2toc -l 2 thoughts-on-metadb-pr-66.md -->
* [Introduction](#introduction)
* [Minor changes in various files](#minor-changes-in-various-files)
* [Major changes in `poll.go`](#major-changes-in-pollgo)
* [Appendix A. Things to do in the MetaDB source code](#appendix-a-things-to-do-in-the-metadb-source-code)
* [Appendix B. Skeleton of developer documentation](#appendix-b-skeleton-of-developer-documentation)
* [Appendix C. Changes to make to user documentation](#appendix-c-changes-to-make-to-user-documentation)


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


## Appendix B. Skeleton of developer documentation

As noted in the first bullet point of Appendix A, I intend to add developer documentation reflecting my learning as I read the code and consult Nassib. The following skeleton is work in progress, of will be of no use to anyone but me.

* `mdb` is an old client for the `metadb` server.  It probably is legacy code and probably should be removed.

* The list of top-level features mentioned at the start of [the user documentation](https://metadb.dev/doc/) is probably worth using as a high-level overview of the code:
  * streaming data sources
    * probably multiple kinds of source?
  * data model transforms
  * historical data

* "The data contained in the Metadb database originally come from another place: a *data source*." I assume multiple sources can contribute to a single MetaDB store, but this is worth checking.


## Appendix C. Changes to make to user documentation

Again, these observations arise from my reading of the user documentation, coming to it as a newbie. Some of the issues raised here are merely requests for clarification, and should be addressed just in the documentation; but some may turn out to be questions about implementation decisions.

* Leaping straight from "1.1. Getting started" to "1.2. Main tables" is bewildering. We really need more discussion of the concepts in between these sections. For example, a MetaDB database _is_ a Postgres database that is expected to be used in the same ways as other Postgres databases; that certain GUI tools can be used for querying; that records have start and end datetimes which bound the period of their relevance in historic data.

* The "Main tables", "Current tables" and "Transformed tables" sections should be subsections of a higher-level section that describes the concepts in a high-level way and outlines the naming conventions (including transformed main tables such as `patrongroup__t__`).

* The example in section 1.3 shows that the `__id`, `__start` and `__origin` columns exist in current tables (as in main tables). But what about `__end` and `__current`?

* In section 1.4, "In the current version of Metadb, only top-level, scalar JSON fields are extracted into transformed tables." We need to explain what non-top-level and non-scalar fields are. Section 4.1.4.4 impliciy explains this when it discusses extracting array values from JSONB fields.

* "Note that JSON data are treated as "schemaless," and fields are inferred from
their presence in the data rather than read from a JSON schema.  As a result, a
column is only created from a JSON field if the field is present in at least
one JSON record." More precisely, the table definition is created dynamically continually: _each_ record added will result in the addition of a new column if a JSON field in that record contains a field not previously seen.

* Is there any particular reason why the names of special tables _end_ with `__`, as in `patrongroup__`, while the names of special fields _begin_ with `__`, as in `__start`? All of the related tables are grouped together when sorting alphabetically.

* In section 1.7: "For monetary amounts, `numeric(19, 4)` is usually a good choice.  For exchange rates, `numeric(19, 14)` may be used." Why?

* Section 1.8. on Creating reports should include information on publishing them to GitHub for use in the FOLIO Reporting app.

* Section 1.10 explains for a running MetaDB instance can provide information about itself, such as its version number, when the various tables were last updated, system logging messages and query status. It would be helpful to expose at least some of this in the Reporting app. I have filed [a UILDP issue in Jira](https://folio-org.atlassian.net/browse/UILDP-148) and [a mod-reporting issue in the GitHub tracker](https://github.com/indexdata/mod-reporting/issues/66) so we don't forget about this.

* Section 2.1. Data type conversion seems to be referring to on-the-fly adjustments to the definitions of existing tables based on ingesting new data. Am I interpreting that correctly? In the table in section 2.1.1, how does MetaDB decide which conversion to apply? This is about how existing columns get their types “promoted” to enable them to encompass values of both the old and new records’ types.

* In section 2.4, "Metadb allows scheduling external SQL files to run on a regular basis." How? Also, "any tables created should not specify a schema name." Why not? Are all these tables implicitly in an "external directives" schema?

* In section 2.4.1, "The --metadb:table directive declares that the SQL file updates a specific table.  This allows Metadb to report on the status of the table." How?

* In section 2.5, "These statements are only available when connecting to the Metadb server (not the database)." How does one do that?

* Section 2.5.2. ALTER TABLE: how is this different from regular `ALTER TABLE`? (Comments along the lines of "It differs from GRANT in thatthe authorization will also apply to tables created at a later time in the data source" in section 2.5.3 would be helpful here.)

* The section on "CREATE DATA ORIGIN" does not have a section number, and so appears as part of section 2.5.3 on AUTHORIZE. Also, since we don't know what a data origin is, and how it differs from a data source, this is not very informative.

* "When the initial snapshot has finished streaming, the message "source snapshot complete (deadline exceeded)" will be written to the log." What does the "deadline exceeded" part mean?

* Reading section 2.5.4. CREATE DATA SOURCE, it's quickly apparent that we need a narrative guide on setting up Kafka for a FOLIO or ReShare installation, then creating MetaDB data sources for it. This may exist further down the document -- if so it should be referenced here -- or may need to be written.

* In section 2.5.5, "CREATE USER defines a new database user that will be managed by Metadb." What does it mean for MetaDB to manage the user?

* In section 3.1.1 (Hardware requirements), “Architecture: x86-64 (AMD64)“. Why does MetaDB care what architecture CPU it’s compiled for? Metadb is tested primarily on x86-64.  DevOps will want to try running it on ARM64, which is relatively new and not fully supported (or was not fully supported at last check) in some of the libraries.  Things like atomics used for spinlocks and also the Kafka client library.  This should be revisited once in a while to see if we can test on and support ARM64.

* In section 3.1.2 (Software requirements), "Operating system: Debian 12 or later". Is that saying that Debian is _required_ (i.e. I won't be able to run MetaDB on my MacBook), or just that it's the primary development OS?

* Section 3.3. Server configuration should be preceded by a section on the different ways to invoke the `metadb` binary as a command-line tool or server.

* In section 3.3, "superuser = postgres // superuser_password = zpreCaWS7S79dt82zgvD". Why does MetaDB need superuser access to Postgres?

* In section 3.4 (Backups): "In general persistent data are stored in the database, and so the database should be backed up often." This should have a link to how this is best done for Postgres.

* In section 3.7, it seems that a MetaDB server _is_ a Postgres server, but with extensions. If that's correct, it's worth saying explicitly. If so, then reason the `mdb` command-line client is legacy code is probably that we use `psql` instead.

* In section 3.8.1, "Metadb currently supports reading Kafka messages in the format produced by the Debezium PostgreSQL connector for Kafka Connect." Links are needed here ... and I have a lot of learning to do.

* In that section "A source PostgreSQL database" is presumably a FOLIO or ReShare database in our primary use cases.

* In section 3.8.2 (Creating a connector), "... by setting `wal_level = logical` in `postgresql.conf`." This is in the source (FOLIO) Postgres database, not Metadb's database.

* Sections 3.8.3 (Monitoring replication) 3.8.6 (Deleting a connection) are completely opaque to me. When I understand them, I will expand them.

* In section 4.1.1, "Metadb transforms MARC records from the tables `marc_records_lb` and `records_lb` in schema `folio_source_record` to a tabular form which is stored in a new table, `folio_source_record.marc__t`." Is this set of schema and table names hardwired, or is there configuration for it?

* "Only records considered to be current are transformed, where current is defined as having `state = 'ACTUAL'` and an identifier present in `999 ff $i`." Is this a widespread notion of what it means for a MARC record to be “current”, or a FOLIO-specific convention?

* Also, what does the `ff` in `999 ff $i` mean?

* "The MARC transform stores partition tables in the schema `marctab`." What is a partition table in this context? It doesn't seem to be to do with horizontal or vertical partitioning of tables.

* "FOLIO "derived tables" are automatically updated once per day, usually at about 3:00 UTC by default." The document does not introduce the term "derived table". What are they? How are they configured?

* Section 4.1.3. (Data model) should talk more about what can be known about the FOLIO-derived tables, as well as noting what is not documented. And maybe also list some specific important tables.

* In section 4.1.4.1. "Table names have changed and now are derived from FOLIO internal table names". The bigger change here seems to be that the tables are spread across many different schemas. Where do the schema names come from? Are they simply copied from the schema names in the source database? Is it that LDP Classic used to map FOLIO's schema.table pairs to its own favoured names but that MetaDB has dropped that mapping step?

* In section 4.1.4.5, "Note that JSON data contained in the imported records are not transformed into columns." Is there a way to trigger this transformation after the import is complete?

* Section 4.1.5 (Configuring Metadb for FOLIO) should come much earlier in the document. It raises several questions:
  * "use the `module 'folio'` option" -- what exactly does this do? Section 2.5.4 on CREATE DATA SOURCE says that `module` specifies "Name of pre-defined configuration", but is this merely a shortcut, or is doing something distinctive of its own?
  * "Set `trimschemaprefix` to remove the tenant from schema names": why? Don't we want separate tenants' data to be separated in MetaDB? Or do we expect to use an entire separate Postgres database for each tenant?
  * "Set [...] `addschemaprefix` to add a `folio_` prefix to the schema names" -- what does this get us?
  * "In the Debezium PostgreSQL connector configuration, the following exclusions are suggested [list]". It would be interested to know the reasons for these exclusions. I am guessing most of them are omitted just because they are not of interest (e.g. pubsub state) but that `mod_login` is omitted for security reasons?

* In section 4.2.2 (Configuring Metadb for ReShare), "Before defining a ReShare data source, create a data origin for each consortial tenant". Does this mean each tenant _in_ a consortium, or each tenant _that represents_ a consortium? More generally, do these instructions pertain to using MetaDB for a ReShare tenant or for a ReShare consortium?

* Why no `trimschemaprefix` when ingesting from ReShare?

* At the end of section 4.2.2, some backquote slippage results in a hunk of text being in code font.

* Section 4.3 (MARC transform for LDP) makes it clear that MARC transformation for LDP Classic is done by an external program. Is that true of MetaDB, too, or is the MARC transformation integrated?

