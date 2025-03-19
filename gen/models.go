// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package gen

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Transaction struct {
	ID              int64
	UserID          int64
	OrderID         string
	TransactionDate pgtype.Timestamp
	Amount          pgtype.Numeric
	PaymentMethod   string
	Status          string
	CreatedAt       pgtype.Timestamp
	UpdatedAt       pgtype.Timestamp
}
