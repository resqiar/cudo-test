package repos

import (
	"context"
	"cudo-test/gen"
)

type MainRepo interface {
	GetUserTransactionWithinTimeframe(ctx context.Context, userID int64) ([]gen.GetUserTransactionsWithinTimeframeRow, error)
	GetUserTransactions(ctx context.Context, userID int64) ([]gen.Transaction, error)
}

type MainRepoImpl struct {
	repo *gen.Queries
}

func InitMainRepo(repo *gen.Queries) *MainRepoImpl {
	return &MainRepoImpl{
		repo: repo,
	}
}

func (r *MainRepoImpl) GetUserTransactions(ctx context.Context, userID int64) ([]gen.Transaction, error) {
	return r.repo.GetUserTransactions(ctx, userID)
}

func (r *MainRepoImpl) GetUserTransactionWithinTimeframe(ctx context.Context, userID int64) ([]gen.GetUserTransactionsWithinTimeframeRow, error) {
	return r.repo.GetUserTransactionsWithinTimeframe(ctx, userID)
}
