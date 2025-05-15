package repository

import (
	"context"
	"database/sql"
	"transfer/db/sqlc"
	"transfer/model"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// TransferRepository handles database operations for transfers.
type TransferRepository struct {
	queries *sqlc.Queries
	db      *sqlx.DB
}

func NewTransferRepository(db *sqlx.DB) *TransferRepository {
	return &TransferRepository{queries: sqlc.New(db), db: db}
}

// WithTx returns a new TransferRepository that uses the provided transaction.
func (r *TransferRepository) WithTx(tx *sql.Tx) *TransferRepository {
	return &TransferRepository{
		queries: r.queries.WithTx(tx),
		db:      r.db,
	}
}

func convertToModelTransfer(transfer sqlc.Transfer) *model.Transfer {
	return &model.Transfer{
		TransferID:     transfer.ID,
		FromAccountID:  transfer.FromAccountID,
		ToAccountID:    transfer.ToAccountID,
		IdempotencyKey: transfer.IdempotencyKey,
		Amount:         transfer.Amount,
		Status:         transfer.Status,
	}
}

// converts a model.Transfer to sqlc.CreateTransferParams.
func convertToCreateTransferParams(transfer *model.Transfer) sqlc.CreateTransferParams {
	return sqlc.CreateTransferParams{
		ID:             transfer.TransferID,
		FromAccountID:  transfer.FromAccountID,
		ToAccountID:    transfer.ToAccountID,
		IdempotencyKey: transfer.IdempotencyKey,
		Amount:         transfer.Amount,
		Status:         transfer.Status,
	}
}

func (r *TransferRepository) CreateTransfer(ctx context.Context, transfer *model.Transfer) (*model.Transfer, error) {
	params := convertToCreateTransferParams(transfer)
	if params.ID == uuid.Nil {
		params.ID = uuid.New()
	}
	createdTransfer, err := r.queries.CreateTransfer(ctx, params)
	if err != nil {
		return nil, err
	}
	return convertToModelTransfer(createdTransfer), nil
}

func (r *TransferRepository) GetTransferByID(ctx context.Context, id uuid.UUID) (*model.Transfer, error) {
	transfer, err := r.queries.GetTransferByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return convertToModelTransfer(transfer), nil
}

// retrieves all transfers originating from a specific account ID.
func (r *TransferRepository) GetTransfersByFromID(ctx context.Context, fromAccountID uuid.UUID) ([]model.Transfer, error) {
	transfers, err := r.queries.GetTransfersByFromID(ctx, fromAccountID)
	if err != nil {
		return nil, err
	}
	modelTransfers := make([]model.Transfer, len(transfers))
	for i, transfer := range transfers {
		modelTransfers[i] = *convertToModelTransfer(transfer)
	}
	return modelTransfers, nil
}

// retrieves all transfers destined for a specific account ID.
func (r *TransferRepository) GetTransfersByToID(ctx context.Context, toAccountID uuid.UUID) ([]model.Transfer, error) {
	transfers, err := r.queries.GetTransfersByToID(ctx, toAccountID)
	if err != nil {
		return nil, err
	}
	modelTransfers := make([]model.Transfer, len(transfers))
	for i, transfer := range transfers {
		modelTransfers[i] = *convertToModelTransfer(transfer)
	}
	return modelTransfers, nil
}
