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

* "In the current version of Metadb, only top-level, scalar JSON fields are
extracted into transformed tables." I need to find out what non-top-level and non-scalar fields are.

* Is there any particular reason why the names of special tables _end_ with `__`, as in `patrongroup__`, while the names of special fields _begin_ with `__`, as in `__start`?

* "For monetary amounts, `numeric(19, 4)` is usually a good choice.  For exchange rates, `numeric(19, 14)` may be used." Why?

* Section 1.8. on Creating reports should include information on publishing them to GitHub for use in the FOLIO Reporting app.

* Section 1.10 explains for a running MetaDB instance can provide information about itself, such as its version number, when the various tables were last updated, system logging messages and query status. It would be helpful to expose at least some of this in the Reporting app. I have filed [a UILDP issue in Jira](https://folio-org.atlassian.net/browse/UILDP-148) and [a mod-reporting issue in the GitHub tracker](https://github.com/indexdata/mod-reporting/issues/66) so we don't forget about this.

* Section 2.1. Data type conversion seems to be referring to on-the-fly adjustments to the definitions of existing tables based on ingesting new data. Am I interpreting that correctly? In the table in section 2.1.1, how does MetaDB decide which conversion to apply?

* In section 2.4, "Metadb allows scheduling external SQL files to run on a regular basis." How? Also, "any tables created should not specify a schema name." Why not? Are all these tables implicitly in an "external directives" schema?

* In section 2.4.1, "The --metadb:table directive declares that the SQL file updates a specific table.  This allows Metadb to report on the status of the table." How?

* In section 2.5, "These statements are only available when connecting to the Metadb server (not the database)." How does one do that?

* Section 2.5.2. ALTER TABLE: how is this different from regular `ALTER TABLE`? (Comments along the lines of "It differs from GRANT in thatthe authorization will also apply to tables created at a later time in the data source" in section 2.5.3 would be helpful here.)

* The section on "CREATE DATA ORIGIN" does not have a section number, and so appears as part of section 2.5.3 on AUTHORIZE. Also, since we don't know what a data origin is, and how it differs from a data source, this is not very informative.

* "When the initial snapshot has finished streaming, the message "source snapshot complete (deadline exceeded)" will be written to the log." What does the "deadline exceeded" part mean?

* Reading section 2.5.4. CREATE DATA SOURCE, it's quickly apparent that we need a narrative guide on setting up Kafka for a FOLIO or ReShare installation, then creating MetaDB data sources for it. This may exist further down the document -- if so it should be referenced here -- or may need to be written.

* In section 2.5.5, "CREATE USER defines a new database user that will be managed by Metadb." What does it mean for MetaDB to manage the user?

XXX note to self: read to the end of section 2.
