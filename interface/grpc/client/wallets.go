package client

import (
	"context"
	"time"

	wpb "github.com/MuhammadMiftaa/Golang-Refina-Protobuf/wallet"
)

type WalletClient interface {
	GetWalletByID(ctx context.Context, walletID string) (*wpb.Wallet, error)
	UpdateWallet(ctx context.Context, req *wpb.Wallet) (*wpb.Wallet, error)
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

func (w *walletClientImpl) UpdateWallet(ctx context.Context, req *wpb.Wallet) (*wpb.Wallet, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return w.client.UpdateWallet(ctx, req)
}
