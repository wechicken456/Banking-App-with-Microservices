package utils

import (
	"account/model"
	"math/rand"

	"github.com/google/uuid"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-+"

// generates and return a random integer between min and max
func RandMinMax[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](min, max T) T {
	switch any(min).(type) {
	case int64:
		return T(min + T(rand.Int63n(int64(max-min+1))))
	case int, int8, int16, int32, uint, uint8, uint16, uint32, uint64:
		return T(min + T(rand.Intn(int(max-min+1))))
	default:
		// Fallback to float conversion
		return min + T(float64(max-min+1)*rand.Float64())
	}
}

func RandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// return a random 50 character email address
func RandomEmail() string {
	return RandomString(42) + "@" + RandomString(5) + "." + RandomString(3)
}

func RandomUser() *model.User {
	return &model.User{
		UserID:  uuid.New(),
		Balance: int64(RandMinMax(0, 100_000_000_000_000)),
	}
}

func RandomAccountNumber() int64 {
	return int64(RandMinMax(1_000_000_000, 1_000_000_000_000_000_000))
}

func RandomAccount() *model.Account {
	return &model.Account{
		AccountID:     uuid.New(),
		UserID:        uuid.New(),
		AccountNumber: RandomAccountNumber(),
		Balance:       int64(RandMinMax(1, 100)),
	}
}

func RandomTransactionType() string {
	types := []string{"CREDIT", "DEBIT", "TRANSFER_CREDIT", "TRANSFER_DEBIT"}
	return types[rand.Intn(len(types))]
}

func RandomTransactionStatus() string {
	statuses := []string{"PENDING", "COMPLETED", "FAILED"}
	return statuses[rand.Intn(len(statuses))]
}

func RandomTransferID() uuid.NullUUID {
	t := []uuid.NullUUID{
		{
			UUID:  uuid.New(),
			Valid: true,
		},
		{
			UUID:  uuid.Nil,
			Valid: false,
		},
	}
	return t[rand.Intn(len(t))]
}

func RandomTransaction() *model.Transaction {
	return &model.Transaction{
		TransactionID:   uuid.New(),
		AccountID:       uuid.New(),
		TransactionType: RandomTransactionType(),
		Status:          RandomTransactionStatus(),
		TransferID:      RandomTransferID(),
		Amount:          int64(RandMinMax(1, 100)),
	}
}

func RandomIdempotencyKey() uuid.UUID {
	return uuid.New()
}
