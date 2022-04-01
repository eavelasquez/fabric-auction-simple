package main

import (
	"log"

	. "auction-chaincode/contract"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-contract-api-go/metadata"
)

// main function starts up the chaincode
func main() {
	auctionContract := new(AuctionContract)
	auctionContract.Info.Version = "0.0.1"
	auctionContract.Info.Description = "Auction Simple Smart Contract"
	auctionContract.Info.Contact = new(metadata.ContactMetadata)
	auctionContract.Info.Contact.Name = "Esteban Velasquez"

	chaincode, err := contractapi.NewChaincode(auctionContract)
	chaincode.Info.Title = "auction-chaincode chaincode"
	chaincode.Info.Version = "0.0.1"

	if err != nil {
		log.Panicf("Error creating AuctionContract chaincode: %v", err)
	}

	err = chaincode.Start()

	if err != nil {
		log.Panicf("Error starting AuctionContract chaincode: %v", err)
	}
}
