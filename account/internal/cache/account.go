package cache

import (
	"account/internal/redis"
	"account/model"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	AccountKeyPrefix = "acct"
	TTL              = 2 * time.Second
)

func AccountKey(accountID uuid.UUID) string {
	return fmt.Sprintf("%s:{%s}", AccountKeyPrefix, accountID.String())
}

func Get(ctx context.Context, accountID uuid.UUID) (*model.Account, error) {
	key := AccountKey(accountID)
	data, err := redis.Client.Get(ctx, key).Bytes()
	if err == model.ErrCacheMiss {
		return nil, err
	} else if err != nil {
		return nil, err
	}

	var acct model.Account
	if err = json.Unmarshal(data, &acct); err != nil {
		return nil, err
	}
	return &acct, nil
}

func Set(ctx context.Context, account *model.Account) error {
	key := AccountKey(account.AccountID)
	data, _ := json.Marshal(account)
	return redis.Client.Set(ctx, key, data, TTL).Err()
}

func Invalidate(ctx context.Context, accountID uuid.UUID) error {
	key := AccountKey(accountID)
	return redis.Client.Del(ctx, key).Err()
}
