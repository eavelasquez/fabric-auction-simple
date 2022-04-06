package contract

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"

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
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
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
		Orgs:         []string{clientOrgID},
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
	err = setAssetStateBasedEndorsement(ctx, auctionID, clientOrgID)
	if err != nil {
		return fmt.Errorf("Failed setting state based endorsement for new organization: %v", err)
	}

	return nil
}

// QueryAuction allows all members of the channel to read a public auction.
func (c *AuctionContract) QueryAuction(ctx contractapi.TransactionContextInterface, auctionID string) (*Auction, error) {
	// Get Auction from the ledger.
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
func (c *AuctionContract) Bid(ctx contractapi.TransactionContextInterface, auctionID string) (string, error) {
	// Get Bid from transient map.
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return "", fmt.Errorf("Error getting bid from transient map: %v", err)
	}

	bid, ok := transientMap["bid"]
	if !ok {
		return "", fmt.Errorf("Bid key not found in the transient map")
	}

	// Get the implicit collection name using the bidder's organization ID.
	collection, err := getCollectionName(ctx)
	if err != nil {
		return "", fmt.Errorf("Failed to get implicit collection name: %v", err)
	}

	// The bidder has to target their peer to store the bid.
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return "", fmt.Errorf("Cannot store bid on this peer, not a member of this org: %v", err)
	}

	// The transaction ID is used as a unique index for the bid.
	txID := ctx.GetStub().GetTxID()

	// Create a composite key using the transaction ID.
	bidKey, err := ctx.GetStub().CreateCompositeKey(bidKeyType, []string{auctionID, txID})
	if err != nil {
		return "", fmt.Errorf("Failed to create composite key: %v", err)
	}

	// Put the bid into the organization's implicit data collection.
	err = ctx.GetStub().PutPrivateData(collection, bidKey, bid)
	if err != nil {
		return "", fmt.Errorf("Failed to input price into collection: %v", err)
	}

	// Return the transaction ID so that the used can identity and query the bid later.
	return txID, nil
}

// QueryBid allows the submitter of the bid to query the bid from public state.
func (c *AuctionContract) QueryBid(ctx contractapi.TransactionContextInterface, auctionID string, txID string) (*FullBid, error) {
	// Verify that the bidder is a member of the bidder's organization.
	err := verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to verify that the bidder is a member of the bidder's organization': %v", err)
	}

	// Get ID of submitting client identity.
	clientID, err := c.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to get ID of the client identity: %v", err)
	}

	// Get the implicit collection name of bidder's org.
	collection, err := getCollectionName(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to get implicit collection name: %v", err)
	}

	// Create a composite key using the auction ID and transaction ID.
	bidKey, err := ctx.GetStub().CreateCompositeKey(bidKeyType, []string{auctionID, txID})
	if err != nil {
		return nil, fmt.Errorf("Failed to create composite key: %v", err)
	}

	// Get the bid from the bidder's org's private data collection.
	bytes, err := ctx.GetStub().GetPrivateData(collection, bidKey)
	if err != nil {
		return nil, fmt.Errorf("Failed to get bid from collection: %v", err)
	}
	if bytes == nil {
		return nil, fmt.Errorf("Bid key does not exist in the collection")
	}

	// Unmarshal the bid into a FullBid object.
	var bid *FullBid

	err = json.Unmarshal(bytes, &bid)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal bid: %v", err)
	}

	// Check that client querying the bid is the bid owner.
	if bid.Bidder != clientID {
		return nil, fmt.Errorf("Permission denied, client id %v is not the bid owner of the bid", clientID)
	}

	return bid, nil
}

// Submit is used by the bidder to add the hash of that bid stored in private data
// to the auction. Note that this functions alters the auction in private state,
// and needs to meet the auction endorsement policy. Transaction ID is used to
// identify the bid.
func (c *AuctionContract) SubmitBid(ctx contractapi.TransactionContextInterface, auctionID string, txID string) error {
	// Get the MSP ID of the bidder's org.
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("Failed to get client identity MSP ID: %v", err)
	}

	// Get the auction from public state.
	auction, err := c.QueryAuction(ctx, auctionID)
	if err != nil {
		return fmt.Errorf("Failed to get auction from public state: %v", err)
	}

	// The auction needs to be open for users to add their bid.
	Status := auction.Status
	if Status != "open" {
		return fmt.Errorf("Cannot join closed or ended auction")
	}

	// Get the implicit collection name of bidder's org.
	collection, err := getCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get implicit collection name: %v", err)
	}

	// Use the transaction ID passed as a parameter to create composite bid key.
	bidKey, err := ctx.GetStub().CreateCompositeKey(bidKeyType, []string{auctionID, txID})
	if err != nil {
		return fmt.Errorf("Failed to create composite key: %v", err)
	}

	// Get the hash of the bid stored in private data collection.
	bidHash, err := ctx.GetStub().GetPrivateDataHash(collection, bidKey)
	if err != nil {
		return fmt.Errorf("Failed to get bid hash from private data collection: %v", err)
	}
	if bidHash == nil {
		return fmt.Errorf("Bid Hash does not exist in private data collection: %v", bidKey)
	}

	// Store the hash along with the bidder's organization.
	NewBidHash := BidHash{
		Org:  clientOrgID,
		Hash: fmt.Sprintf("%x", bidKey),
	}

	// Add the bid hash to the auction bidders hash.
	bidders := make(map[string]BidHash)
	bidders = auction.PrivateBids
	bidders[bidKey] = NewBidHash
	auction.PrivateBids = bidders

	// Add the bindding organizarion to the list of participating organizations if it is not already.
	Orgs := auction.Orgs
	if !contains(Orgs, clientOrgID) {
		newOrgs := append(Orgs, clientOrgID)
		auction.Orgs = newOrgs

		err = addAssetStateBasedEndorsement(ctx, auctionID, clientOrgID)
		if err != nil {
			return fmt.Errorf("Failed setting state based endorsement for new organizations: %v", err)
		}
	}

	newAuction, _ := json.Marshal(auction)

	// Update the auction in private state.
	err = ctx.GetStub().PutState(auctionID, newAuction)
	if err != nil {
		return fmt.Errorf("Failed to update auction state: %v", err)
	}

	return nil
}

// RevealBid is used by the bidder to reveal the bid after the auction is closed.
func (c *AuctionContract) RevealBid(ctx contractapi.TransactionContextInterface, auctionID string, txID string) error {
	// Get Bid from transient map.
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("Error getting bid from transient map: %v", err)
	}

	transientBid, ok := transientMap["bid"]
	if !ok {
		return fmt.Errorf("Bid key not found in the transient map")
	}

	// Get the implicit collection name using the bidder's organization ID.
	collection, err := getCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get implicit collection name: %v", err)
	}

	// Use transaction ID to create composite bid key.
	bidKey, err := ctx.GetStub().CreateCompositeKey(bidKeyType, []string{auctionID, txID})
	if err != nil {
		return fmt.Errorf("Failed to create composite bid key: %v", err)
	}

	// Get Bid Hash of bid if private bid on the public ledger.
	bidHash, err := ctx.GetStub().GetPrivateDataHash(collection, bidKey)
	if err != nil {
		return fmt.Errorf("Failed to get private bid hash from collection: %v", err)
	}
	if bidHash == nil {
		return fmt.Errorf("Bid hash does not exist in private data collection: %v", bidKey)
	}

	// Get auction from public state
	auction, err := c.QueryAuction(ctx, auctionID)
	if err != nil {
		return fmt.Errorf("Failed to get auction from public state: %v", err)
	}

	// Complete a series of three checks before we add the bid to the auction.

	// Check 1: check that the auction is closed. We cannot reveal a
	// bid if the auction is not closed.
	Status := auction.Status
	if Status != "closed" {
		return fmt.Errorf("Cannot reveal bid for open or ended auction")
	}

	// Check 2: check that hash of revealed bid matches hash of private bid
	// on the public ledger. This checks that the bidder is telling the truth
	// about the value of their bid.
	hash := sha256.New()
	hash.Write(transientBid)
	calculatedBidHash := hash.Sum(nil)

	// Verify that the hash of the passed immutable properties matches the on-chain hash.
	if !bytes.Equal(calculatedBidHash, bidHash) {
		return fmt.Errorf("Hash %x for bid hash %s does not match hash in auction: %x", calculatedBidHash, transientBid, bidHash)
	}

	// Check 3: check hash of revealed bid matches hash of private bid that was
	// added earlier. This ensures that the bid hash not changed since it
	// was added to the auction.
	bidders := auction.PrivateBids
	privateBidHashString := bidders[bidKey].Hash

	onChainBidHashString := fmt.Sprintf("%x", bidKey)
	if privateBidHashString != onChainBidHashString {
		return fmt.Errorf("Hash %s for bid %s does not match hash in auction: %s, bidder must have changed bid", privateBidHashString, transientBid, onChainBidHashString)
	}

	// We can add the bid to the auction if all checks have passed.
	type transientBidInput struct {
		Price  int    `json:"price"`
		Org    string `json:"org"`
		Bidder string `json:"bidder"`
	}

	// Unmarhsal the bid into a transientBidInput struct.
	var bidInput transientBidInput

	err = json.Unmarshal(transientBid, &bidInput)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal bid: %v", err)
	}

	// Get ID of submitting client identity.
	clientID, err := c.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get submitting client identity: %v", err)
	}

	// Marshal transient parameters and ID and MSP ID into bid object.
	NewBid := FullBid{
		Type:   bidKeyType,
		Price:  bidInput.Price,
		Org:    bidInput.Org,
		Bidder: bidInput.Bidder,
	}

	// Check 4: make sure that the transaction is being submitted is the bidder.
	if bidInput.Bidder != clientID {
		return fmt.Errorf("Permission denied, client id %v is not the owner of the bid", clientID)
	}

	// Add the bid to the auction.
	revealedBids := make(map[string]FullBid)
	revealedBids = auction.RevealedBids
	revealedBids[bidKey] = NewBid
	auction.RevealedBids = revealedBids

	// Update the auction in state.
	newAuction, _ := json.Marshal(auction)

	// Put auction with bid added back into state.
	err = ctx.GetStub().PutState(auctionID, newAuction)
	if err != nil {
		return fmt.Errorf("Failed to update auction state: %v", err)
	}

	return nil
}

// CloseAuction can be used by the seller to close the auction. This prevents bids
// from being added to the auction, and allows the auction to be revealed.
func (c *AuctionContract) CloseAuction(ctx contractapi.TransactionContextInterface, auctionID string) error {
	// Get auction from public state.
	auction, err := c.QueryAuction(ctx, auctionID)
	if err != nil {
		return fmt.Errorf("Failed to get auction from public state: %v", err)
	}

	// The auction can only closed by the seller.

	// Get ID of submitting client identity.
	clientID, err := c.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get submitting client identity: %v", err)
	}

	Seller := auction.Seller
	if Seller != clientID {
		return fmt.Errorf("Auction can only be closed by seller: %v", err)
	}

	// Check if auction is already closed.
	Status := auction.Status
	if Status != "open" {
		return fmt.Errorf("Cannot close auction that is not open")
	}

	// Change status of auction to closed.
	auction.Status = string("closed")

	// Update the auction in state.
	closedAuction, _ := json.Marshal(auction)

	err = ctx.GetStub().PutState(auctionID, closedAuction)
	if err != nil {
		return fmt.Errorf("Failed to close auction: %v", err)
	}

	return nil
}

// EndAuction both changes the auction state to closed, and reveals the winning bid.
func (c *AuctionContract) EndAuction(ctx contractapi.TransactionContextInterface, auctionID string) error {
	// Get auction from public state.
	auction, err := c.QueryAuction(ctx, auctionID)
	if err != nil {
		return fmt.Errorf("Failed to get auction from public state: %v", err)
	}

	// Check that the auction is being ended by the seller.

	// Get ID of submitting client identity.
	clientID, err := c.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get submitting client identity: %v", err)
	}

	Seller := auction.Seller
	if Seller != clientID {
		return fmt.Errorf("Auction can only be ended by seller: %v", err)
	}

	// Check if auction is already closed.
	Status := auction.Status
	if Status != "closed" {
		return fmt.Errorf("Cannot end auction that is not closed")
	}

	// Get the list of revealed bids
	revealedBids := auction.RevealedBids

	// Check if there are any revealed bids in the auction.
	if len(auction.RevealedBids) == 0 {
		return fmt.Errorf("No bids have been revealed, cannot end auction %v", err)
	}

	// Determine the highest bid.
	for _, bid := range revealedBids {
		if bid.Price > auction.Price {
			auction.Price = bid.Price
			auction.Winner = bid.Bidder
		}
	}

	// Check if there is a winning bid that has yet to be revealed.
	err = checkForHigherBid(ctx, auction.Price, auction.RevealedBids, auction.PrivateBids)
	if err != nil {
		return fmt.Errorf("Failed to check for higher bid, cannot end auction: %v", err)
	}

	// Change status of auction to ended.
	auction.Status = string("ended")

	// Update the auction in state.
	endedAuction, _ := json.Marshal(auction)

	err = ctx.GetStub().PutState(auctionID, endedAuction)
	if err != nil {
		return fmt.Errorf("Failed to end auction: %v", err)
	}

	return nil
}
