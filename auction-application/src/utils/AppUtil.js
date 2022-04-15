"use strict";

const { Wallets } = require("fabric-network");
const fs = require("fs");
const path = require("path");

/**
 * @description This function is used to load the common connection profile.
 * @param {string} orgName - Name of the organization.
 * @returns {Object} The common connection profile.
 */
exports.buildCCPOrg = (orgName) => {
  // Load the common connection configuration file.
  const ccpPath = path.resolve(
    __dirname,
    "..",
    "..",
    "..",
    "test-network",
    "organizations",
    "peerOrganizations",
    `${orgName}.example.com`,
    `connection-${orgName}.json`
  );

  // Check if the common connection profile exists.
  const fileExists = fs.existsSync(ccpPath);
  if (!fileExists) {
    throw new Error(`No such file or directory: ${ccpPath}`);
  }

  // Read the common connection profile.
  const contents = fs.readFileSync(ccpPath, "utf8");

  // Build a JSON object from the file contents.
  const ccp = JSON.parse(contents);

  console.log(`Loaded the network configuration located at ${ccpPath}`);
  return ccp;
};

/**
 * @description This function is used to create a new wallet.
 * @param {string} walletPath - Directory path to the wallet.
 * @returns {*} The wallet object.
 */
exports.buildWallet = async (walletPath) => {
  // Create a new wallet, note that wallet is for managing identities.
  let wallet;
  if (walletPath) {
    wallet = await Wallets.newFileSystemWallet(walletPath);
    console.log(`Built a file system wallet at ${walletPath}`);
  } else {
    wallet = await Wallets.newInMemoryWallet();
    console.log("Built an in memory wallet");
  }

  return wallet;
};

/**
 * @description This function is used to stringify a JSON object.
 * @param {string} inputString - The input string.
 * @returns {string} The pretty string.
 */
exports.prettyJSONString = (inputString) => {
  return inputString
    ? JSON.stringify(JSON.parse(inputString), null, 2)
    : inputString;
};
