package handler

import (
	"auth/proto"
	"auth/service"
)

type AuthHandler struct {
	proto.UnimplementedAuthServiceServer
	service *service.AuthService
}

func NewAuthHandler(s *service.AuthService) *AuthHandler {
	return &AuthHandler{service: s}
}

func CreateUser(ctx context.Context, req *proto.CreateUserRequest) (res *proto.CreateUserResponse, error) {	
}








