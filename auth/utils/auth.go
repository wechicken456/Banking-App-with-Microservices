package utils

import (
	"auth/model"
	"auth/proto"
)

func ConvertUserToProfile(user *model.User) *model.UserProfile {
	return &model.UserProfile{
		Email: user.Email,
	}
}

func ConvertProfileToProtoProfile(profile *model.UserProfile) *proto.UserProfile {
	return &proto.UserProfile{
		Email: profile.Email,
	}
}
