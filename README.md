# Mint & Cycle: The "Boring" Phygital Stack

A pragmatic Point of Sale and Recycling system using Go, SQLite, and HTMX.

## Setup

### Install Dependencies

```bash
go mod tidy

Environment Configuration
Create a .env file in the root:

# Polygon Amoy Testnet RPC
RPC_URL=https://rpc-amoy.polygon.technology/

# Your Warehouse Wallet Private Key (No 0x prefix)
PRIVATE_KEY=your_private_key_here

# Your deployed ERC721 Contract Address
CONTRACT_ADDRESS=0x...

Run the System

go run .

Usage
Open http://localhost:8080.

Dashboard
Use the form to "Mint" a new product. This mocks the blockchain interaction and prints a QR code to the screen.

POS
Open http://localhost:8080/pos on a mobile device (or enable webcam on laptop).

Sell
Scan the Minted QR code. The system transfers the NFT to a new generated wallet.

Recycle
Switch tabs, scan again. The system marks it recycled and awards points.

Architecture
State: retail.db (SQLite)
Logic: Go (Standard Lib + Chi)
UI: Server-side Templates + HTMX
```

