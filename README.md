goamt
=====

An automatic media transcoder written in Go with an emphasis on ease of management and performance.

Usage
=====

The goamt tool is separated into logical commands which each operate on the provided database.

Creating a new goamt database
-----------------------------

A new database can be created using the create command. For example:

```sh
$ goamt create --database goamt.db
2021-02-19T21:06:08Z INFO Created new database | {"version":1}
2021-02-19T21:06:08Z INFO Closing database
```

Running an initial update
-------------------------

After creating an initial database, it should be populated with the contents of an untranscoded
media library which can be done using the update command.

```sh
$ goamt update --database goamt.db --path .
2021-02-19T21:06:10Z INFO Opened existing database | {"version":1}
2021-02-19T21:06:10Z DEBU Beginning transaction | {"number":1}
2021-02-19T21:06:10Z INFO Adding entry | {"discovered":1613768770,"hash":1733426259,"path":"movie.mkv"}
2021-02-19T21:06:10Z DEBU Committing transaction | {"number":1}
2021-02-19T21:06:10Z DEBU Beginning transaction | {"number":2}
2021-02-19T21:06:10Z INFO Adding entry | {"discovered":1613768770,"hash":1283239824,"path":"tv show - S01E01.mp4"}
2021-02-19T21:06:10Z DEBU Committing transaction | {"number":2}
2021-02-19T21:06:10Z INFO Closing database
```

We can see that two media files were discovered and added to the database, using the SQLite CLI we
can look at these entries.

```sh
$ sqlite3 goamt.db 'select * from library;'
id  path                   discovered  transcoded  hash
--  ---------------------  ----------  ----------  ----------
1   movie.mkv              1613768770              1733426259
2   tv show - S01E01.mp4   1613768770              1283239824
```

Running another update won't modify these entries unless the path/hash changes for a file. For
example, lets rename a file.

```sh
$ mv movie.mkv a_different_movie.mkv

$ goamt update --database goamt.db --path .
2021-02-19T21:09:40Z INFO Opened existing database | {"version":1}
2021-02-19T21:09:40Z DEBU Beginning transaction | {"number":1}
2021-02-19T21:09:40Z INFO Adding entry | {"discovered":1613768980,"hash":1283239824,"path":"tv show - S01E01.mp4"}
2021-02-19T21:09:40Z DEBU Committing transaction | {"number":1}
2021-02-19T21:09:40Z DEBU Beginning transaction | {"number":2}
2021-02-19T21:09:40Z INFO Adding entry | {"discovered":1613768980,"hash":1733426259,"path":"a_different_movie.mkv"}
2021-02-19T21:09:40Z DEBU Committing transaction | {"number":2}
2021-02-19T21:09:40Z INFO Closing database

$ sqlite3 goamt.db 'select * from library;'
id  path                   discovered  transcoded  hash
--  ---------------------  ----------  ----------  ----------
1   a_different_movie.mp4  1613768770              1733426259
2   tv show - S01E01.mp4   1613768770              1283239824
```

Looking at the contents of the database, we can see that the rename has been correctly picked up and
we can continue using/updating the database as much as we need.

Generally updates should be performed after changing a media library for example:
1) When adding new media
2) When renaming media (i.e. with a tool such as [yamr](https://github.com/jamesl33/yamr))

Transcoding entries from the database
-------------------------------------

Transcoding can be performed using the transcode command. By default goamt will transcode n vCPU
entries using n vCPU threads, these options can be configured with the --entries/--threads flags.

```sh
$ goamt transcode --database goamt.db --path .
2021-02-19T21:17:06Z INFO Opened existing database | {"version":1}
2021-02-19T21:17:06Z DEBU Beginning transaction | {"number":1}
2021-02-19T21:17:06Z INFO Scheduling job to transcode entry | {"hash":1733426259,"id":1,"path":"a_different_movie.mkv"}
2021-02-19T21:17:06Z DEBU Added job for entry | {"hash":1733426259,"id":1,"path":"a_different_movie.mkv"}
2021-02-19T21:17:06Z DEBU Committing transaction | {"number":1}
2021-02-19T21:17:06Z DEBU Beginning transaction | {"number":2}
2021-02-19T21:17:06Z INFO Scheduling job to transcode entry | {"hash":1283239824,"id":2,"path":"tv show - S01E01.mp4"}
2021-02-19T21:17:06Z DEBU Added job for entry | {"hash":1283239824,"id":2,"path":"tv show - S01E01.mp4"}
2021-02-19T21:17:06Z DEBU Committing transaction | {"number":2}
2021-02-19T21:17:06Z DEBU Beginning transaction | {"number":3}
2021-02-19T21:17:06Z DEBU Rolling back transaction | {"number":3}
2021-02-19T21:17:06Z INFO Beginning job to transcode entry | {"hash":1733426259,"id":1,"path":"a_different_movie.mkv"}
2021-02-19T21:17:06Z INFO Beginning job to transcode entry | {"hash":1283239824,"id":2,"path":"tv show - S01E01.mp4"}
2021-02-19T21:17:06Z DEBU Beginning transaction | {"number":4}
2021-02-19T21:17:06Z INFO Completing job to transcode entry | {"hash":1733426259,"id":1,"path":"a_different_movie.mp4"}
2021-02-19T21:17:06Z DEBU Removing job for entry | {"hash":1733426259,"id":1,"path":"a_different_movie.mp4"}
2021-02-19T21:17:06Z DEBU Committing transaction | {"number":4}
2021-02-19T21:17:06Z DEBU Beginning transaction | {"number":5}
2021-02-19T21:17:06Z INFO Completing job to transcode entry | {"hash":1283239824,"id":2,"path":"tv show - S01E01.mp4"}
2021-02-19T21:17:06Z DEBU Removing job for entry | {"hash":1283239824,"id":2,"path":"tv show - S01E01.mp4"}
2021-02-19T21:17:06Z DEBU Committing transaction | {"number":5}
2021-02-19T21:17:06Z INFO Closing database
```

Looking at the logging you should be able to see the process taken by goamt when transcoding one or
more files. Note that these log statements may be interlaced since both files were being transcoded
concurrently.

Interrogating the database once again shows that a transcoded timestamp has been updated/populated;
these media file will not be re-transcoded by goamt. The entry will remain to allow rename detection
to function as expected.

```sh
$ sqlite3 goamt.db 'select * from library;'
id  path                   discovered  transcoded  hash
--  ---------------------  ----------  ----------  ----------
1   a_different_movie.mp4  1613768770  1613769426  1733426259
2   tv show - S01E01.mp4   1613768770  1613769426  1283239824
```

Logging
-------

By default goamt logs at the debug level which some may consider verbose. This is because goamt is
intended to be scripted/scheduled, therefore, logging at the debug level ensures logging is
correctly captured/recorded allowing retrospective debugging.

The log level may be configured using the `GOAMT_LOG_LEVEL` environment variable. Valid options are
`debug`, `info`, `warn`, `error` and `fatal`.

Concepts
========

[pytranscoder](https://github.com/jamesl33/pytranscoder) has served as a useful tool over the years,
however, it started to become more of a burden that it needed to be. goamt is intended to build upon
pytranscoder whilst solving a few fundamentals issues.

Ease of Management
------------------

Ideally, an automatic transcoding tool will:
1) Correctly detect and handle file renaming without re-transcoding
2) Safely handle/cleanup orhpaned transcode tasks without losing data
3) Should be stable enough to be setup (using cron/systemd) then forgotten about

The first issue here being that media liberties are organic, they evolve and change over time, grow
and shrink as tastes chance. The leads to the issue of being able to correctly detect renamed files;
if a media file has already been transcoded why waste time/electricity re-transcoding it.

goamt handles file renaming seamlessly by storing additional information about files other than the
filename (such as a hash). This means goamt is able update its database without "forgetting" that a
file has already been transcoded.

Secondly, an automatic transcoding tool should be able to handle/recover from partial transcoding
tasks which may have been interrupted by power cuts or other such unforeseen issues.

The goamt tool leverages an SQLite database, job tracking and ordered multi-stage modifications to
always allow a task to be correctly rolled back or completed (depending on how far though the
process the failed task was).

Finally, an automatic transcoding tool should be setup and then forgotten about it; it should do its
thing and be stable/resilient enough to proceed with mass transcoding without human iteration. goamt
builds upon point #2 with extensive unit testing to provide a stable tool which is intended to serve
this exact purpose.

Performance
-----------

Performance is of importance when automatically transcoding; a media library should be managed with
as little overhead as possible.

As mentioned the ease of management section, goamt hashes media files; it does this using a
non-cryptographic seeking hash to hash files quickly whilst providing hashes which shouldn't
collide when used with a normal sized media library.

Performing a full update of a cold cached 2TB media library stored on spinning disks takes ~5
minutes at 1% CPU. Re-running when cached takes ~40 seconds at 4% CPU.

Building
========

goamt is built using Go modules so can be built using `go build`.

```sh
$ go build

$ ./goamt --help
An automatic media transcoder written in Go with an emphasis on ease of management and performance

Usage:
   [command]

Available Commands:
  convert     Convert from the pytranscoder yaml format into the goamt SQLite format
  create      Create a new goamt SQLite database
  help        Help about any command
  transcode   Concurrently transcode a number of files
  update      Update a goamt SQLite database
  version     Display version information

Flags:
  -h, --help   help for this command

Use " [command] --help" for more information about a command.
```

Testing
=======

goamt is extensively unit tested all of which can be run using `go test`.

```sh
$ go test ./... -count=1 -cover
?   	github.com/jamesl33/goamt	[no test files]
ok  	github.com/jamesl33/goamt/cmd	0.015s	coverage: 75.4% of statements
ok  	github.com/jamesl33/goamt/database	0.034s	coverage: 80.4% of statements
ok  	github.com/jamesl33/goamt/utils	0.003s	coverage: 50.0% of statements
ok  	github.com/jamesl33/goamt/utils/sqlite	0.007s	coverage: 83.9% of statements
?   	github.com/jamesl33/goamt/value	[no test files]
ok  	github.com/jamesl33/goamt/version	0.002s	coverage: 100.0% of statements
```

License
=======
Copyright 2020 James Lee <jamesl33info@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
