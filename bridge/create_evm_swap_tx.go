package bridge

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/incognitochain/bridge-eth/bridge/vault"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/evmproof"
)

func CreateOutChainSwapTx(incTxHash string, isUnifiedToken bool) error {
	var proof *evmproof.DecodedProof
	incognitoFullnodeRPC := ""
	proof, err := evmproof.GetAndDecodeBurnProofUnifiedToken(incognitoFullnodeRPC, incTxHash, 0)
	if err != nil {
		return err
	}
	if proof == nil {
		return fmt.Errorf("could not get proof for network %s", "eth")
	}

	if len(proof.InstRoots) == 0 {
		return fmt.Errorf("could not get proof for network %s", "eth")
	}
	evmPrivateKey := ""
	privKey, _ := crypto.HexToECDSA(evmPrivateKey)

	evmClient, err := ethclient.Dial("your eth rpc endpoint")
	if err != nil {
		return err
	}

	c, err := vault.NewVault(common.HexToAddress("incognitoVaultContractAddress"), evmClient)
	if err != nil {
		return err
	}

	gasPrice, err := evmClient.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privKey, new(big.Int).SetInt64(1))
	if err != nil {
		return err
	}

	gasPrice = gasPrice.Mul(gasPrice, big.NewInt(12))
	gasPrice = gasPrice.Div(gasPrice, big.NewInt(10))

	auth.GasPrice = gasPrice
	auth.GasLimit = wcommon.EVMGasLimitETH

	tx, err := evmproof.ExecuteWithBurnProof(c, auth, proof)
	if err != nil {
		if strings.Contains(err.Error(), "insufficient funds") {
			return errors.New("submit tx outchain failed err insufficient funds")
		}
		return err
	}
	log.Println("tx outchain:", tx.Hash().String())
	return nil
}
