package bridge

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/incognitochain/bridge-eth/bridge/vault"
	"github.com/incognitochain/go-incognito-sdk-v2/coin"
	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
)

const ADDRESS_0 = "0x0000000000000000000000000000000000000000"

var encodeBufferPool = sync.Pool{
	New: func() interface{} { return new(bytes.Buffer) },
}

func getETHDepositProof(
	evmClient *ethclient.Client,
	txHashStr string,
) (*big.Int, string, uint, []string, string, string, bool, string, uint64, string, bool, error) {
	var contractID string
	var paymentaddress string
	var otaStr string
	var shieldAmount uint64
	var isRedeposit bool
	var logResult string
	var isTxPass bool

	txHash := common.Hash{}
	err := txHash.UnmarshalText([]byte(txHashStr))
	if err != nil {
		return nil, "", 0, nil, "", "", false, "", 0, "", isTxPass, err
	}
	txReceipt, err := evmClient.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		return nil, "", 0, nil, "", "", false, "", 0, "", isTxPass, err
	}

	txIndex := txReceipt.TransactionIndex
	blockHash := txReceipt.BlockHash.String()
	blockNumber := txReceipt.BlockNumber

	blk, err := evmClient.BlockByHash(context.Background(), txReceipt.BlockHash)
	if err != nil {
		return nil, "", 0, nil, "", "", false, "", 0, "", isTxPass, err
	}

	if txReceipt.Status == 1 {
		isTxPass = true
	}
	siblingTxs := blk.Body().Transactions
	keybuf := new(bytes.Buffer)
	receiptTrie := new(trie.Trie)
	receipts := make([]*types.Receipt, 0)

	for i, siblingTx := range siblingTxs {
		//NOTE: skip tx that doesn't exist
		siblingReceipt, err := evmClient.TransactionReceipt(context.Background(), siblingTx.Hash())
		if err != nil {
			if siblingTx.To().String() == ADDRESS_0 {
				log.Println("evmClient.TransactionReceipt error:", err)
				continue
			}
			return nil, "", 0, nil, "", "", false, "", 0, "", isTxPass, err
		}
		if i == len(siblingTxs)-1 {
			txData, _, err := evmClient.TransactionByHash(context.Background(), siblingTx.Hash())
			if err != nil {
				return nil, "", 0, nil, "", "", false, "", 0, "", isTxPass, err
			}
			from, err := evmClient.TransactionSender(context.Background(), txData, txReceipt.BlockHash, uint(i))
			if err != nil {
				return nil, "", 0, nil, "", "", false, "", 0, "", isTxPass, err
			}
			if txData.To() != nil {
				if txData.To().String() == ADDRESS_0 && from.String() == ADDRESS_0 {
					break
				}
			}
		}
		receipts = append(receipts, siblingReceipt)
		time.Sleep(100 * time.Millisecond)
	}

	receiptList := types.Receipts(receipts)
	receiptTrie.Reset()

	valueBuf := encodeBufferPool.Get().(*bytes.Buffer)
	defer encodeBufferPool.Put(valueBuf)

	vaultABI, err := abi.JSON(strings.NewReader(vault.VaultABI))
	if err != nil {
		fmt.Println("abi.JSON", err.Error())
		return nil, "", 0, nil, "", "", false, "", 0, "", isTxPass, err
	}
	for _, d := range txReceipt.Logs {
		switch len(d.Data) {
		case 256, 288:
			topicHash := strings.ToLower(d.Topics[0].String())
			if !strings.Contains(topicHash, "00b45d95b5117447e2fafe7f34def913ff3ba220e4b8688acf37ae2328af7a3d") {
				continue
			}
			if paymentaddress == "" && otaStr == "" {
				unpackResult, err := vaultABI.Unpack("Redeposit", d.Data)
				if err != nil {
					log.Println("unpackResult3 err", err)
					continue
				}
				if len(unpackResult) < 3 {
					err = errors.New(fmt.Sprintf("Unpack event not match data needed %v\n", unpackResult))
					log.Println("len(unpackResult) err", err)
					return nil, "", 0, nil, "", "", false, "", 0, "", isTxPass, err
				}
				contractID = unpackResult[0].(common.Address).String()
				amount := unpackResult[2].(*big.Int)
				shieldAmount = amount.Uint64()
				var ok bool
				paymentaddress, ok = unpackResult[1].(string)
				if !ok {
					OTAReceiver := unpackResult[1].([]byte)
					newOTA := coin.OTAReceiver{}
					err = newOTA.SetBytes(OTAReceiver)
					if err != nil {
						log.Println("unpackResult4 err", err)
						continue
					}
					isRedeposit = true
					otaStr = newOTA.String()
				}
			}
		default:
			unpackResult, err := vaultABI.Unpack("ExecuteFnLog", d.Data)
			if err != nil {
				log.Println("unpackResult2 err", err)
				continue
			} else {
				logResult = fmt.Sprintf("%s", unpackResult)
				log.Println("logResult", logResult)
			}
		}
	}
	var indexBuf []byte
	for i := 1; i < receiptList.Len() && i <= 0x7f; i++ {
		indexBuf = rlp.AppendUint64(indexBuf[:0], uint64(i))
		value := encodeForDerive(receiptList, i, valueBuf)
		receiptTrie.Update(indexBuf, value)
	}
	if receiptList.Len() > 0 {
		indexBuf = rlp.AppendUint64(indexBuf[:0], 0)
		value := encodeForDerive(receiptList, 0, valueBuf)
		receiptTrie.Update(indexBuf, value)
	}
	for i := 0x80; i < receiptList.Len(); i++ {
		indexBuf = rlp.AppendUint64(indexBuf[:0], uint64(i))
		value := encodeForDerive(receiptList, i, valueBuf)
		receiptTrie.Update(indexBuf, value)
	}

	// Constructing the proof for the current receipt (source: go-ethereum/trie/proof.go)
	proof := light.NewNodeSet()
	keybuf.Reset()
	err = rlp.Encode(keybuf, uint(txIndex))
	if err != nil {
		return nil, "", 0, nil, "", "", false, "", 0, "", isTxPass, err
	}
	err = receiptTrie.Prove(keybuf.Bytes(), 0, proof)
	if err != nil {
		return nil, "", 0, nil, "", "", false, "", 0, "", isTxPass, err
	}
	nodeList := proof.NodeList()
	encNodeList := make([]string, 0)
	for _, node := range nodeList {
		str := base64.StdEncoding.EncodeToString(node)
		encNodeList = append(encNodeList, str)
	}
	return blockNumber, blockHash, uint(txIndex), encNodeList, contractID, paymentaddress, isRedeposit, otaStr, shieldAmount, logResult, isTxPass, nil
}

func GetProof(txhash string, endpoint string) (*wcommon.EVMProofRecordData, *incclient.EVMDepositProof, error) {
	evmClient, err := ethclient.Dial(endpoint)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	blockNumber, blockHash, txIdx, proof, contractID, paymentAddr, isRedeposit, otaStr, amount, _, isTxPass, err := getETHDepositProof(evmClient, txhash)
	if err != nil {
		return nil, nil, err
	}
	if len(proof) == 0 {
		return nil, nil, fmt.Errorf("invalid proof or tx not found")
	}
	depositProof := incclient.NewETHDepositProof(uint(blockNumber.Int64()), common.HexToHash(blockHash), txIdx, proof)

	proofBytes, _ := json.Marshal(proof)
	if len(proof) == 0 {
		return nil, nil, fmt.Errorf("invalid proof or tx not found")
	}
	result := wcommon.EVMProofRecordData{
		Proof:       string(proofBytes),
		BlockNumber: blockNumber.Uint64(),
		BlockHash:   blockHash,
		TxIndex:     txIdx,
		ContractID:  contractID,
		PaymentAddr: paymentAddr,
		IsRedeposit: isRedeposit,
		IsTxPass:    isTxPass,
		OTAStr:      otaStr,
		Amount:      amount,
		Network:     "eth",
	}

	return &result, depositProof, nil
}

func encodeForDerive(list types.DerivableList, i int, buf *bytes.Buffer) []byte {
	buf.Reset()
	list.EncodeIndex(i, buf)
	// It's really unfortunate that we need to do perform this copy.
	// StackTrie holds onto the values until Hash is called, so the values
	// written to it must not alias.
	return common.CopyBytes(buf.Bytes())
}
