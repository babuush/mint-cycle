package main

type Product struct {
	ID       int
	Name     string
	Status   string // MANUFACTURED, SOLD, RECYCLED
	TokenID  string
	TxHash   string
}

type User struct {
	ID            int
	WalletAddress string
	Points        int
}

type PageData struct {
	Title string
	Data  interface{}
}
