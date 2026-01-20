package dto

type AttachmentsResponse struct {
	ID            string `json:"id"`
	TransactionID string `json:"transaction_id"`
	Image         string `json:"image"`
	Format        string `json:"format"`
	Size          int64  `json:"size"`
	CreatedAt     string `json:"created_at"`
}

type AttachmentsRequest struct {
	TransactionID string `json:"transaction_id"`
	Image         string `json:"image"`
}

type Attachments struct {
	Files []string `json:"files"`
}
