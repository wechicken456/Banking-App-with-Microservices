package service

import (
	"auth/model"
	"auth/repository"
	"auth/utils"
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type AuthService struct {
	repo *repository.AuthRepository
	db   *sqlx.DB
}

var (
	JwtSecretKey string
	JwtPublicKey string
)

// r and db should be created in the main function and passed to the service
// sqlx.DB object maintains a connection pool internally, and will attempt to connect when a connection is first needed.
func NewAuthService(repo *repository.AuthRepository, db *sqlx.DB) *AuthService {
	return &AuthService{
		repo: repo,
		db:   db,
	}
}

func (s *AuthService) CreateUser(ctx context.Context, user *model.User, idempotencyKey string) (*model.User, error) {
	// For user creation, we might have different authorization rules
	// For example, only admin users might be allowed to create other users
	// This would require additional role-based checks

	user.UserID = uuid.New()
	passwordHash, err := utils.HashPassword(user.Password)
	if err != nil {
		return nil, model.ErrInternalServer
	}
	user.Password = passwordHash

	createdUser, err := s.createUserTx(ctx, user, idempotencyKey)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == pgerrcode.UniqueViolation {
			return nil, model.ErrUserAlreadyExists
		}
		log.Printf("CreateUser: %v", err)
		return nil, model.ErrInternalServer
	}
	return createdUser, nil
}

func (s *AuthService) createUserTx(ctx context.Context, user *model.User, idempotencyKey string) (*model.User, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("createUserTx: failed to beign transaction: %v", err)
		return nil, model.ErrInternalServer
	}
	defer tx.Rollback()

	txRepo := s.repo.WithTx(tx)

	// check if user already exists
	user, err = txRepo.GetUserByEmail(ctx, user.Email)
	if err == nil {
		return user, errors.New("user already exists")
	}

	// check if this is a duplicate request
	// Try to insert idempotency key with status "PENDING".
	// the statement will block if another concurrent transactional already to inserts the same key, even if it hasn't committed yet.
	key, err := txRepo.GetOrClaimIdempotencyKey(ctx, &model.IdempotencyKey{
		KeyID:  idempotencyKey,
		Status: "PENDING",
	})
	if err == nil {
		if key.Status != "PENDING" { // "PENDING" implies that we (the current transaction) is the first one to create the idempotency key. Otherwise, we blocked while another transaction inserted the same key.
			log.Printf("createUserTx: idempotency key already exists: %v\n", key)
			cachedTransaction := &model.User{}
			err := json.Unmarshal([]byte(key.ResponseMessage), cachedTransaction)
			if err != nil {
				log.Printf("createTransactionTx: Failed to unmarshal transaction: %v\n", err)
				return nil, model.ErrInternalServer
			}
			return cachedTransaction, nil
		}
	} else {
		log.Printf("createTransactionTx: Failed to get idempotency key: %v\n", err)
		return nil, model.ErrInternalServer
	}

	passwordHash, err := utils.HashPassword(user.Password)
	if err != nil {
		return nil, model.ErrInternalServer
	}
	user.Password = passwordHash
	user, err = txRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, model.ErrInternalServer
	}

	// Update the idempotency key status
	key.Status = "COMPLETED"
	marshalled, err := json.Marshal(user)
	if err != nil {
		log.Printf("createAccountTx: Failed to marshal transaction: %v\n", err)
		return nil, model.ErrInternalServer
	}
	key.ResponseMessage = string(marshalled)

	if _, err = txRepo.UpdateIdempotencyKey(ctx, key); err != nil {
		log.Printf("createAccountTx: Failed to create idempotency key: %v\n", err)
		return nil, model.ErrInternalServer
	}

	if err = tx.Commit(); err != nil {
		log.Printf("createAccountTx: Failed to commit transaction: %v\n", err)
		return nil, model.ErrInternalServer
	}

	return user, nil
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
		log.Printf("UpdateUserPassword: Failed to update password for user %v: %v\n", user.UserID, err)
		return nil, model.ErrInternalServer
	}
	return updatedUser, nil
}

func (s *AuthService) DeleteUser(ctx context.Context, userID uuid.UUID, requestingUserID uuid.UUID) error {
	// Check ownership - only allow users to delete their own account
	if userID != requestingUserID {
		log.Printf("DeleteUser: Unauthorized delete attempt for user %v by user %v\n",
			userID, requestingUserID)
		return model.ErrInternalServer
	}

	err := s.repo.DeleteUser(ctx, userID)
	if err != nil {
		log.Printf("DeleteUser: Failed to delete user %v: %v\n", userID, err)
		return model.ErrInternalServer
	}
	return nil
}

func (s *AuthService) LoginUser(ctx context.Context, user *model.User, idempotencyKey string) (*model.LoginResult, error) {
	hashedPwd, err := utils.HashPassword(user.Password)
	if err != nil {
		log.Printf("LoginUser: failed to hash password: %v", err)
		return nil, model.ErrInternalServer
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("createUserTx: failed to beign transaction: %v", err)
		return nil, model.ErrInternalServer
	}
	defer tx.Rollback()

	txRepo := s.repo.WithTx(tx)

	user, err = txRepo.GetUserByEmail(ctx, user.Email)
	if err != nil {
		log.Printf("LoginUser: failed to get user: %v", err)
		return nil, model.ErrInternalServer
	}

	// user variable has been overwritten by the value fetched from db, so the field Password should contain the stored hash
	if hashedPwd != user.Password {
		log.Printf("LoginUser: password hash mismatch for user %v", user.Email)
		return nil, model.ErrInternalServer
	}

	accessToken, err := utils.RandomAccessToken(user.UserID, JwtSecretKey)
	if err != nil {
		log.Printf("LoginUser: %v", err)
		return nil, model.ErrInternalServer
	}

	// check if this is a duplicate request. If so, we shouldn't genereate another refresh token
	// Try to insert idempotency key with status "PENDING".
	// the statement will block if another concurrent transactional already to inserts the same key, even if it hasn't committed yet.
	key, err := txRepo.GetOrClaimIdempotencyKey(ctx, &model.IdempotencyKey{
		KeyID:  idempotencyKey,
		Status: "PENDING",
	})
	if err == nil {
		if key.Status != "PENDING" { // "PENDING" implies that we (the current transaction) is the first one to create the idempotency key. Otherwise, we blocked while another transaction inserted the same key.
			log.Printf("createUserTx: idempotency key already exists: %v\n", key)
			cachedTransaction := &model.LoginResult{}
			err := json.Unmarshal([]byte(key.ResponseMessage), cachedTransaction)
			if err != nil {
				log.Printf("createTransactionTx: Failed to unmarshal transaction: %v\n", err)
				return nil, model.ErrInternalServer
			}
			return cachedTransaction, nil
		}
	} else {
		log.Printf("createTransactionTx: Failed to get idempotency key: %v\n", err)
		return nil, model.ErrInternalServer
	}

	refreshToken, err := utils.RandomRefreshToken()
	if err != nil {
		log.Printf("LoginUser: %v", err)
		return nil, model.ErrInternalServer
	}

	// Store refresh token in db
	_, err = s.repo.CreateRefreshToken(ctx, &model.RefreshTokenRepo{
		UserID:    user.UserID,
		TokenHash: refreshToken.TokenHash,
		ExpiredAt: time.Now().Add(model.RefreshTokenDuration),
	})
	if err != nil {
		log.Printf("LoginUser: failed to store refresh token in db: %v", err)
		return nil, model.ErrInternalServer
	}

	ret := &model.LoginResult{
		AccessToken:          accessToken.TokenString,
		UserID:               user.UserID.String(),
		FingerprintAsCookie:  accessToken.FingerprintCookieString,
		RefreshTokenAsCookie: refreshToken.TokenAsCookie,
	}

	// Update the idempotency key status
	key.Status = "COMPLETED"
	marshalled, err := json.Marshal(ret)
	if err != nil {
		log.Printf("createAccountTx: Failed to marshal transaction: %v\n", err)
		return nil, model.ErrInternalServer
	}
	key.ResponseMessage = string(marshalled)

	if _, err = txRepo.UpdateIdempotencyKey(ctx, key); err != nil {
		log.Printf("createAccountTx: Failed to create idempotency key: %v\n", err)
		return nil, model.ErrInternalServer
	}

	if err = tx.Commit(); err != nil {
		log.Printf("createAccountTx: Failed to commit transaction: %v\n", err)
		return nil, model.ErrInternalServer
	}

	return ret, nil
}

func (s *AuthService) RenewAccessToken(ctx context.Context)
