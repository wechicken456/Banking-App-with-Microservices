package model

type CreateAccountRequest struct {
	Balance int64 `json:"balance"`
}

type CreateAccountResponse struct {
	AccountID     string `json:"accountId"`
	AccountNumber int32  `json:"accountNumber"`
}

type Account struct {
	AccountID     string `json:"accountId"`
	AccountNumber int32  `json:"accountNumber"`
	Balance       int64  `json:"balance"`
	UserID        string `json:"userId"`
}

type UserProfile struct {
	Email string `json:"email"`
}

type GetAccountsByUserIDResponse struct {
	Accounts []Account `json:"accounts"`
}

type GetAccountResponse struct {
	Account Account `json:"account"`
}

type DeleteAccountByAccountNumberRequest struct {
	AccountNumber int32 `json:"accountNumber"`
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

// the TransactionType can only be "CREDIT" or "DEBIT".
// Only the transfer service can create "TRANSFER_CREDIT" or "TRANSFER_DEBIT" transactions, and it will use gRPC to call the account service directly.
// Hence, the API Gateway does NOT directly handle the transfer requests.
type CreateTransactionRequest struct {
	AccountID       string `json:"accountId"`
	Amount          int64  `json:"amount"`
	TransactionType string `json:"transactionType"`
}

type CreateTransactionResponse struct {
	TransactionID string `json:"transactionId"`
}

type GetTransactionsByAccountIdResponse struct {
	Transactions []Transaction `json:"transactions"`
}
