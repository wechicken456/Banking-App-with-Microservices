package repository

import (
	"transfer/db/sqlc"
	"transfer/model"
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)


type TransferRepository struct {
	queries *sqlc.Queries
	db      *sqlx.DB
}

func NewTransferRepository(db *sqlx.DB) *TransferRepository {
	return &TransferRepository{queries: sqlc.New(db), db: db}
}

func (r *TransferRepository) CreateTransfer(ctx context.Context, transfer *model.Transfer) (*sqlc.Transfer, error) {
	createdtransfer, err := r.queries.CreateTransfer(ctx, sqlc.CreateTransferParams{
		ID:            uuid.New(),
		FromAccountID: transfer.FromAccountID,
		ToAccountID:   transfer.ToAccountID,
		IdempotencyKey: transfer.IdempotencyKey,
		Amount:        transfer.Amount,
		Status:        "PENDING",
	})
	if err != nil {
		return nil, err
	}
	return &createdtransfer, nil
}

func (r *TransferRepository) GetTransfersByFromID(ctx context.Context, fromAccountID uuid.UUID) ([]sqlc.Transfer, error) {
	transfer, err := r.queries.GetTransfersByFromID(ctx, fromAccountID)
	if err != nil {
		return nil, err
	}
	return transfer, nil
}

func (r *TransferRepository) GetTransfersByToID(ctx context.Context, toAccountID uuid.UUID) ([]sqlc.Transfer, error) {
	transfer, err := r.queries.GetTransfersByToID(ctx, toAccountID)
	if err != nil {
		return nil, err
	}
	return transfer, nil
}


func (r *TransferRepository) GetTransferByID(ctx context.Context, id uuid.UUID) (*sqlc.Transfer, error) {
	transfer, err := r.queries.GetTransferByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &transfer, nil
}

