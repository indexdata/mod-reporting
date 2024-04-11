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

* "The data contained in the Metadb database originally come from another place: a *data source*". I assume multiple sources can contribute to a single MetaDB store, but this is worth checking.


## Appendix C. Changes to make to user documentation

Again, these observations arise from my reading of the user documentation, coming to it as a newbie.

* Leaping straight from "1.1. Getting started" to "1.2. Main tables" is bewildering. We really need more discussion of the concepts in between these sections. For example, that records have start and end datetimes which bound the period of their relevance in historic data.

* The "Main tables" and "Current tables" sections should be subsections of a higher-level section that describes the concepts in a high-level way.

* In current tables, I assume that `__id` and `__origin` columns exist (as in main tables) but probably not `__start`, `__end` or `__current`?

