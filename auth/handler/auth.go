package handler

import (
	"auth/model"
	"auth/proto"
	"auth/service"
	"context"
)

type AuthHandler struct {
	proto.UnimplementedAuthServiceServer
	service *service.AuthService
}

func NewAuthHandler(s *service.AuthService) *AuthHandler {
	return &AuthHandler{service: s}
}

func (h *AuthHandler) CreateUser(ctx context.Context, req *proto.CreateUserRequest) (*proto.CreateUserResponse, error) {
	user := &model.User{
		Email:    req.Email,
		Password: req.Password,
	}
	user, err := h.service.CreateUser(ctx, user, req.IdempotencyKey)
	if err != nil {
		return &proto.CreateUserResponse{
			UserId: "",
		}, err
	}
	return &proto.CreateUserResponse{
		UserId: user.UserID.String(),
	}, nil
}

func (h *AuthHandler) DeleteUser(ctx context.Context, req *proto.DeleteUserRequest) (*proto.DeleteUserResponse, error) {
}
