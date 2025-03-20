package repos

import (
	"context"
	"cudo-test/gen"
)

type MainRepo interface {
	GetRecentTransactions(ctx context.Context, limit int32) ([]gen.Transaction, error)
}

type MainRepoImpl struct {
	repo *gen.Queries
}

func InitMainRepo(repo *gen.Queries) *MainRepoImpl {
	return &MainRepoImpl{
		repo: repo,
	}
}

func (r *MainRepoImpl) GetRecentTransactions(ctx context.Context, limit int32) ([]gen.Transaction, error) {
	return r.repo.GetRecentTransactions(ctx, limit)
}
