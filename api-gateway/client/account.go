package client

import (
	proto "buf.build/gen/go/banking-app/account/grpc/go/_gogrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AccountClient struct {
	proto.AccountServiceClient
}

func NewAccountClient(connString string) *AccountClient {
	conn, err := grpc.NewClient(connString, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	client := proto.NewAccountServiceClient(conn)
	return &AccountClient{client}
}
