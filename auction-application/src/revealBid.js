'use strict';

const path = require('path');
const { Gateway } = require('fabric-network');

const {
  buildCCPOrg,
  buildWallet,
  checkArgs,
  handleError,
  prettyJSONString,
} = require('./utils/AppUtil');

const myChannel = 'mychannel';
const myChaincodeName = 'auction-chaincode';

/**
 * @description Submits the reveal bid transaction to the ledger and evaluates the result.
 * @param {*} ccp - The common connection profile.
 * @param {Wallet} wallet - The wallet.
 * @param {string} user - The user.
 * @param {string} auctionID - The auction ID.
 * @param {string} bidID - The bid ID.
 * @returns {Promise<void>}
 */
async function addBid(ccp, wallet, user, auctionID, bidID) {
  try {
    // Create a new gateway for connecting to our peer node.
    const gateway = new Gateway();

    // Connect using Discovery enabled.
    await gateway.connect(ccp, {
      wallet,
      identity: user,
      discovery: { enabled: true, asLocalhost: true },
    });

    // Get the network (channel) our contract is deployed to.
    const network = await gateway.getNetwork(myChannel);
    const contract = network.getContract(myChaincodeName);

    // Query the bid. (This is a read-only transaction.)
    console.log('\n--> Evaluate Transaction: Query Bid');
    let bid = await contract.evaluateTransaction('QueryBid', auctionID, bidID);
    bid = JSON.parse(bid); // Convert the JSON string to an object.

    // Query the auction. (This is a read-only transaction.)
    console.log('\n--> Evaluate Transaction: Query Auction');
    let auction = await contract.evaluateTransaction('QueryAuction', auctionID);
    auction = JSON.parse(auction); // Convert the JSON string to an object.

    // Bid Data Structure.
    let bidData = {
      objectType: 'bid',
      price: parseInt(bid.price),
      org: bid.org,
      bidder: bid.bidder,
    };
    console.log('*** Result: Bid: ', JSON.stringify(bidData, null, 2));

    // Submit the transaction.
    let statefulTxt = contract.createTransaction('RevealBid');

    let transientMapData = Buffer.from(JSON.stringify(bidData)); // Convert the bid data to a buffer.
    statefulTxt.setTransient({ bid: transientMapData }); // Set the transient data.

    // Set the endorsing orgs.
    if (auction.organizations.length === 2) {
      statefulTxt.setEndorsingOrganizations(
        auction.organizations[0],
        auction.organizations[1]
      );
    } else {
      statefulTxt.setEndorsingOrganizations(auction.organizations[0]);
    }

    console.log('\n-> Submit Transaction: Reveal Bid');
    await statefulTxt.submit(auctionID, bidID);
    console.log('\n*** Result: committed');

    // Evaluate the transaction.
    console.log(
      '\n--> Evaluate Transaction: Query the auction to see that our bid was added'
    );
    let result = await contract.evaluateTransaction('QueryAuction', auctionID);
    console.log('\n*** Result: Auction: ', prettyJSONString(result.toString()));

    // Disconnect from the gateway.
    await gateway.disconnect();
  } catch (error) {
    console.error(`Failed to submit reveal bid transaction: ${error}`);
    process.exit(1);
  }
}

// Argument list for the script.
const fileAndArgs = 'revealBid.js <org> <userID> <auctionID> <bidID>';

/**
 * @description Creates an auction and submits it to the ledger.
 */
async function main() {
  try {
    // Check if the user has provided all the required inputs.
    checkArgs(
      process.argv.length < 5 ||
        process.argv[2] === undefined ||
        process.argv[3] === undefined ||
        process.argv[4] === undefined ||
        process.argv[5] === undefined,
      fileAndArgs,
      'Missing required arguments: org, userID, auctionID, bidID'
    );

    // Get all the arguments.
    let [, , org, user, auctionID, bidID] = process.argv;
    checkArgs(
      /^(org1|Org1|org2|Org2)$/.test(org),
      fileAndArgs,
      'Org must be either org1 or Org1 or org2 or Org2'
    );
    checkArgs(
      /^[a-zA-Z0-9]+$/.test(user),
      fileAndArgs,
      'User ID must be a non-empty string'
    );
    checkArgs(
      /^[0-9]+$/.test(auctionID),
      fileAndArgs,
      'Auction ID must be a non-empty string and must be a number'
    );
    checkArgs(
      /^[a-zA-Z0-9]+$/.test(bidID),
      fileAndArgs,
      'Bid ID must be a non-empty string'
    );

    org = org.toLowerCase();

    const ccp = buildCCPOrg(org);
    const walletPath = path.join(__dirname, `wallet/${org}`);
    const wallet = await buildWallet(walletPath);

    await addBid(ccp, wallet, user, auctionID, bidID);
  } catch (error) {
    handleError('Failed to run the reveal bid transaction: ', error);
  }
}

// Execute the main function.
main();
