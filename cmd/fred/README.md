# fred

**fred** stands for Front-running Resistant Exchange Daemon.
It uses a timelock puzzle based protocol and batch matching to protect the exchange from front-running orders.
**fred** is the second daemon implemented, and should serve as a good reference for how OpenCX should be used.

## The FRED protocol

The FRED protocol is the protocol that the front-running resistant exchange daemon follows.
The main idea is that the exchange must commit to timelock-puzzle encrypted orders, in order to prevent front-running.

There are a couple stages to the exchange protocol, and some public parameters.
In this protocol, trading happens in rounds that we're going to be calling **auctions**.

## Global public parameters: 
 
`t`: Parameter for timelock puzzles, should be an integer. 
Based on the hardware in existence, it represents the minimum amount of time to solve a timelock puzzle.

`n`: How many "Auction periods" should fit in to the time it should take to solve a timelock puzzle with time parameter `t`.

`b`: The amount of time it takes to solve a puzzle with time parameter `t=1`.

## Auction protocol
Each auction has an Auction ID parameter.

  1. Submit
    * The submit stage should take roughly (b\*t)/n time. n should be greater than 1.
  2. Commit
  3. Respond
  4. Decrypt
  5. Match
  6. Execute