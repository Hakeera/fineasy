package storage

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func NewConnection(ctx context.Context) (*pgx.Conn, error) {
	return pgx.Connect(ctx, "postgres://fineasy:40028922@localhost:5432/fineasy?sslmode=disable")
}
