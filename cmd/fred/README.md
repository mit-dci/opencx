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

  1. **Submit**
      * The submit stage should take roughly `(b*t)/n` time. `n` should be greater than 1.
      * During the submit stage, users will submit timelock-encrypted orders with time parameter `t`.
  2. **Commit**
      * The commit stage marks the end of the "Submit" stage.
      * During the commit stage, the exchange broadcasts a commitment to a set of encrypted orders.
      * These encrypted orders include an unsolved puzzle, ciphertext, intended auction, and a hash.
  3. **Respond**
      * Users who receive the commitment before `b*t` send a signature on the commitment to the exchange.
      * If the exchange receives unanimous signatures before `b*t`, the exchange broadcasts these signatures.
      * If not all users signed off on the commitment, the entire auction is marked as invalid and must start over.
      * Users should sign during this period if they're confident that the exchange could not have possibly solved a single one of the puzzles in the commitment.
      * A single malicious user can halt the exchange during this step.
  4. **Decrypt**
      * This stage starts once the exchange has solved a puzzle, and decrypted an order in the set it committed to.
      * This should happen after `b*t`.
      * The exchange signs this data.
      * Once the exchange has decrypted all orders, it broadcasts these decrypted orders.
      * Some of the ciphertexts, once the puzzle is solved, will decrypt to garbage data, or invalid orders.
      * Valid encrypted orders will have a puzzle that reveals a decryption key to a ciphertext when solved. 
      This ciphertext must decrypt to a message that includes a valid order, and a proof that this message, when encrypted by puzzle, results to the same puzzle that was committed to.
      In the case of the RSW96 puzzle, this would be the trapdoor, meaning either p or q, or both.
      The hash of the message must be the same as the hash included in the encrypted order.
      * All users verify these rules. 
      If a user suspects that any part of any order may have been manipulated by the exchange, they can solve the puzzle and release the correct information.
      The exchange's signature on the incorrect data and the user's signature on the correct data is a sufficient proof that the exchange did something wrong.
      If this proof is provided it can be broadcast, and either the entire auction can be considered invalid, or the data can be updated and signed again.
      * One possibility is that the exchange produces an R value in the commit stage. 
      Then the exchange won't be able to sign an update to the data without revealing their private key.
      In this case the auction would have to be restarted.
      * In any case of an invalid auction, users don't need to pay attention to anything the exchange does that depends on the invalid auction.
  5. **Match**
      * The exchange matches the orders, and signs the outcome of the matching.
      * If the matching is incorrect, then it's incorrect and you can prove it.
  6. **Execute**
      * The exchange facilitates the trades through whatever settlement it feels like.
      * Ideally this would be done through lightning atomic swaps, since proofs can be produced for an honest execution.
      * In the case of lightning atomic swaps, blame can't be assigned to the user or the exchange if either party refuses to execute the trade.
      Because of this, we can only be sure that the exchange and user were cooperative. 
      Being able to assign blame in this situation is a problem for all non-custodial exchanges.

## Matching algorithms for this protocol

Because we have this period where orders can be committed to being matched (if valid) and not front-run, we can come up with matching algorithms that we otherwise wouldn't be able to trust to be fair.

This gives us options of both stateful and stateless matching algorithms.

### Stateless matching algorithms

Since this is a stateless matching algorithm, we only know about orders in the current auction.
None of the orders have any time priority, so price/time priority can't be used.
We can, however, use pro-rata matching for matching the orders in the auction.
We could also use priority for valid orders with the lowest hash. 
If we did base priority off of the orders with the lowest hash, instead of including the direct hash of the message alongside the puzzle, we should include the double hash.
This way users would be able to submit the double hash of the message, while revealing nothing about either the message or the preimage of the second hash / hash of the message that is used for priority.
Users could do a proof of work on their message once the auction has started, and the exchange would be able to do that as well. 
This would only be an issue if there is a large domain for proof of work.
This will normally be the case, since the message will include a signature, and the signature will be variable, depending on R.
This doesn't mean anyone can front-run, since the orders are still hidden and the exchange will be committing to taking a position if it commits to its own orders before a puzzle could have possibly been solved.

Stateless matching algorithms can be considered to be time independent.

### Stateful matching algorithms

We don't *have to* be stuck with only stateless matching algorithms.
We can use a combination of known, stateful algorithms that have a time priority requirement, and a stateless matching algorithm to have a persistent orderbook.
Every order in an auction must have the same time priority, so we can do the following steps to be compatible with a persistent orderbook:

  1. Match in auction
      * Nothing is different here than in the **Match** section of the protocol.
      * Here, we must match according to a stateless matching algorithm.
  2. Drop unfilled onto orderbook
      * Here, we'll still be using a stateless matching algorithm, but only to settle "ties".
      * Unfilled orders get put on the order queue with the highest time priority so far.
      The sequence of auctions determines the time priority for orders on this orderbook.
      In the case that multiple orders can be filled, but those orders have the same time priority, a stateless algorithm is used.
  3. Match according to any matching algorithm
      * Now, since we can settle ties with a stateless algorithm, we can use a stateful matching algorithm with the persistent orderbook.
