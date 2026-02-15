package server

import (
	"net"

	"refina-transaction/config/db"
	"refina-transaction/config/env"
	"refina-transaction/internal/repository"

	tpb "github.com/MuhammadMiftaa/Refina-Protobuf/transaction"
	"google.golang.org/grpc"
)

func SetupGRPCServer() (*grpc.Server, *net.Listener, error) {
	lis, err := net.Listen("tcp", ":"+env.Cfg.Server.GRPCPort)
	if err != nil {
		return nil, nil, err
	}

	s := grpc.NewServer()

	transactionServer := &transactionServer{
		transactionsRepository: repository.NewTransactionRepository(db.DB),
	}
	tpb.RegisterTransactionServiceServer(s, transactionServer)

	return s, &lis, nil
}
