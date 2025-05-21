package service

import (
	"auth/model"
	"auth/repository"
	"auth/utils"
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type AuthService struct {
	repo *repository.AuthRepository
	db   *sqlx.DB
}

// r and db should be created in the main function and passed to the service
// sqlx.DB object maintains a connection pool internally, and will attempt to connect when a connection is first needed.
func NewAuthService(repo *repository.AuthRepository, db *sqlx.DB) *AuthService {
	return &AuthService{
		repo: repo,
		db:   db,
	}
}

func (s *AuthService) CreateUser(ctx context.Context, user *model.User, requestingUserID uuid.UUID) (*model.User, error) {
	// For user creation, we might have different authorization rules
	// For example, only admin users might be allowed to create other users
	// This would require additional role-based checks

	user.UserID = uuid.New()
	passwordHash, err := utils.HashPassword(user.Password)
	if err != nil {
		return nil, model.ErrInternalServer
	}
	user.Password = passwordHash

	createdUser, err := s.repo.CreateUserTx(ctx, user)
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == pgerrcode.UniqueViolation {
		return nil, model.ErrUserAlreadyExists
	}

	return createdUser, nil
}

func (s *AuthService) GetUserByEmail(ctx context.Context, email string, requestingUserID uuid.UUID) (*model.User, error) {
	// First get the user to check if the requesting user is authorized
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, model.ErrInternalServer
	}

	// Check ownership - only allow users to access their own information
	if user.UserID != requestingUserID {
		log.Printf("Unauthorized access attempt for user email %v by user %v\n",
			email, requestingUserID)
		return nil, model.ErrInternalServer
	}

	return user, nil
}

func (s *AuthService) UpdateUser(ctx context.Context, user *model.User, requestingUserID uuid.UUID) (*model.User, error) {
	// Check ownership - only allow users to update their own information
	if user.UserID != requestingUserID {
		log.Printf("Unauthorized update attempt for user %v by user %v\n",
			user.UserID, requestingUserID)
		return nil, model.ErrInternalServer
	}

	updatedUser, err := s.repo.UpdateUser(ctx, user)
	if err != nil {
		log.Printf("Failed to update user %v: %v\n", user.UserID, err)
		return nil, model.ErrInternalServer
	}
	return updatedUser, nil
}

func (s *AuthService) UpdateUserPassword(ctx context.Context, user *model.User, requestingUserID uuid.UUID, newPassword string) (*model.User, error) {
	// Check ownership - only allow users to update their own password
	if user.UserID != requestingUserID {
		log.Printf("Unauthorized password update attempt for user %v by user %v\n",
			user.UserID, requestingUserID)
		return nil, model.ErrInternalServer
	}

	passwordHash, err := utils.HashPassword(newPassword)
	if err != nil {
		log.Printf("UpdateUserPassword: Failed to hash password for user %v: %v\n", user.UserID, err)
		return nil, model.ErrInternalServer
	}
	user.Password = passwordHash

	updatedUser, err := s.UpdateUser(ctx, user, requestingUserID)
	if err != nil {
		log.Printf("Failed to update password for user %v: %v\n", user.UserID, err)
		return nil, model.ErrInternalServer
	}
	return updatedUser, nil
}

func (s *AuthService) DeleteUser(ctx context.Context, userID uuid.UUID, requestingUserID uuid.UUID) error {
	// Check ownership - only allow users to delete their own account
	if userID != requestingUserID {
		log.Printf("Unauthorized delete attempt for user %v by user %v\n",
			userID, requestingUserID)
		return model.ErrInternalServer
	}

	err := s.repo.DeleteUser(ctx, userID)
	if err != nil {
		log.Printf("Failed to delete user %v: %v\n", userID, err)
		return model.ErrInternalServer
	}
	return nil
}
