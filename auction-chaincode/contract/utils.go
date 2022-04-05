package contract

import (
	"encoding/base64"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/pkg/statebased"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// GetSubmittingClientIdentity is an internal utility function to get submitting client identity.
func (c *AuctionContract) GetSubmittingClientIdentity(ctx contractapi.TransactionContextInterface) (string, error) {
	// Get the MSP ID of submitting client identity.
	b64ID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("Failed to get client identity: %v", err)
	}

	// Decode the base64 encoded ID.
	decodeID, err := base64.StdEncoding.DecodeString(b64ID)
	if err != nil {
		return "", fmt.Errorf("Failed to base64 decode client identity: %v", err)
	}

	return string(decodeID), nil
}

// setAssetStateBasedEndorsement sets the state of the asset to be endorsed by the specified org.
func setAssetStateBasedEndorsement(ctx contractapi.TransactionContextInterface, auctionID string, orgToEndorse string) error {
	// Get the endorsement policy.
	endorsementPolicy, err := statebased.NewStateEP(nil)
	if err != nil {
		return fmt.Errorf("Failed to create endorsement policy: %v", err)
	}

	// Add the org to endorse to the policy.
	err = endorsementPolicy.AddOrgs(statebased.RoleTypePeer, orgToEndorse)
	if err != nil {
		return fmt.Errorf("Failed to add org to endorsement policy: %v", err)
	}

	// Set the endorsement policy.
	policy, err := endorsementPolicy.Policy()
	if err != nil {
		return fmt.Errorf("Failed to create endorsement policy bytes from org: %v", err)
	}

	// Set validation parameter on the asset.
	err = ctx.GetStub().SetStateValidationParameter(auctionID, policy)
	if err != nil {
		return fmt.Errorf("FAiled to set validation parameter on auction: %v", err)
	}

	return nil
}

// addAssetStateBasedEndorsement adds a new organization as an endorser of the auction
func addAssetStateBasedEndorsement(ctx contractapi.TransactionContextInterface, auctionID string, orgsToEndorse string) error {
	// Get the endorsement policy.
	endorsementPolicy, err := ctx.GetStub().GetStateValidationParameter(auctionID)
	if err != nil {
		return fmt.Errorf("Failed to get endorsement policy: %v", err)
	}

	// Create a new endorsement policy from the existing policy.
	newEndorsementPolicy, err := statebased.NewStateEP(endorsementPolicy)
	if err != nil {
		return fmt.Errorf("Failed to create new endorsement policy: %v", err)
	}

	// Add the org to endorse to the policy.
	err = newEndorsementPolicy.AddOrgs(statebased.RoleTypePeer, orgsToEndorse)
	if err != nil {
		return fmt.Errorf("Failed to add org to endorsement policy: %v", err)
	}

	// Get the new endorsement policy bytes.
	policy, err := newEndorsementPolicy.Policy()
	if err != nil {
		return fmt.Errorf("Failed to create endorsement policy bytes from org: %v", err)
	}

	// Set validation parameter on the asset.
	err = ctx.GetStub().SetStateValidationParameter(auctionID, policy)
	if err != nil {
		return fmt.Errorf("FAiled to set validation parameter on auction: %v", err)
	}

	return nil
}

// getCollectionName is an internal utility function to get collection of submitting client identity.
func getCollectionName(ctx contractapi.TransactionContextInterface) (string, error) {
	// Get the MSP ID of submitting client identity.
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("Failed to get verified MSP ID of submitting client identity: %v", err)
	}

	// Create the collection name.
	orgCollectionName := "_implicit_org_" + clientMSPID

	return orgCollectionName, nil
}

// verifyClientOrgMatchesPeerOrg is an internal utility function used to verify that client org
// matches peer org id.
func verifyClientOrgMatchesPeerOrg(ctx contractapi.TransactionContextInterface) error {
	// Get the MSP ID of client identity.
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("Failed to get verified MSP ID of client identity: %v", err)
	}

	// Get the MSP ID of peer.
	peerMSPID, err := shim.GetMSPID()
	if err != nil {
		return fmt.Errorf("Failed to get verified MSP ID of peer: %v", err)
	}

	// Verify that MSP ID of client identity matches MSP ID of peer org.
	if clientMSPID != peerMSPID {
		return fmt.Errorf("Client MSP ID %s is not authorized to read or write private data from an org %s peer", clientMSPID, peerMSPID)
	}

	return nil
}

// contains returns true if the string is in the slice, otherwise false
func contains(s []string, str string) bool {
	for _, a := range s {
		if a == str {
			return true
		}
	}

	return false
}
