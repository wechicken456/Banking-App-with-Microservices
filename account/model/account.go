package model

type User struct {
	UserID  string `json:"user_id"`
	Balance int64  `json:"balance"`
}

type Account struct {
	AccountID     string `json:"account_id"`
	UserID        string `json:"user_id"`
	Balance       int64  `json:"balance"`
	AccountNumber string `json:"account_number"`
}
