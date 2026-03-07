package dto

// investmentBuyEvent represents the payload published by investment-service on investment.buy
type InvestmentBuyEvent struct {
	ID               string `json:"id"`
	UserID           string `json:"userId"`
	Code             string `json:"code"`
	Quantity         string `json:"quantity"`
	InitialValuation string `json:"initialValuation"`
	Amount           string `json:"amount"`
	Date             string `json:"date"`
	Description      string `json:"description"`
	WalletID         string `json:"walletId"`
}

// investmentSellEvent represents a single sold record from investment-service on investment.sell
type InvestmentSellEvent struct {
	ID           string `json:"id"`
	UserID       string `json:"userId"`
	InvestmentID string `json:"investmentId"`
	Quantity     string `json:"quantity"`
	SellPrice    string `json:"sellPrice"`
	Amount       string `json:"amount"`
	Date         string `json:"date"`
	Description  string `json:"description"`
	Deficit      string `json:"deficit"`
	WalletID     string `json:"walletId"`
}
