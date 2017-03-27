WindTurbine
===

[![Travis](https://img.shields.io/travis/WindPT/WindTurbine.svg)](https://travis-ci.org/WindPT/WindTurbine)
[![Gitter](https://img.shields.io/gitter/room/WindPT/WindTurbine.svg)](https://gitter.im/WindPT/WindTurbine)

An experimental tracker server designed to work with [WindPT](https://github.com/WindPT/WindPT).

## Requirements

 * Go 1.5 or higher
 * MySQL (4.1+) or MariaDB

## Installation

Download the zip file provided in Release and unzip it.

## Usage

Make a copy of `config.sample.xml`, rename it to `config.xml` and modify it.

Then, run this application.

## Compiling manually

You can compile this project manually by yourself.

```bash
$ go get github.com/WindPT/WindTurbine
$ cd $GOPATH/src/github.com/WindPT/WindTurbine
$ godep restore
$ make
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
* Peer: `downloaded`, `downloaded_add`, `uploaded`, `uploaded_add`, `rotio`, `time`, `time_la`, `time_leeched`, `time_seeded`

Functions supported by this project:

* Trigonometrics: `sin` `cos` `tan` `sinh` `cosh` `tanh` `arcsin` `arccos` `arctan` `arcsinh` `arccosh` `arctanh` `hypot`
* Roots: `sqrt` `cbrt`
* Logarithms: `lb` `ln` `lg`
* Exponentials: `pow10` `pow`
* Others: `abs` `ceil` `floor` `mod` `max` `min` `remainder`

*Restricted to PHPWind, you should change types of all fields named `credit(n)` of `pw_user_data` table and `pw_windid_user_data` table in your databse from `int` to `double`.*

## TODO

* [x] Peer Exchanging
* [x] Logging for Data Transfer and History
* [x] Credit

## Donate us

[Donate us](https://7in0.me/#donate)

## License

GNU GENERAL PUBLIC LICENSE Version 2

More info see [LICENSE](LICENSE)
