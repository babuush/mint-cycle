# mint cycle — Phygital Retail System

A Point of Sale and Recycling system that connects physical inventory to the blockchain using **boring, reliable technology**.

---

## 🚀 Tech Stack

We prioritize **robustness and simplicity**.  

### 🟢 Core and Backend
- **Language:** Go (Golang) 1.23+
- **Router:** go-chi/chi  
- **Database:** mattn go-sqlite3  
  (Embedded, single-file SQL database)

### 🔵 Frontend (No Build Tools)
- **Rendering:** Go html/template  
- **Interactivity:** HTMX  
- **Styling:** Tailwind CSS (CDN)  
- **Hardware:** html5-qrcode for QR scanning  

### 🟣 Blockchain
- **Network:** Polygon Amoy Testnet  
- **Client:** go-ethereum  
- **Standard:** ERC-721 (NFTs)

---

## 📁 Directory Structure

```
mint-cycle/
├── main.go
├── handlers.go
├── blockchain.go
├── models.go
├── go.mod
├── .env
└── templates/
    ├── layout.html
    ├── index.html
    └── pos.html
```

---

## ⚙️ Setup

### 1. Install Dependencies

```bash
go mod tidy
```

### 2. Environment Variables

Create a `.env` file:

```env
RPC_URL=https://rpc-amoy.polygon.technology/
PRIVATE_KEY=your_private_key_here
CONTRACT_ADDRESS=0x...
```

### 3. Fund Your Wallet

- Use the [Polygon Faucet](https://faucet.polygon.technology/) (Amoy network)
- Request POL test tokens for gas

---

## ▶️ Run the System

```bash
go run .
```

---

## 🧭 Usage Flow

### 1. Dashboard

Open: `http://localhost:8080`

- Mint new products using the "Factory Output" form
- A blockchain transaction is created
- A unique QR code is generated

### 2. POS Terminal

Open: `http://localhost:8080/pos`

Use on a mobile device or any webcam-enabled laptop.

**Sell Mode**
- Scan the product QR code
- NFT transfers from the Warehouse Wallet to a Guest Wallet

**Recycle Mode**
- Scan the same code again
- Status updates to RECYCLED
- Guest earns reward points

---

## 🔍 Verify Transactions

- Click any entry in the Dashboard table
- View on-chain details
- Copy the transaction hash and search it on [PolygonScan](https://amoy.polygonscan.com/) to verify blockchain execution
