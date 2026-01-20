package external

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"refina-transaction/config/env"
	"refina-transaction/internal/types/dto"
)

type WalletClient struct {
	baseURL    string
	httpClient *http.Client
}

type WalletAPIResponse struct {
	StatusCode int                 `json:"statusCode"`
	Status     bool                `json:"status"`
	Message    string              `json:"message"`
	Data       dto.WalletsResponse `json:"data"`
}

type WalletUpdateRequest struct {
	UserID       string  `json:"user_id"`
	WalletTypeID string  `json:"wallet_type_id"`
	Name         string  `json:"name"`
	Number       string  `json:"number"`
	Balance      float64 `json:"balance"`
}

func NewWalletClient() *WalletClient {
	return &WalletClient{
		baseURL: env.Cfg.WalletService.BaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (wc *WalletClient) GetWalletByID(ctx context.Context, walletID string) (*dto.WalletsResponse, error) {
	url := fmt.Sprintf("%s/wallets/%s", wc.baseURL, walletID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := wc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call wallet service: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wallet service returned status %d: %s", resp.StatusCode, string(body))
	}

	var apiResponse WalletAPIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !apiResponse.Status {
		return nil, errors.New(apiResponse.Message)
	}

	return &apiResponse.Data, nil
}

func (wc *WalletClient) UpdateWallet(ctx context.Context, walletID string, wallet *dto.WalletsResponse) (*dto.WalletsResponse, error) {
	url := fmt.Sprintf("%s/wallets/%s", wc.baseURL, walletID)

	updateReq := WalletUpdateRequest{
		UserID:       wallet.UserID,
		WalletTypeID: wallet.WalletTypeID,
		Name:         wallet.Name,
		Number:       wallet.Number,
		Balance:      wallet.Balance,
	}

	jsonData, err := json.Marshal(updateReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := wc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call wallet service: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wallet service returned status %d: %s", resp.StatusCode, string(body))
	}

	var apiResponse WalletAPIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !apiResponse.Status {
		return nil, errors.New(apiResponse.Message)
	}

	return &apiResponse.Data, nil
}
