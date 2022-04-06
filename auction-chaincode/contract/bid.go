package contract

// FullBid stores revealed bid's data
type FullBid struct {
	Type   string `json:"objectType"`
	Price  int    `json:"price"`
	Org    string `json:"org"`
	Bidder string `json:"bidder"`
}

// BidHash stores private bid's data
type BidHash struct {
	Org  string `json:"org"`
	Hash string `json:"hash"`
}

const bidKeyType = "bid"
