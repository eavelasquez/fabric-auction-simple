"use strict";

const path = require("path");
const { Gateway, Wallet } = require("fabric-network");

const {
  prettyJSONString,
  buildCCPOrg,
  buildWallet,
} = require("./utils/AppUtil");

const myChannel = "mychannel";
const myChaincodeName = "auction-chaincode";

/**
 * @description Submits the query auction transaction to the ledger and evaluates the result.
 * @param {*} ccp - The common connection profile.
 * @param {Wallet} wallet - The wallet.
 * @param {string} user - The user.
 * @param {string} auctionID - The auction ID.
 */
async function queryAuction(ccp, wallet, user, auctionID) {
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

    // Evaluate the transaction.
    console.log("\n--> Evaluate Transaction: Query Auction");
    let result = await contract.evaluateTransaction("QueryAuction", auctionID);
    console.log("\n*** Result: Auction: ", prettyJSONString(result.toString()));

    // Disconnect from the gateway.
    await gateway.disconnect();
  } catch (error) {
    console.error(`Failed to submit query auction transaction: ${error}`);
    process.exit(1);
  }
}

/**
 * @description Checks if the argument is valid.
 * @param {boolean} condition - The condition to check.
 * @param {string} message - The message to display if the condition is false.
 */
function checkArgs(condition, message = "") {
  if (!condition) {
    console.log("\nUsage: node queryAuction.js <org> <userID> <auctionID>");
    console.log(message);
    process.exit(-1);
  }
}

/**
 * @description Query an auction and submits it to the ledger.
 */
async function main() {
  try {
    // Check if the user has provided all the required inputs.
    checkArgs(
      process.argv.length < 4 ||
        process.argv[2] === undefined ||
        process.argv[3] === undefined ||
        process.argv[4] === undefined,
      "Missing required arguments: org, userID, auctionID"
    );

    // Get all the arguments.
    const [, , org, user, auctionID] = process.argv;
    checkArgs(
      /^(org1|Org1|org2|Org2)$/.test(org),
      "Org must be either org1 or Org1 or org2 or Org2"
    );
    checkArgs(
      /^[a-zA-Z0-9]+$/.test(user),
      "User ID must be a non-empty string"
    );
    checkArgs(
      /^[0-9]+$/.test(auctionID),
      "Auction ID must be a non-empty string and must be a number"
    );

    org = org.toLowerCase();

    const ccp = buildCCPOrg(org);
    const walletPath = path.join(__dirname, `wallet/${org}`);
    const wallet = await buildWallet(walletPath);

    await queryAuction(ccp, wallet, user, auctionID);
  } catch (error) {
    console.error(`Failed to run the query auction: ${error}`);
    process.exit(1);
  }
}

// Execute the main function.
main();
