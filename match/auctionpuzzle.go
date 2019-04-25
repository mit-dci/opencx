package match

// AuctionPuzzle holds the encrypted message and the puzzle. The message when decrypted with the correct
// puzzle answer should be serializable to an auction order.
type AuctionPuzzle struct {
	puzzle  []byte
	message []byte
}
