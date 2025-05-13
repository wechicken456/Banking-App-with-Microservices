package utils

import (
	"math/rand"
	"transfer/model"

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

func RandomTransfer() *model.Transfer {
	return &model.Transfer{
		FromAccountID:  uuid.New(),
		ToAccountID:    uuid.New(),
		IdempotencyKey: RandomString(10),
		Amount:         int64(RandMinMax(1, 100_000_000)),
	}
}
