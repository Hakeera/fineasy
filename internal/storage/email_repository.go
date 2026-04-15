package storage

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Email struct {
	ID         int
	GmailID    string
	Subject    string
	From       string
	ReceivedAt string
}

type EmailRepository struct {
	conn *pgx.Conn
}

func NewEmailRepository(conn *pgx.Conn) *EmailRepository {
	return &EmailRepository{conn: conn}
}

// insere tratando duplicados
func (r *EmailRepository) Save(ctx context.Context, e Email) (bool, error) {
	cmdTag, err := r.conn.Exec(ctx, `
		INSERT INTO emails (gmail_id, subject, from_email, received_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (gmail_id) DO NOTHING
	`,
		e.GmailID,
		e.Subject,
		e.From,
		e.ReceivedAt,
	)

	if err != nil {
		return false, err
	}

	return cmdTag.RowsAffected() > 0, nil
}
