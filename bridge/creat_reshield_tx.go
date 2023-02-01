package bridge

import (
	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
)

func submitProofTx(tokenID string, pUTokenID string, key string, txhash string) (string, error) {
	_, proof, err := getProof(txhash)
	if err != nil {
		return "", err
	}
	incClient, err := incclient.NewIncClient("https://mainnet.incognito.org/fullnode", "", 2, "mainnet")
	if err != nil {
		return "", err
	}
	networkID := 1
	if tokenID == wcommon.PRV_TOKENID {
		result, err := incClient.CreateAndSendIssuingPRVPeggingRequestTransaction(key, *proof, networkID-1)
		if err != nil {
			return result, err
		}
		return result, err
	}
	if tokenID == pUTokenID {
		result, err := incClient.CreateAndSendIssuingEVMRequestTransaction(key, tokenID, *proof, networkID-1)
		if err != nil {
			return result, err
		}
		return result, nil
	}
	result, err := incClient.CreateAndSendIssuingpUnifiedRequestTransaction(key, tokenID, pUTokenID, *proof, networkID)
	if err != nil {
		return result, err
	}
	return result, err
}
