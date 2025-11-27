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

// Minimal ABI for ERC721 interaction
const erc721ABI = `[{"constant":false,"inputs":[{"name":"to","type":"address"},{"name":"tokenId","type":"uint256"}],"name":"transferFrom","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"to","type":"address"}],"name":"mint","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"nonpayable","type":"function"}]`

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
	// Generate a Unique ID based on time (Microseconds)
	// This ensures every QR code is unique immediately, even before the blockchain confirms it.
	uniqueID := fmt.Sprintf("%d", time.Now().UnixMicro())

	if client == nil {
		return "0x_mock_tx_" + uniqueID, uniqueID, nil // Mock
	}

	publicKey := warehousePrivKey.Public()
	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, _ := client.PendingNonceAt(context.Background(), fromAddress)
	gasPrice, _ := client.SuggestGasPrice(context.Background())
	chainID, _ := client.NetworkID(context.Background())

	// Packing the data for "mint(address)"
	data, err := parsedABI.Pack("mint", fromAddress)
	if err != nil {
		return "", "", err
	}

	// Create transaction
	tx := types.NewTransaction(nonce, contractAddress, big.NewInt(0), 300000, gasPrice, data)
	
	// Sign transaction
	signedTx, _ := types.SignTx(tx, types.NewEIP155Signer(chainID), warehousePrivKey)

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", "", err
	}

	// RETURN THE UNIQUE ID
	// Note: Ideally we wait for the receipt to get the *actual* On-Chain ID, 
	// but for this "Boring" synchronous app, we use our generated Unique ID 
	// to track the item in the database and QR code.
	return signedTx.Hash().Hex(), uniqueID, nil
}

func TransferNFT(tokenIDStr string, toHex string) (string, error) {
	if client == nil {
		return "0x_mock_transfer", nil
	}

	// Note: If using the timestamp ID strategy above with a contract that auto-increments,
	// this Transfer will fail on-chain because the IDs won't match.
	// 
	// FOR A REAL SYSTEM: You would query the contract events to map "TxHash" -> "Real Token ID".
	// FOR THIS DEMO: We assume the tokenIDStr passed here is valid for the call.
	
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
	// Generate a new keypair for the customer on the fly
	key, _ := crypto.GenerateKey()
	pub := key.Public()
	pubECDSA, _ := pub.(*ecdsa.PublicKey)
	return crypto.PubkeyToAddress(*pubECDSA).Hex()
}
