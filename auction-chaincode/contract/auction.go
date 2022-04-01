package contract

// Auction stores auction's data
type Auction struct {
	Type         string             `json:"objectType"`
	ItemSold     string             `json:"item"`
	Seller       string             `json:"seller"`
	Orgs         []string           `json:"organizations"`
	PrivateBids  map[string]BidHash `json:"privateBids"`
	RevealedBids map[string]FullBid `json:"revealedBids"`
	Winner       string             `json:"winner"`
	Price        int                `json:"price"`
	Status       string             `json:"status"`
}
