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
 * @description Submits the create auction transaction to the ledger and evaluates the result.
 * @param {*} ccp - The common connection profile.
 * @param {Wallet} wallet - The wallet.
 * @param {string} user - The user.
 * @param {string} auctionID - The auction ID.
 * @param {string} item - The item.
 * @returns {Promise<void>}
 */
async function createAuction(ccp, wallet, user, auctionID, item) {
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

    // Submit the transaction.
    let statefulTxt = contract.createTransaction('CreateAuction');

    console.log('\n-> Submit Transaction: Propose a new auction');
    await statefulTxt.submit(auctionID, item);
    console.log('\n*** Result: committed');

    // Evaluate the transaction.
    console.log(
      '\n--> Evaluate Transaction: Query the auction that was just created'
    );
    let result = await contract.evaluateTransaction('QueryAuction', auctionID);
    console.log('\n*** Result: Auction: ', prettyJSONString(result.toString()));

    // Disconnect from the gateway.
    await gateway.disconnect();
  } catch (error) {
    console.error(`Failed to submit bid transaction: ${error}`);
  }
}

// Argument list for the script.
const fileAndArgs = 'createAuction.js <org> <userID> <auctionID> <item>';

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
      'Missing required arguments: org, userID, auctionID, item'
    );

    // Get all the arguments.
    let [, , org, user, auctionID, item] = process.argv;
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
      /^[a-zA-Z0-9]+$/.test(item),
      fileAndArgs,
      'Item must be a non-empty string'
    );

    org = org.toLowerCase();

    const ccp = buildCCPOrg(org);
    const walletPath = path.join(__dirname, `wallet/${org}-wallet`);
    const wallet = await buildWallet(walletPath);

    await createAuction(ccp, wallet, user, auctionID, item);
  } catch (error) {
    handleError('Failed to run the create auction', error);
  }
}

// Execute the main function.
main();
