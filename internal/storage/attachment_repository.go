package storage

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Attachment struct {
	ID       int
	EmailID  int // FK para emails.id (INTEGER, não string)
	Filename string
	MimeType string
	Path     string
}

type AttachmentRepository struct {
	conn *pgx.Conn
}

func NewAttachmentRepository(conn *pgx.Conn) *AttachmentRepository {
	return &AttachmentRepository{conn: conn}
}

func (r *AttachmentRepository) Save(ctx context.Context, a Attachment) error {
	_, err := r.conn.Exec(ctx, `
		INSERT INTO attachments (email_id, filename, mime_type, file_path)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (email_id, filename) DO NOTHING
	`,
		a.EmailID,
		a.Filename,
		a.MimeType,
		a.Path,
	)
	return err
}
