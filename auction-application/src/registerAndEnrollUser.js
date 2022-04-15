'use strict';

const path = require('path');

const { buildCCPOrg, buildWallet } = require('./utils/AppUtil');
const { buildCAClient, registerAndEnrollUser } = require('./utils/CAUtil');

const mspOrg1 = 'Org1MSP';
const mspOrg2 = 'Org2MSP';

/**
 * @description Connects to the CA of the specified Org and, register and enrolls the user.
 * @param {string} userId - The ID of the user.
 * @param {string} orgName - The name of the organization e.g. org1 or org2.
 * @param {string} mspOrg - The MSP ID of the Org.
 */
async function connectToOrgCA(userId, orgName, mspOrg) {
	console.log(`\n--> Register and enrolling a new user: ${userId}`);
	const ccpOrg = buildCCPOrg(orgName);
	const caOrgClient = buildCAClient(ccpOrg, `ca.${orgName}.example.com`);

	const walletPathOrg = path.join(__dirname, `wallet/${orgName}`);
	const walletOrg = await buildWallet(walletPathOrg);

	await registerAndEnrollUser(
		caOrgClient,
		walletOrg,
		mspOrg,
		userId,
		`${orgName}.department1`
	);
}

/**
 * Register and enrolls the user of the Org1 CA or Org2 CA.
 */
async function main() {
	if (process.argv[2] === undefined && process.argv[3] === undefined) {
		console.log('Usage: node registerAndEnrollUser.js Org userID');
		process.exit(1);
	}

	let [, , org, userId] = process.argv;
	try {
		if (!/^(org1|Org1|org2|Org2)$/.test(org)) {
			console.log('Usage: node registerAndEnrollUser.js Org userID');
			console.log('Org must be either org1 or Org1 or org2 or Org2');
			process.exit(1);
		}

		org = org.toLowerCase();
		await connectToOrgCA(userId, org, org === 'org1' ? mspOrg1 : mspOrg2);
	} catch (error) {
		console.error(`Error in registering and enrolling user: ${error}`);
		process.exit(1);
	}
}

// Execute the main function.
main();
