# Fabric Auction Simple

## About the Fabric Blind Auction

The blind auction simple uses Hyperledger Fabric to run an auction where bids are kept private until the auction period is over. Instead of displaying the full bid on the public ledger, buyers can only see hashes of the bids while the bidding is underway. This prevents buyers from changing their bids in response to bids submitted by others. After the bidding period ends, participants reveal their bid to try to win the auction. The organizations participating in the auction verify that a revealed bid matches the hash on the public ledger. Whichever has the highest bid wins.

A user that wants to sell one item can use the smart contract to create an auction. The auction is stored on the channel ledger and can be read by all channel members. The auctions created by the smart contract are run in three steps:

1. Each auction is created with the status **open**. While the auction is open, buyers can add new bids to the auction. The full bids of each buyer are stored in the implicit private data collections of their organization. After the bid is created, the bidder can submit the hash of the bid to the auction. A bid is added to the auction in two steps because the transaction that creates the bid only needs to be endorsed by a peer of the bidder's organization, while a transaction that updates the auction may need to be endorsed by multiple organizations. When the bid is added to the auction, the bidder's organization is added to the list of organizations that need to endorse any updates to the auction.
1. The auction is **closed** to prevent additional bids from being added to the auction. After the auction is closed, bidders that submitted bids to the auction can reveal their full bid. Only revealed bids can win the auction.
1. The auction is **ended** to calculate the winner from the set of revealed bids. All organizations participating in the auction calculate the price that clears the auction and the winning bid. The seller can end the auction only if all bidding organizations endorse the same winner and price.

Before endorsing the transaction that ends the auction, each organization queries the implicit private data collection on their peers to check if any organization member has a winning bid that has not yet been revealed. If a winning bid is found, the organization will withhold its endorsement and prevent the auction from being closed. This prevents the seller from ending the auction prematurely or colluding with buyers to end the auction at an artificially low price.

The sample uses several Fabric features to make the auction private and secure. Bids are stored in private data collections to prevent bids from being distributed to other peers in the channel. When bidding is closed, the auction smart contract uses the `GetPrivateDataHash()` API to verify that the bid stored in private data is the same bid that is being revealed. State based endorsement is used to add the organization of each bidder to the auction endorsement policy. The smart contract uses the `GetClientIdentity.GetID()` API to ensure that only the potential buyer can read their bid from private state and only the seller can close or end the auction.

## To Deploy the Auction Chaincode

We'll run the auction smart contract using the Fabric test network. The auction smart contract is deployed to the **default** channel.

```bash
git clone https://github.com/hyperledger/fabric-samples.git
cd fabric-samples/test-network
./network.sh up createChannel -ca
```

Note that use the `-ca` flag to deploy the network using certificate authorities. We'll use the CA to register and enroll our sellers and buyers.

Run the following command to deploy the auction smart contract. We will override the default endorsement policy to allow any channel member to create an auction without requiring an endorsement from another organization.

```bash
./network.sh deployCC -ccn auction-chaincode -ccp path/to/auction-chaincode -ccl go -ccep "OR('Org1MSP.peer','Org2MSP.peer')"
```

## To Set Up the Auction Application

We'll interact with the auction smart contract through a set of Node.js applications.

```bash
cd fabric-auction-simple/auction-application
npm install # install dependencies
```

## To Clean Up

When you're done, you can clean up the network by running the following commands.

```bash
cd fabric-samples/test-network
./network.sh down
```
