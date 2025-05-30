package handler

import (
	"auth/model"
	"auth/proto"
	"auth/service"
	"context"

	"github.com/google/uuid"
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
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return &proto.DeleteUserResponse{}, err
	}
	err = h.service.DeleteUser(ctx, userID)
	return &proto.DeleteUserResponse{}, err
}

func (h *AuthHandler) LoginUser(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	user := &model.User{
		Email:    req.Email,
		Password: req.Password,
	}
	res, err := h.service.LoginUser(ctx, user, req.IdempotencyKey)
	if err != nil {
		return nil, err
	}
	return &proto.LoginResponse{
		UserId:       res.UserID.String(),
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshTokenAsCookie,
		Fingerprint:  res.FingerprintAsCookie,
	}, nil
}

func (h *AuthHandler) RenewAccessToken(ctx context.Context, req *proto.RenewAccessTokenRequest) (*proto.RenewAccessTokenResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, err
	}

	res, err := h.service.RenewAccessToken(ctx, userID, req.RefreshToken, req.IdempotencyKey)
	if err != nil {
		return nil, err
	}

	return &proto.RenewAccessTokenResponse{
		AccessToken: res.TokenString,
		Fingerprint: res.FingerprintCookieString,
	}, nil
}
