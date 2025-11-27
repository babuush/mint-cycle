package main

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/skip2/go-qrcode"
)

// Helper to parse templates
func render(w http.ResponseWriter, tmpl string, data PageData) {
	t, err := template.ParseFiles("templates/layout.html", "templates/"+tmpl)
	if err != nil {
		log.Println("Template Parse Error:", err)
		http.Error(w, "Template Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.ExecuteTemplate(w, "layout", data)
	if err != nil {
		log.Println("Template Execute Error:", err)
		http.Error(w, "Render Error: "+err.Error(), http.StatusInternalServerError)
	}
}

// 1. Dashboard Handler
func HandleDashboard(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name, status, token_id FROM products ORDER BY id DESC LIMIT 10")
	if err != nil {
		log.Println("DB Error:", err)
		http.Error(w, "Database Error", 500)
		return
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		rows.Scan(&p.ID, &p.Name, &p.Status, &p.TokenID)
		products = append(products, p)
	}

	render(w, "index.html", PageData{Title: "Dashboard", Data: products})
}

// 2. POS Page Handler
func HandlePOS(w http.ResponseWriter, r *http.Request) {
	render(w, "pos.html", PageData{Title: "Point of Sale", Data: nil})
}

// 3. Mint Logic (Admin)
func HandleMint(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")

	// A. Interact with Blockchain
	txHash, tokenID, err := MintNFTToWarehouse(name)
	if err != nil {
		log.Println("Minting failed:", err)
		// Fallback for demo
		tokenID = fmt.Sprintf("%d", time.Now().UnixMicro())
		txHash = "0x_mock_hash_offline"
	}

	// B. Save to SQLite
	_, dbErr := db.Exec("INSERT INTO products (name, token_id, tx_hash) VALUES (?, ?, ?)", name, tokenID, txHash)
	if dbErr != nil {
		log.Println("DB Insert Error:", dbErr)
		http.Error(w, "Database Insert Failed", 500)
		return
	}

	// C. Generate QR Code
	png, _ := qrcode.Encode(tokenID, qrcode.Medium, 256)
	encoded := base64.StdEncoding.EncodeToString(png)

	// HTMX Response
	// FIX: Added txHash back into the template and arguments
	tmpl := `
		<div class="p-4 bg-green-100 border border-green-400 rounded mt-4 animate-pulse">
			<h3 class="font-bold">Minted: %s</h3>
			<p class="text-xs font-mono mb-1">ID: %s</p>
            <p class="text-[10px] text-gray-500 break-all mb-2">Tx: %s</p>
			<img src="data:image/png;base64,%s" class="mt-2 mx-auto border-4 border-white shadow-lg"/>
			<p class="text-center text-sm text-gray-500 mt-2">Print and stick on product</p>
		</div>
	`
	fmt.Fprintf(w, tmpl, name, tokenID, txHash, encoded)
}

// 4. Sell Logic (Transfer Ownership)
func HandleSell(w http.ResponseWriter, r *http.Request) {
	tokenID := r.FormValue("token_id")

	// Create a dummy "guest" wallet for the buyer
	newOwnerAddr := CreateGuestWallet()

	// Blockchain Transfer
	txHash, err := TransferNFT(tokenID, newOwnerAddr)
	if err != nil {
		log.Println("Transfer Error:", err)
		// Even if error, we might want to show UI, but normally we'd return 500.
		// For demo robustness, ensure txHash isn't empty to prevent display issues
		if txHash == "" { txHash = "error_or_offline" }
	}

	// Update DB
	db.Exec("UPDATE products SET status = 'SOLD', tx_hash = ? WHERE token_id = ?", txHash, tokenID)

	// HTMX Response
	// FIX: Removed dangerous [:10] slicing which caused panic if hash was short/empty
	w.Write([]byte(fmt.Sprintf(`
		<div class="bg-blue-100 p-4 rounded text-center animate-pulse">
			<h1 class="text-2xl font-bold text-blue-800">SOLD!</h1>
			<p>NFT transferred to new wallet:</p>
			<code class="text-xs block bg-blue-200 p-1 rounded mt-1 break-all">%s</code>
			<p class="mt-2 text-xs text-gray-500 break-all">Tx: %s</p>
		</div>
	`, newOwnerAddr, txHash)))
}

// 5. Recycle Logic
func HandleRecycle(w http.ResponseWriter, r *http.Request) {
	tokenID := r.FormValue("token_id")

	res, _ := db.Exec("UPDATE products SET status = 'RECYCLED' WHERE token_id = ? AND status = 'SOLD'", tokenID)
	rowsAff, _ := res.RowsAffected()

	if rowsAff == 0 {
		w.Write([]byte(`<div class="bg-red-100 p-2 text-red-700 font-bold text-center">Invalid Item or Not Sold Yet</div>`))
		return
	}

	w.Write([]byte(fmt.Sprintf(`
		<div class="bg-green-600 text-white p-6 rounded text-center">
			<h1 class="text-3xl">♻️ RECYCLED!</h1>
			<p class="text-xl mt-2">+10 Points Added</p>
			<p class="text-sm opacity-75">Product #%s processed.</p>
		</div>
	`, tokenID)))
}

// 6. Get Product Details (Click from Table)
func HandleGetProduct(w http.ResponseWriter, r *http.Request) {
	tokenID := chi.URLParam(r, "token_id")

	var p Product
	err := db.QueryRow("SELECT id, name, status, token_id, tx_hash FROM products WHERE token_id = ?", tokenID).Scan(&p.ID, &p.Name, &p.Status, &p.TokenID, &p.TxHash)
	if err != nil {
		http.Error(w, "Product not found", 404)
		return
	}

	png, _ := qrcode.Encode(p.TokenID, qrcode.Medium, 256)
	encoded := base64.StdEncoding.EncodeToString(png)

	tmpl := `
		<div class="p-4 bg-gray-50 border border-gray-200 rounded-lg animate-fade-in">
            <div class="flex justify-between items-start">
			    <h3 class="font-bold text-lg">%s</h3>
                <span class="text-xs bg-gray-200 px-2 py-1 rounded">%s</span>
            </div>
            <p class="text-xs font-mono text-gray-500 mb-2">ID: %s</p>
			
			<img src="data:image/png;base64,%s" class="mt-4 mx-auto border-4 border-white shadow-lg"/>
			
            <div class="mt-4 text-center">
                <p class="text-xs text-gray-400 mb-1">Latest Blockchain Tx:</p>
                <a href="https://amoy.polygonscan.com/tx/%s" target="_blank" class="text-[10px] text-blue-500 hover:underline bg-blue-50 p-1 rounded block break-all cursor-pointer" title="View on PolygonScan">
                    %s
                </a>
            </div>
		</div>
	`
	fmt.Fprintf(w, tmpl, p.Name, p.Status, p.TokenID, encoded, p.TxHash, p.TxHash)
}
