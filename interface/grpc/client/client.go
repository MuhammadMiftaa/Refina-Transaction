package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"refina-transaction/config/env"

	wpb "github.com/MuhammadMiftaa/Golang-Refina-Protobuf/wallet"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type GRPCClientManager struct {
	walletClient wpb.WalletServiceClient

	connections []*grpc.ClientConn
	mu          sync.RWMutex
}

var (
	manager *GRPCClientManager
	once    sync.Once
)

// GetManager returns singleton instance of GRPCClientManager
func GetManager() *GRPCClientManager {
	once.Do(func() {
		manager = &GRPCClientManager{
			connections: make([]*grpc.ClientConn, 0),
		}
	})
	return manager
}

// SetupGRPCClient sets up all gRPC clients
func (m *GRPCClientManager) SetupGRPCClient() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Setup Wallet Client
	if err := m.setupWalletClient(env.Cfg.GRPCConfig.WalletAddress); err != nil {
		return fmt.Errorf("failed to setup wallet client: %w", err)
	}

	return nil
}

func (m *GRPCClientManager) setupWalletClient(address string) error {
	conn, err := m.createConnection(address)
	if err != nil {
		return err
	}

	m.walletClient = wpb.NewWalletServiceClient(conn)
	m.connections = append(m.connections, conn)
	return nil
}

func (m *GRPCClientManager) createConnection(address string) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(10*1024*1024),
			grpc.MaxCallSendMsgSize(10*1024*1024),
		),
	}

	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", address, err)
	}

	return conn, nil
}

// GetWalletClient returns the wallet service client
func (m *GRPCClientManager) GetWalletClient() wpb.WalletServiceClient {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.walletClient
}

// Shutdown closes all gRPC connections gracefully
func (m *GRPCClientManager) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errors []error
	for i, conn := range m.connections {
		if err := conn.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close connection %d: %w", i, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errors)
	}

	return nil
}
