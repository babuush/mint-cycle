package main

import (
	"context"
	"crypto/ecdsa"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum"
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

// ABI Updated to include 'currentTokenId' view function
const erc721ABI = `[{"constant":true,"inputs":[],"name":"currentTokenId","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"to","type":"address"},{"name":"tokenId","type":"uint256"}],"name":"transferFrom","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"to","type":"address"}],"name":"mint","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"nonpayable","type":"function"}]`

func initBlockchain() {
	rpcURL := os.Getenv("RPC_URL")
	privKeyHex := os.Getenv("PRIVATE_KEY")
	cAddr := os.Getenv("CONTRACT_ADDRESS")

	if rpcURL == "" || privKeyHex == "" {
		log.Println("Blockchain env vars missing. Running in OFFLINE mode.")
		return
	}

	var err error
	client, err = ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatal("Failed to connect to eth client:", err)
	}

	warehousePrivKey, err = crypto.HexToECDSA(privKeyHex)
	if err != nil {
		log.Fatal("Invalid private key:", err)
	}

	contractAddress = common.HexToAddress(cAddr)
	parsedABI, _ = abi.JSON(strings.NewReader(erc721ABI))

	log.Println("Blockchain connected. Contract:", cAddr)
}

func MintNFTToWarehouse(productName string) (string, string, error) {
	if client == nil {
		return "0x_mock_tx", "999", nil
	}

	// 1. READ: Get the current Token ID from the contract to keep DB in sync
	// We call the 'currentTokenId' view function
	var currentId big.Int
	callData, _ := parsedABI.Pack("currentTokenId")
	msg := ethereum.CallMsg{To: &contractAddress, Data: callData}
	output, err := client.CallContract(context.Background(), msg, nil)
	
	tokenID := big.NewInt(1) // Default if call fails
	if err == nil {
		parsedABI.UnpackIntoInterface(&currentId, "currentTokenId", output)
		// The next ID will be current + 1
		tokenID.Add(&currentId, big.NewInt(1))
	} else {
        log.Println("Warning: Could not fetch token ID, using prediction", err)
    }

	// 2. WRITE: Prepare Mint Transaction
	publicKey := warehousePrivKey.Public()
	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, _ := client.PendingNonceAt(context.Background(), fromAddress)
	gasPrice, _ := client.SuggestGasPrice(context.Background())
	chainID, _ := client.NetworkID(context.Background())

	data, err := parsedABI.Pack("mint", fromAddress)
	if err != nil {
		return "", "", err
	}

	tx := types.NewTransaction(nonce, contractAddress, big.NewInt(0), 300000, gasPrice, data)
	signedTx, _ := types.SignTx(tx, types.NewEIP155Signer(chainID), warehousePrivKey)

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", "", err
	}

	return signedTx.Hash().Hex(), tokenID.String(), nil
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

	// transferFrom(from, to, tokenId)
	data, err := parsedABI.Pack("transferFrom", fromAddress, toAddress, tokenID)
	if err != nil {
		return "", err
	}

	tx := types.NewTransaction(nonce, contractAddress, big.NewInt(0), 100000, gasPrice, data)
	signedTx, _ := types.SignTx(tx, types.NewEIP155Signer(chainID), warehousePrivKey)

	err = client.SendTransaction(context.Background(), signedTx)
	return signedTx.Hash().Hex(), err
}

func CreateGuestWallet() string {
	key, _ := crypto.GenerateKey()
	pub := key.Public()
	pubECDSA, _ := pub.(*ecdsa.PublicKey)
	return crypto.PubkeyToAddress(*pubECDSA).Hex()
}
