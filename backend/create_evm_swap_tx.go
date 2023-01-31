package backend

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/incognitochain/bridge-eth/bridge/vault"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/evmproof"
)

func createOutChainSwapTx(network string, incTxHash string, isUnifiedToken bool, txType int) (*wcommon.ExternalTxStatus, error) {
	var result wcommon.ExternalTxStatus

	var proof *evmproof.DecodedProof
	incognitoFullnodeRPC := ""
	proof, err := evmproof.GetAndDecodeBurnProofUnifiedToken(incognitoFullnodeRPC, incTxHash, 0)
	if err != nil {
		return nil, err
	}
	if proof == nil {
		return nil, fmt.Errorf("could not get proof for network %s", "eth")
	}

	if len(proof.InstRoots) == 0 {
		return nil, fmt.Errorf("could not get proof for network %s", "eth")
	}
	evmPrivateKey := ""
	privKey, _ := crypto.HexToECDSA(evmPrivateKey)
	i := 0
retry:
	if i == 10 {
		return nil, errors.New("submit tx outchain failed")
	}
	evmClient, err := ethclient.Dial("your eth rpc endpoint")
	if err != nil {
		return nil, err
	}

	c, err := vault.NewVault(common.HexToAddress("incognitoVaultContractAddress"), evmClient)
	if err != nil {
		return nil, err
	}

	gasPrice, err := evmClient.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privKey, new(big.Int).SetInt64(1))
	if err != nil {
		return nil, err
	}

	gasPrice = gasPrice.Mul(gasPrice, big.NewInt(12))
	gasPrice = gasPrice.Div(gasPrice, big.NewInt(10))

	auth.GasPrice = gasPrice
	auth.GasLimit = wcommon.EVMGasLimitETH

	result.Type = txType
	result.Network = network
	result.IncRequestTx = incTxHash
	tx, err := evmproof.ExecuteWithBurnProof(c, auth, proof)
	if err != nil {
		if strings.Contains(err.Error(), "insufficient funds") {
			return nil, errors.New("submit tx outchain failed err insufficient funds")
		}
		return nil, err
	}
	result.Txhash = tx.Hash().String()
	result.Status = wcommon.StatusPending
	result.Nonce = tx.Nonce()
	if result.Txhash == "" {
		i++
		time.Sleep(2 * time.Second)
		goto retry
	}
	return &result, nil
}
