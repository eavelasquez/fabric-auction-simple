'use strict';

const path = require('path');
const { Gateway } = require('fabric-network');

const {
	buildWallet,
  buildCCPOrg,
  prettyJSONString,
} = require('./utils/AppUtil');

const orgMSP1 = 'Org1MSP';
const orgMSP2 = 'Org2MSP';
const myChannel = 'mychannel';
const myChaincodeName = 'auction-chaincode';

/**
 * @description Submits the create bids transaction to the ledger and evaluates the result.
 * @param {*} ccp - The common connection profile.
 * @param {Wallet} wallet - The wallet.
 * @param {string} user - The user.
 * @param {string} orgMSP - The org MSP.
 * @param {string} auctionID - The auction ID.
 * @param {number} price - The price.
 */
async function createBid(ccp, wallet, user, orgMSP, auctionID, price) {
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

		// Evaluate the submitting client identity.
		console.log('\n--> Evaluate Transaction: Get your client ID');
		let bidder = await contract.evaluateTransaction(
			'GetSubmittingClientIdentity'
		);
		console.log('*** Result: Bidder ID is ' + bidder.toString());

		// Bid Data Structure.
		let bidData = {
			objectType: 'bid',
			price: parseInt(price),
			org: orgMSP,
			bidder: bidder.toString(),
		};

		// Submit the transaction.
		let statefulTxt = contract.createTransaction('CreateBid');

		statefulTxt.setEndorsingOrganizations(orgMSP); // Set the endorsing orgs.
		let transientMapData = Buffer.from(JSON.stringify(bidData)); // Convert the bid data to a buffer.
		statefulTxt.setTransient({ bid: transientMapData }); // Set the transient data.

		// Get the transaction ID.
		let bidID = statefulTxt.getTransactionID();

		console.log(
			'\n-> Submit Transaction: Create the bid that is stored in your organization\'s private data collection'
		);
		await statefulTxt.submit(auctionID);
		console.log('\n*** Result: committed');

		// Print the transaction ID.
		console.log('*** Result ***SAVE THIS VALUE*** BidID: ' + bidID.toString());

		// Evaluate the transaction.
		console.log(
			'\n--> Evaluate Transaction: Query the bid that was just created'
		);
		let result = await contract.evaluateTransaction(
			'QueryBid',
			auctionID,
			bidID
		);
		console.log('\n*** Result: Bid: ', prettyJSONString(result.toString()));

		// Disconnect from the gateway.
		await gateway.disconnect();
	} catch (error) {
		console.error(`Failed to submit bid transaction: ${error}`);
		if (error.stack) {
			console.error(error.stack);
		}
		process.exit(1);
	}
}

/**
 * @description Checks if the argument is valid.
 * @param {boolean} condition - The condition to check.
 * @param {string} message - The message to display if the condition is false.
 */
function checkArgs(condition, message = '') {
	if (!condition) {
		console.log(
			'\nUsage: node createBid.js.js <org> <userID> <auctionID> <price>'
		);
		console.log(message);
		process.exit(-1);
	}
}

/**
 * @description Creates an bid and submits it to the ledger.
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
			'Missing required arguments: org, userID, auctionID, price'
		);

		// Get all the arguments.
		let [, , org, user, auctionID, price] = process.argv;
		checkArgs(
			/^(org1|Org1|org2|Org2)$/.test(org),
			'Org must be either org1 or Org1 or org2 or Org2'
		);
		checkArgs(
			/^[a-zA-Z0-9]+$/.test(user),
			'User ID must be a non-empty string'
		);
		checkArgs(
			/^[0-9]+$/.test(auctionID),
			'Auction ID must be a non-empty string and must be a number'
		);
		checkArgs(
			/^[0-9]+$/.test(price),
			'Price must be a non-empty string and must be a number'
		);

		org = org.toLowerCase();

		const ccp = buildCCPOrg(org);
		const walletPath = path.join(__dirname, `wallet/${org}`);
		const wallet = await buildWallet(walletPath);

		await createBid(
			ccp,
			wallet,
			user,
			org === 'org1' ? orgMSP1 : orgMSP2,
			auctionID,
			price
		);
	} catch (error) {
		console.error(`Failed to run the create auction: ${error}`);
		process.exit(1);
	}
}

// Execute the main function.
main();
