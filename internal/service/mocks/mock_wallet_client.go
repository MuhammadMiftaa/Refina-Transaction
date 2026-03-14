package mocks

import (
	"context"

	wpb "github.com/MuhammadMiftaa/Refina-Protobuf/wallet"
	"github.com/stretchr/testify/mock"
)

type MockWalletClient struct {
	mock.Mock
}

func (m *MockWalletClient) GetWalletByID(ctx context.Context, walletID string) (*wpb.Wallet, error) {
	args := m.Called(ctx, walletID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*wpb.Wallet), args.Error(1)
}

func (m *MockWalletClient) UpdateWallet(ctx context.Context, wallet *wpb.Wallet) (*wpb.Wallet, error) {
	args := m.Called(ctx, wallet)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*wpb.Wallet), args.Error(1)
}
