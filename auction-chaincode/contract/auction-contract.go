package contract

import "github.com/hyperledger/fabric-contract-api-go/contractapi"

type AuctionContract struct {
	contractapi.Contract
}

// CreateAuction creates on auction on the public channel. The identity that
// submits the transaction becomes the seller of the auction.
func (c *AuctionContract) CreateAuction(ctx contractapi.TransactionContextInterface, auctionID string, itemsold string) error {
	return nil
}

// QueryAuction allows all members of the channel to read a public auction.
func (c *AuctionContract) QueryAuction(ctx contractapi.TransactionContextInterface, auctionID string) (*Auction, error) {
	var auction *Auction

	return auction, nil
}

// Bid is used to add a user's bid to an auction. The bid is stored in the private
// data collection on the peer of the bidder's organization. The function returns
// the transaction ID so that users can identify and query the bid later.
func (c *AuctionContract) Bid(ctx contractapi.TransactionContextInterface, auctionID string) error {
	return nil
}

// QueryBid allows the submitter of the bid to query the bid from public state.
func (c *AuctionContract) QueryBid(ctx contractapi.TransactionContextInterface, auctionID string, txtID string) (*FullBid, error) {
	var bid *FullBid

	return bid, nil
}

// Submit is used by the bidder to add the hash of that bid stored in private data
// to the auction. Note that this functions alters the auction in private state,
// and needs to meet the auction endorsement policy. Transaction ID is used to
// identify the bid.
func (c *AuctionContract) SubmitBid(ctx contractapi.TransactionContextInterface, auctionID string, txtID string) error {
	return nil
}

// RevealBid is used by the bidder to reveal the bid after the auction is closed.
func (c *AuctionContract) RevealBid(ctx contractapi.TransactionContextInterface, auctionID string) error {
	return nil
}

// CheckForHighterBid if an internal function that is used to determine if a
// winning bid has yet to be revealed.
func (c *AuctionContract) CheckForHighterBid(ctx contractapi.TransactionContextInterface, auctionPrice int, revealedBidders map[string]FullBid, bidders map[string]BidHash) error {
	return nil
}

// CloseAuction can be used by the seller to close the auction. This prevents bids
// from being added to the auction, and allows the auction to be revealed.
func (c *AuctionContract) CloseAuction(ctx contractapi.TransactionContextInterface, auctionID string) error {
	return nil
}

// EndAuction both changes the auction state to closed, and reveals the winning bid.
func (c *AuctionContract) EndAuction(ctx contractapi.TransactionContextInterface, auctionID string) error {
	return nil
}
