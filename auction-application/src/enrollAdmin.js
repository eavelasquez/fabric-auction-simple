'use strict';

const path = require('path');

const {
  buildCCPOrg,
  buildWallet,
  checkArgs,
  handleError,
} = require('./utils/AppUtil');
const { buildCAClient, enrollAdmin } = require('./utils/CAUtil');

const mspOrg1 = 'Org1MSP';
const mspOrg2 = 'Org2MSP';

/**
 * @description Connects to the CA of the specified Org and enrolls the admin.
 * @param {string} orgName - The name of the organization e.g. org1 or org2.
 * @param {string} mspOrg - The MSP ID of the Org.
 * @returns {Promise<void>}
 */
async function connectToOrgCA(orgName, mspOrg) {
  console.log(`\n--> Enrolling the ${orgName} CA admin`);
  const ccpOrg = buildCCPOrg(orgName);
  const caOrgClient = buildCAClient(ccpOrg, `ca.${orgName}.example.com`);

  const walletPathOrg = path.join(__dirname, `wallet/${orgName}`);
  const walletOrg = await buildWallet(walletPathOrg);

  await enrollAdmin(caOrgClient, walletOrg, mspOrg);
}

// Argument list for the script.
const fileAndArgs = 'enrollAdmin.js <org>';

/**
 * Enrolls the admin of the Org1 CA or Org2 CA.
 */
async function main() {
  try {
    checkArgs(
      process.argv[2] === undefined,
      fileAndArgs,
      'Missing required argument: org'
    );

    let [, , org] = process.argv;
    checkArgs(
      /^(org1|Org1|org2|Org2)$/.test(org),
      fileAndArgs,
      'Org must be either org1 or Org1 or org2 or Org2'
    );

    org = org.toLowerCase();

    await connectToOrgCA(org, org === 'org1' ? mspOrg1 : mspOrg2);
  } catch (error) {
    handleError('Failed to run the enroll admin', error);
  }
}

// Execute the main function.
main();
