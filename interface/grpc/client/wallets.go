package client

import (
	"context"
	"time"

	wpb "github.com/MuhammadMiftaa/Refina-Protobuf/wallet"
)

type WalletClient interface {
	GetWalletByID(ctx context.Context, walletID string) (*wpb.Wallet, error)
	UpdateWallet(ctx context.Context, wallet *wpb.Wallet) (*wpb.Wallet, error)
}

type walletClientImpl struct {
	client wpb.WalletServiceClient
}

func NewWalletClient(grpcClient wpb.WalletServiceClient) WalletClient {
	return &walletClientImpl{
		client: grpcClient,
	}
}

func (w *walletClientImpl) GetWalletByID(ctx context.Context, walletID string) (*wpb.Wallet, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req := &wpb.WalletID{
		Id: walletID,
	}

	return w.client.GetWalletByID(ctx, req)
}

// UpdateWallet converts the full Wallet to an UpdateWalletRequest (which now
// includes balance) and sends it to the wallet-service.
func (w *walletClientImpl) UpdateWallet(ctx context.Context, wallet *wpb.Wallet) (*wpb.Wallet, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req := &wpb.UpdateWalletRequest{
		Id:           wallet.GetId(),
		Name:         wallet.GetName(),
		Number:       wallet.GetNumber(),
		WalletTypeId: wallet.GetWalletTypeId(),
		Balance:      wallet.GetBalance(),
	}

	return w.client.UpdateWallet(ctx, req)
}
