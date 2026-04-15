package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// Formatos de data usados pelos servidores de e-mail
var dateFormats = []string{
	time.RFC1123Z,                // "Mon, 02 Jan 2006 15:04:05 -0700"
	time.RFC1123,                 // "Mon, 02 Jan 2006 15:04:05 MST"
	"02 Jan 2006 15:04:05 -0700", // sem dia da semana
	"02 Jan 2006 15:04:05 MST",
}

func parseEmailDate(raw string) (time.Time, error) {
	for _, format := range dateFormats {
		if t, err := time.Parse(format, raw); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("formato de data não reconhecido: %q", raw)
}

type Email struct {
	ID         int
	GmailID    string
	Subject    string
	From       string
	ReceivedAt string // string crua vinda do header do Gmail
}

type EmailRepository struct {
	conn *pgx.Conn
}

func NewEmailRepository(conn *pgx.Conn) *EmailRepository {
	return &EmailRepository{conn: conn}
}

// Save insere o e-mail e retorna (true, id) se foi inserido, (false, 0) se já existia.
func (r *EmailRepository) Save(ctx context.Context, e Email) (bool, int, error) {
	receivedAt, err := parseEmailDate(e.ReceivedAt)
	if err != nil {
		return false, 0, fmt.Errorf("received_at inválido: %w", err)
	}

	var id int
	err = r.conn.QueryRow(ctx, `
		INSERT INTO emails (gmail_id, subject, from_email, received_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (gmail_id) DO NOTHING
		RETURNING id
	`,
		e.GmailID,
		e.Subject,
		e.From,
		receivedAt,
	).Scan(&id)

	if err != nil {
		// RETURNING não retorna nada em caso de DO NOTHING — não é erro real
		if err.Error() == "no rows in result set" {
			return false, 0, nil
		}
		return false, 0, err
	}

	return true, id, nil
}
