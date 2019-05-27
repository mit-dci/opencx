# crypto
[![Build Status](https://travis-ci.org/mit-dci/opencx.svg?branch=master)](https://travis-ci.org/mit-dci/opencx)
[![License](https://img.shields.io/badge/License-MIT-brightgreen.svg)](https://github.com/mit-dci/opencx/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/mit-dci/opencx)](https://goreportcard.com/report/github.com/mit-dci/opencx)
[![GoDoc](https://godoc.org/github.com/mit-dci/opencx/crypto?status.svg)](https://godoc.org/github.com/mit-dci/opencx/crypto)

The crypto package currently has an interface for Timelock Puzzles, and an implementation of both the RSW96 timelock puzzle and a simple hash-based timelock puzzle.
In the case of the hash-based timelock puzzle, it takes just as long to create the puzzle (if you are encrypting information with the result) as it does to solve it.
With RSW96, this is not the case, since RSW96 has a trapdoor.

General overview of the **crypto** package:
  - hashtimelock
    - Implements the `Timelock` and `Puzzle` interfaces from [timelock.go](./timelock.go)
    - Uses sequential hashing as the inherently sequential function
    - Hash function used in sequential hashing can be anything implementing `hash.Hash`
  - provisions
    - Should implement [Provisions: Privacy-preserving proofs of solvency for Bitcoin](https://eprint.iacr.org/2015/1008)
    - Work in progress.
  - rsw
    - Implements the `Timelock` and `Puzzle` interfaces from [timelock.go](./timelock.go)
    - Uses repeated squaring over the integers modulo N, where N is a group of unknown order.
    In this case N is, as defined in [RSW96](https://people.csail.mit.edu/rivest/pubs/RSW96.pdf), an RSA modulus
    - Implements the `VDF` interface from [vdf.go](./vdf.go)
    - Uses the [Wesolowski construction](https://eprint.iacr.org/2018/623) for VDF proofs.
  - timelockencoders
    - Uses the `rsw` and `hashtimelock` packages to expose an easy-to-use API for creating publishable timelock puzzles