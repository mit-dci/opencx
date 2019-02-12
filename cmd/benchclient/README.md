# BenchClient

BenchClient is a go API for use in benchmarking. All of its methods call RPC Commands from a running server. You could use it as a generic golang client API as well, if you'd like to build your own client. It supports the same methods that `ocx` does, and shares a lot of the same code, but it doesn't take argument lists and instead takes in actual arguments.