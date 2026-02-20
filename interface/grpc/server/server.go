package server

import (
	"net"

	"refina-transaction/config/db"
	"refina-transaction/config/env"
	"refina-transaction/internal/repository"

	tpb "github.com/MuhammadMiftaa/Refina-Protobuf/transaction"
	"google.golang.org/grpc"
)

func SetupGRPCServer(dbInstance db.DatabaseClient) (*grpc.Server, *net.Listener, error) {
	lis, err := net.Listen("tcp", ":"+env.Cfg.Server.GRPCPort)
	if err != nil {
		return nil, nil, err
	}

	s := grpc.NewServer()

	transactionServer := &transactionServer{
		txManager:              repository.NewTxManager(dbInstance.GetDB()),
		transactionsRepository: repository.NewTransactionRepository(dbInstance.GetDB()),
		categoryRepository:     repository.NewCategoryRepository(dbInstance.GetDB()),
		outboxRepository:       repository.NewOutboxRepository(dbInstance.GetDB()),
	}
	tpb.RegisterTransactionServiceServer(s, transactionServer)

	return s, &lis, nil
}
