package dto

type WalletsResponse struct {
	ID           string  `json:"id"`
	UserID       string  `json:"user_id"`
	WalletTypeID string  `json:"wallet_type_id"`
	Name         string  `json:"name"`
	Number       string  `json:"number"`
	Balance      float64 `json:"balance"`
}
