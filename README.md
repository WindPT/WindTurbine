WindTurbine
======

[![Build Status](https://travis-ci.org/kinosang/WindTurbine.svg)](https://travis-ci.org/kinosang/WindTurbine)

An experimental tracker server designed to work with [WindPT](https://github.com/kinosang/WindPT).

This project is designed as a replacement to the tracker implemented in WindPT.

## Requirements

 * Go 1.3 or higher
 * MySQL (4.1+) or MariaDB

## Installation

Simple install the package to your $GOPATH with the go tool from shell:

```bash
$ go get github.com/kinosang/WindTurbine
```

And install dependency with doing a godep restore.

## Usage

Make a copy of `config.sample.xml`, rename it to `config.xml` and modify it.

Then, run this application.

```bash
$ make run
```

OR

```bash
$ make
$ ./WindTurbine
```

## Expression

This project use [Knetic/govaluate](https://github.com/Knetic/govaluate) (Arbitrary expression evaluation for golang) to support for credit expressions.

Operators and types supported by `govaluate`:

* Modifiers: `+` `-` `/` `*` `&` `|` `^` `**` `%` `>>` `<<`
* Comparators: `>` `>=` `<` `<=` `==` `!=` `=~` `!~`
* Logical ops: `||` `&&`
* Numeric constants, as 64-bit floating point (`12345.678`)
* String constants (single quotes: `'foobar'`)
* Date constants (single quotes, using any permutation of RFC3339, ISO8601, ruby date, or unix date; date parsing is automatically tried with any string constant)
* Boolean constants: `true` `false`
* Parenthesis to control order of evaluation `(` `)`
* Arrays (anything separated by `,` within parenthesis: `(1, 2, 'foo')`)
* Prefixes: `!` `-` `~`
* Ternary conditional: `?` `:`
* Null coalescence: `??`

Parameters supported by this project:

* Constants: `e`, `pi`, `phi`
* Torrent: `alive`, `seeders`, `leechers`, `size`
* User: `seeding`, `leeching`, `torrents`, `credit`
* Peer: `downloaded`, `downloaded_add`, `uploaded`, `uploaded_add`, `rotio`, `time`, `time_la`

Functions supported by this project:

* Trigonometrics: `sin` `cos` `tan` `sinh` `cosh` `tanh` `arcsin` `arccos` `arctan` `arcsinh` `arccosh` `arctanh` `hypot`
* Roots: `sqrt` `cbrt`
* Logarithms: `lb` `ln` `lg`
* Exponentials: `pow10` `pow`
* Others: `abs` `ceil` `floor` `mod` `max` `min` `remainder`

*Restricted to PHPWind, results will be convert into integer before saving.*

## TODO

* [x] Peer Exchanging
* [x] Logging for Data Transfer and History
* [x] Credit

## Donate us

[Donate us](https://7in0.me/#donate)

## License

GNU GENERAL PUBLIC LICENSE Version 2

More info see [LICENSE](LICENSE)
