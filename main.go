package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func main() {
	// 1. Load Environment
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found")
	}

	// 2. Setup SQLite (The Boring Database)
	var err error
	db, err = sql.Open("sqlite3", "./retail.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	initDB()

	// 3. Setup Blockchain Client
	initBlockchain()

	// 4. Setup Router (Chi is standard and lightweight)
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Static files (standard library approach)
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "static"))
	FileServer(r, "/static", filesDir)

	// Routes
	r.Get("/", HandleDashboard)
	r.Get("/pos", HandlePOS)
	r.Post("/mint", HandleMint)
	r.Post("/sell", HandleSell)
	r.Post("/recycle", HandleRecycle)
    
    // Add this new route for fetching details
	r.Get("/product/{token_id}", HandleGetProduct)

	log.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", r)
}

func initDB() {
	// Simple schema. No migrations folder. Just Ensure functionality.
	query := `
	CREATE TABLE IF NOT EXISTS products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		status TEXT DEFAULT 'MANUFACTURED', -- MANUFACTURED, SOLD, RECYCLED
		token_id TEXT,
		tx_hash TEXT
	);
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		wallet_address TEXT UNIQUE,
		points INTEGER DEFAULT 0
	);
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("Database init error:", err)
	}
}

// FileServer conveniently sets up a http.FileServer handler at a specific path
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
