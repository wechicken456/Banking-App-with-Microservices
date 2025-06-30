package client

import (
	proto "buf.build/gen/go/banking-app/auth/grpc/go/_gogrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthClient struct {
	proto.AuthServiceClient
}

func NewAuthClient(connString string) *AuthClient {
	conn, err := grpc.NewClient(connString, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	client := proto.NewAuthServiceClient(conn)
	return &AuthClient{client}
}
