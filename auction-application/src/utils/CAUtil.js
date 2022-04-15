"use strict";

/**
 * @description This function is used to build a new Fabric CA client.
 * @param {*} FabricCAServices - The Fabric CA services class.
 * @param {*} ccp - The common connection profile.
 * @param {string} caHostName - The CA host name.
 * @returns {*} The Fabric CA services object.
 */
exports.buildCAClient = (FabricCAServices, ccp, caHostName) => {
  // Create a new CA client for interacting with the CA.
  const caInfo = ccp.certificateAuthorities[caHostName]; // lookup CA details from config.
  const caTLSCACerts = caInfo.tlsCACerts.pem;
  const caClient = new FabricCAServices(
    caInfo.url,
    { trustedRoots: caTLSCACerts, verify: false },
    caInfo.caName
  );

  console.log(`Built a CA Client named ${caInfo.caName}`);
  return caClient;
};
