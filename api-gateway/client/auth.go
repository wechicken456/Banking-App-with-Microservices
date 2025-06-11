package client

import (
	"fmt"

	proto "buf.build/gen/go/banking-app/auth/grpc/go/_gogrpc"
	"google.golang.org/grpc"
)

type AuthClient struct {
	proto.AuthServiceClient
}

func NewAuthClient(url string, port int) *AuthClient {
	connString := fmt.Sprintf("%s:%d", url, port)
	conn, err := grpc.NewClient(connString)
	if err != nil {
		panic(err)
	}
	client := proto.NewAuthServiceClient(conn)
	return &AuthClient{client}
}
