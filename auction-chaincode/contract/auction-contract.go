package contract

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// AuctionContract is the contract that is used to manage auctions.
type AuctionContract struct {
	contractapi.Contract
}

// CreateAuction creates on auction on the public channel. The identity that
// submits the transaction becomes the seller of the auction.
func (c *AuctionContract) CreateAuction(ctx contractapi.TransactionContextInterface, auctionID string, itemsold string) error {
	// Get ID of submitting client identity.
	clientID, err := c.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get ID of the client identity: %v", err)
	}

	// Get Org of submitting client identity.
	clientOrg, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("Failed to get Org client identity: %v", err)
	}

	// Create auction object.
	bidders := make(map[string]BidHash)
	revealedBids := make(map[string]FullBid)

	auction := Auction{
		Type:         "auction",
		ItemSold:     itemsold,
		Price:        0,
		Seller:       clientID,
		Orgs:         []string{clientOrg},
		PrivateBids:  bidders,
		RevealedBids: revealedBids,
		Winner:       "",
		Status:       "open",
	}

	bytes, err := json.Marshal(auction)
	if err != nil {
		return fmt.Errorf("Failed to marshal auction object: %v", err)
	}

	// Store auction object into state.
	err = ctx.GetStub().PutState(auctionID, bytes)
	if err != nil {
		return fmt.Errorf("Failed to put auction object: %v", err)
	}

	// Set the seller of the auction as an endorser.
	err = setAssetStateBasedEndorsement(ctx, auctionID, clientOrg)
	if err != nil {
		return fmt.Errorf("Failed setting state based endorsement for new organization: %v", err)
	}

	return nil
}

// QueryAuction allows all members of the channel to read a public auction.
func (c *AuctionContract) QueryAuction(ctx contractapi.TransactionContextInterface, auctionID string) (*Auction, error) {
	bytes, err := ctx.GetStub().GetState(auctionID)
	if err != nil {
		return nil, fmt.Errorf("Failed to get auction object %v: %v", auctionID, err)
	}
	if bytes == nil {
		return nil, fmt.Errorf("Auction %v does not exist", auctionID)
	}

	auction := new(Auction)

	err = json.Unmarshal(bytes, auction)

	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal auction object %v: %v", auctionID, err)
	}

	return auction, nil
}

// Bid is used to add a user's bid to an auction. The bid is stored in the private
// data collection on the peer of the bidder's organization. The function returns
// the transaction ID so that users can identify and query the bid later.
func (c *AuctionContract) Bid(ctx contractapi.TransactionContextInterface, auctionID string) error {
	return nil
}

// QueryBid allows the submitter of the bid to query the bid from public state.
func (c *AuctionContract) QueryBid(ctx contractapi.TransactionContextInterface, auctionID string, txID string) (*FullBid, error) {
	var bid *FullBid

	return bid, nil
}

// Submit is used by the bidder to add the hash of that bid stored in private data
// to the auction. Note that this functions alters the auction in private state,
// and needs to meet the auction endorsement policy. Transaction ID is used to
// identify the bid.
func (c *AuctionContract) SubmitBid(ctx contractapi.TransactionContextInterface, auctionID string, txID string) error {
	return nil
}

// RevealBid is used by the bidder to reveal the bid after the auction is closed.
func (c *AuctionContract) RevealBid(ctx contractapi.TransactionContextInterface, auctionID string) error {
	return nil
}

// CheckForHigherBid is an internal function that is used to determine if a
// winning bid has yet to be revealed.
func (c *AuctionContract) CheckForHigherBid(ctx contractapi.TransactionContextInterface, auctionPrice int, revealedBidders map[string]FullBid, bidders map[string]BidHash) error {
	// Get MSP ID of peer org.
	peerMSPID, err := shim.GetMSPID()
	if err != nil {
		return fmt.Errorf("Failed to get MSP ID of peer org: %v", err)
	}

	var error error = nil

	// Loop through all bidders and check if they are the highest bidder.
	for bidKey, privateBid := range bidders {
		_, bidInAuction := revealedBidders[bidKey]

		// Bid is not already revealed, so check if it is the highest bidder, otherwise skip.
		if !bidInAuction {
			collection := "_implicit_org_" + privateBid.Org

			// If private bid is from the same org as the peer, then check if it is the highest bidder.
			if privateBid.Org == peerMSPID {
				// Get bid from private data collection.
				bytes, err := ctx.GetStub().GetPrivateData(collection, bidKey)
				if err != nil {
					return fmt.Errorf("Failed to get private data of bid from collection %v: %v", bidKey, err)
				}
				if bytes == nil {
					return fmt.Errorf("Bid %v does not exist", bidKey)
				}

				bid := new(FullBid)

				err = json.Unmarshal(bytes, bid)

				if err != nil {
					return fmt.Errorf("Failed to unmarshal bid %v: %v", bidKey, err)
				}

				// Check if bid is higher than auction price.
				if bid.Price > auctionPrice {
					error = fmt.Errorf("Cannot close auction, bidder has a higher price: %v", err)
				}
			} else {
				// Get bid hash from from private data collection.
				Hash, err := ctx.GetStub().GetPrivateDataHash(collection, bidKey)
				if err != nil {
					return fmt.Errorf("Failed to get private data of bid hash from collection %v: %v", bidKey, err)
				}
				if Hash == nil {
					return fmt.Errorf("Bid hash %v does not exist", bidKey)
				}
			}
		}
	}

	return error
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
