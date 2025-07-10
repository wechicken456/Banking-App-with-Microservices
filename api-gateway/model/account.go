package model

type CreateAccountRequest struct {
	Balance int64 `json:"balance"`
}

type CreateAccountResponse struct {
	AccountID string `json:"accountId"`
}

type Account struct {
	AccountID     string `json:"accountId"`
	AccountNumber int64  `json:"accountNumber"`
	Balance       int64  `json:"balance"`
	UserID        string `json:"userId"`
}

type GetAccountsByUserIDResponse struct {
	Accounts []Account `json:"accounts"`
}

type GetAccountResponse struct {
	Account Account `json:"account"`
}

type DeleteAccountByAccountNumberRequest struct {
	AccountNumber int64 `json:"accountNumber"`
}

type Transaction struct {
	TransactionID   string `json:"transactionId"`
	AccountID       string `json:"accountId"`
	Amount          int64  `json:"amount"`
	Timestamp       int64  `json:"timestamp"`
	TransactionType string `json:"transactionType"`
	Status          string `json:"status"`
	TransferID      string `json:"transferId,omitempty"`
}

type CreateTransactionRequest struct {
	AccountID       string `json:"accountId"`
	Amount          int64  `json:"amount"`
	TransactionType string `json:"transactionType"`
	TransferID      string `json:"transferId,omitempty"`
	Status          string `json:"status"`
}

type CreateTransactionResponse struct {
	TransactionID string `json:"transactionId"`
}
