OED Indexer
===========

This is just a tiny program, written in Golang, to generate index from words to their definition pages, in *the Oxford English Dictionary Second Edition* (and its additions), from [oed.com](https://oed.com).
Just access `https://www.oed.com/oed2/value` and replace `value` to the actual index (availability not ensured, works till the date of most recent commit) if you want the detailed text.

The output is [CSV](https://en.wikipedia.org/wiki/Comma-separated_values)-formatted, and has two columns, `Word` and `Index`.
There will be no headers if output is written to stdout (default).
**Notice #1:** this program will strip the *part of speech* label and only the word (or a list of word forms) is saved to the file, so multiple indices are pointing to a single word which may be of different *parts of speech*.
**Notice #2:** the table is not ensure to be sorted.

Please do not abuse this program. This program is for educational purposes only.

2023 Upgrage
------------

1. Build requirements boosted to Go 1.20.
2. Added custom user-agent string that deals with 403 error.
3. Better progress output.

Download & Install
------------------

You can simply use the standard `go get` method:

```
go get github.com/GreenYun/oed_indexer
```

By default, this method installs the program to `$GOPATH/bin/oed_indexer`.

Alternatively, you can clone the source and build by yourself:

```
git clone https://github.com/GreenYun/oed_indexer
cd oed_indexer
go build
```

Usage
-----

```
oed_indexer [-o FILE] [-c N] [-t N] [-v] [-p]

Options:
  -o FILE  Write to FILE insted of stdout
  -c N     Parse until N pages (default 291601)
  -t N     Start N simultaneous tasks (default nporc)
                [N=0 means default]
  -p       Show the percentage finished
  -pp      Show percentage, time elapsed and ETA
  -v       Verbose output
```

Licence
-------

[![MIT licence](http://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/GreenYun/oed_indexer/blob/master/LICENCE)

Use of this source code is governed by a MIT license that can be found in the [LICENCE](LICENCE) file.
