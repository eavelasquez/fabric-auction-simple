"use strict";

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
