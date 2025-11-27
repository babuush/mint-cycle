package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	client           *ethclient.Client
	warehousePrivKey *ecdsa.PrivateKey
	contractAddress  common.Address
	parsedABI        abi.ABI
)

// FIXED ABI: transferFrom now correctly expects 3 arguments (from, to, tokenId)
const erc721ABI = `[{"constant":false,"inputs":[{"name":"from","type":"address"},{"name":"to","type":"address"},{"name":"tokenId","type":"uint256"}],"name":"transferFrom","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"to","type":"address"},{"name":"id","type":"uint256"}],"name":"mint","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"}]`

func initBlockchain() {
	rpcURL := os.Getenv("RPC_URL")
	privKeyHex := os.Getenv("PRIVATE_KEY")
	cAddr := os.Getenv("CONTRACT_ADDRESS")

	if rpcURL == "" || privKeyHex == "" {
		log.Println("⚠️ Blockchain env vars missing. Running in OFFLINE mode.")
		return
	}

	var err error
	client, err = ethclient.Dial(rpcURL)
	if err != nil {
		log.Println("⚠️ Failed to connect to eth client, running offline:", err)
		client = nil
		return
	}

	warehousePrivKey, err = crypto.HexToECDSA(privKeyHex)
	if err != nil {
		log.Fatal("Invalid private key:", err)
	}

	contractAddress = common.HexToAddress(cAddr)
	parsedABI, _ = abi.JSON(strings.NewReader(erc721ABI))

	log.Println("✅ Blockchain connected. Contract:", cAddr)
}

func MintNFTToWarehouse(productName string) (string, string, error) {
	// 1. GENERATE ID: We create the ID based on time (Microseconds)
	uniqueID := fmt.Sprintf("%d", time.Now().UnixMicro())

	if client == nil {
		return "0x_mock_tx_offline_" + uniqueID, uniqueID, nil
	}

	// 2. PREPARE BLOCKCHAIN WRITE
	publicKey := warehousePrivKey.Public()
	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, _ := client.PendingNonceAt(context.Background(), fromAddress)
	gasPrice, _ := client.SuggestGasPrice(context.Background())
	chainID, _ := client.NetworkID(context.Background())

	// Convert our string ID to Big Int for the contract
	tokenIDBig := new(big.Int)
	tokenIDBig.SetString(uniqueID, 10)

	// We pack "mint" with TWO arguments: (to, id)
	data, err := parsedABI.Pack("mint", fromAddress, tokenIDBig)
	if err != nil {
		log.Println("Pack Error:", err)
		return "", uniqueID, err
	}

	tx := types.NewTransaction(nonce, contractAddress, big.NewInt(0), 300000, gasPrice, data)
	signedTx, _ := types.SignTx(tx, types.NewEIP155Signer(chainID), warehousePrivKey)

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Println("⚠️ WRITE FAILED:", err)
		return "0x_failed_tx", uniqueID, nil 
	}

	return signedTx.Hash().Hex(), uniqueID, nil
}

func TransferNFT(tokenIDStr string, toHex string) (string, error) {
	if client == nil {
		return "0x_mock_transfer", nil
	}
	
	tokenID := new(big.Int)
	tokenID.SetString(tokenIDStr, 10)
	toAddress := common.HexToAddress(toHex)

	publicKey := warehousePrivKey.Public()
	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, _ := client.PendingNonceAt(context.Background(), fromAddress)
	gasPrice, _ := client.SuggestGasPrice(context.Background())
	chainID, _ := client.NetworkID(context.Background())

	// transferFrom(from, to, tokenId) - Correctly passes 3 arguments now
	data, err := parsedABI.Pack("transferFrom", fromAddress, toAddress, tokenID)
	if err != nil {
		return "", err
	}

	tx := types.NewTransaction(nonce, contractAddress, big.NewInt(0), 100000, gasPrice, data)
	signedTx, _ := types.SignTx(tx, types.NewEIP155Signer(chainID), warehousePrivKey)

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}
	
	return signedTx.Hash().Hex(), nil
}

func CreateGuestWallet() string {
	key, _ := crypto.GenerateKey()
	pub := key.Public()
	pubECDSA, _ := pub.(*ecdsa.PublicKey)
	return crypto.PubkeyToAddress(*pubECDSA).Hex()
}
