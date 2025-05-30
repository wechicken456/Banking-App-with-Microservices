// Service layer for the auth microservice. The auth microservice assuems that teh API Gateway has
// already verified the client request JWT token. The verified JWT contains the user id of the request,
// which the API Gateway will pass as gRPC message argument to the auth microservice. Hence, every single
// request can only affect the user whose JWT was validated. So there's no need
// to re-authenticate the request at this microservice, with the exception of the RenewAccessToken request.
// Note that model.User will contain the UserID of a user.
//
// The RenewAccessToken request should compare the userID from the validated JWT against the userID of the
// refresh_token. This ensures that a malicious renew request that carries a stolen JWT won't match the
// attacker's userID in their refresh_token.
package service

import (
	"auth/model"
	"auth/repository"
	"auth/utils"
	"context"
	"database/sql"
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

func (s *AuthService) UpdateUser(ctx context.Context, user *model.User) (*model.User, error) {
	updatedUser, err := s.repo.UpdateUser(ctx, user)
	if err != nil {
		log.Printf("Failed to update user %v: %v\n", user.UserID, err)
		return nil, model.ErrInternalServer
	}
	return updatedUser, nil
}

func (s *AuthService) UpdateUserPassword(ctx context.Context, user *model.User, newPassword string) (*model.User, error) {
	passwordHash, err := utils.HashPassword(newPassword)
	if err != nil {
		log.Printf("UpdateUserPassword: Failed to hash password for user %v: %v\n", user.UserID, err)
		return nil, model.ErrInternalServer
	}
	user.Password = passwordHash

	updatedUser, err := s.UpdateUser(ctx, user)
	if err != nil {
		log.Printf("UpdateUserPassword: Failed to update password for user %v: %v\n", user.UserID, err)
		return nil, model.ErrInternalServer
	}
	return updatedUser, nil
}

func (s *AuthService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
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
		log.Printf("LoginUser: failed to beign transaction: %v", err)
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
			log.Printf("LoginUser: idempotency key already exists: %v\n", key)
			cachedTransaction := &model.LoginResult{}
			err := json.Unmarshal([]byte(key.ResponseMessage), cachedTransaction)
			if err != nil {
				log.Printf("LoginUser: Failed to unmarshal transaction: %v\n", err)
				return nil, model.ErrInternalServer
			}
			return cachedTransaction, nil
		}
	} else {
		log.Printf("LoginUser: Failed to get idempotency key: %v\n", err)
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
		UserID:               user.UserID,
		FingerprintAsCookie:  accessToken.FingerprintCookieString,
		RefreshTokenAsCookie: refreshToken.TokenAsCookie,
	}

	// Update the idempotency key status
	key.Status = "COMPLETED"
	marshalled, err := json.Marshal(ret)
	if err != nil {
		log.Printf("LoginUser: Failed to marshal transaction: %v\n", err)
		return nil, model.ErrInternalServer
	}
	key.ResponseMessage = string(marshalled)

	if _, err = txRepo.UpdateIdempotencyKey(ctx, key); err != nil {
		log.Printf("LoginUser: Failed to create idempotency key: %v\n", err)
		return nil, model.ErrInternalServer
	}

	if err = tx.Commit(); err != nil {
		log.Printf("LoginUser: Failed to commit transaction: %v\n", err)
		return nil, model.ErrInternalServer
	}

	return ret, nil
}

func (s *AuthService) RenewAccessToken(ctx context.Context, userID uuid.UUID, refresh_token string, idempotencyKey string) (*model.AccessToken, error) {
	// get the userID of the refresh_token
	token, err := s.repo.GetRefreshToken(ctx, refresh_token)
	if err != nil {
		log.Printf("RenewAccessToken: %v", err)
		if err == sql.ErrNoRows {
			return nil, model.ErrInvalidJWT // TODO: change error type
		}
		return nil, model.ErrInternalServer
	}

	// make sure the user exists
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		log.Printf("RenewAccessToken: failed to get user: %v", err)
		return nil, model.ErrInternalServer
	}

	// check userID of refresh_token is the same as the requesting userID
	if user.UserID != token.UserID {
		log.Printf("RenewAccessToken: Unauthorized attempt to renew token for user %v from user %v", userID, token.UserID)
		return nil, model.ErrNotAuthorized
	}

	// check refresh_token expiration time
	if time.Now().After(token.ExpiredAt) {
		return nil, model.ErrNotAuthenticated
	}

	// begin a transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("RenewAccessToken: failed to beign transaction: %v", err)
		return nil, model.ErrInternalServer
	}
	defer tx.Rollback()

	txRepo := s.repo.WithTx(tx)

	// check if this is a duplicate request. If so, we shouldn't genereate another refresh token
	// Try to insert idempotency key with status "PENDING".
	// the statement will block if another concurrent transactional already to inserts the same key, even if it hasn't committed yet.
	key, err := txRepo.GetOrClaimIdempotencyKey(ctx, &model.IdempotencyKey{
		KeyID:  idempotencyKey,
		Status: "PENDING",
	})
	if err == nil {
		if key.Status != "PENDING" { // "PENDING" implies that we (the current transaction) is the first one to create the idempotency key. Otherwise, we blocked while another transaction inserted the same key.
			log.Printf("RenewAccessToken: idempotency key already exists: %v\n", key)
			cachedTransaction := &model.AccessToken{}
			err := json.Unmarshal([]byte(key.ResponseMessage), cachedTransaction)
			if err != nil {
				log.Printf("RenewAccessToken: Failed to unmarshal transaction: %v\n", err)
				return nil, model.ErrInternalServer
			}
			return cachedTransaction, nil
		}
	} else {
		log.Printf("RenewAccessToken: Failed to get idempotency key: %v\n", err)
		return nil, model.ErrInternalServer
	}

	// generate a new access token
	accessToken, err := utils.RandomAccessToken(user.UserID, JwtSecretKey)
	if err != nil {
		log.Printf("RenewAccessToken: %v", err)
		return nil, model.ErrInternalServer
	}

	ret := &model.AccessToken{
		TokenString:             accessToken.TokenString,
		FingerprintCookieString: accessToken.FingerprintCookieString,
	}

	// Update the idempotency key status
	key.Status = "COMPLETED"
	marshalled, err := json.Marshal(ret)
	if err != nil {
		log.Printf("RenewAccessToken: Failed to marshal transaction: %v\n", err)
		return nil, model.ErrInternalServer
	}
	key.ResponseMessage = string(marshalled)

	if _, err = txRepo.UpdateIdempotencyKey(ctx, key); err != nil {
		log.Printf("RenewAccessToken: Failed to create idempotency key: %v\n", err)
		return nil, model.ErrInternalServer
	}

	if err = tx.Commit(); err != nil {
		log.Printf("RenewAccessToken: Failed to commit transaction: %v\n", err)
		return nil, model.ErrInternalServer
	}

	return ret, nil
}
