package repos

import (
	"context"
	"cudo-test/gen"
)

type MainRepo interface {
	GetAll(ctx context.Context) ([]gen.Transaction, error)
	GetUserTransactionWithinTimeframe(ctx context.Context) ([]gen.GetUserTransactionsWithinTimeframeRow, error)
}

type MainRepoImpl struct {
	repo *gen.Queries
}

func InitMainRepo(repo *gen.Queries) *MainRepoImpl {
	return &MainRepoImpl{
		repo: repo,
	}
}

func (r *MainRepoImpl) GetAll(ctx context.Context) ([]gen.Transaction, error) {
	return r.repo.GetTransactions(ctx)
}

func (r *MainRepoImpl) GetUserTransactionWithinTimeframe(ctx context.Context) ([]gen.GetUserTransactionsWithinTimeframeRow, error) {
	return r.repo.GetUserTransactionsWithinTimeframe(ctx)
}
