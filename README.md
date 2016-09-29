WindTurbine
======

An experimental tracker server designed to work with [WindPT](https://github.com/labs7in0/WindPT).

This project is designed as an alternative to the tracker implemented in WindPT.

** This project does not update users' credits (coming soon). **

## Requirements

 * Go 1.3 or higher
 * MySQL (4.1+) or MariaDB

## Installation

Simple install the package to your $GOPATH with the go tool from shell:

```bash
$ go get github.com/labs7in0/WindTurbine
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

## TODO

* [x] Peer Exchanging
* [x] Logging for Data Transfer and History
* [ ] Credit

## Donate us

[Donate us](https://7in0.me/#donate)

## License

GNU GENERAL PUBLIC LICENSE Version 2

More info see [LICENSE](LICENSE)
