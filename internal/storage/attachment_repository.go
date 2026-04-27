package storage

import (
	"context"
	"encoding/csv"
	storage "fineasy/internal"
	"fmt"
	"os"
	"strconv"
	"sync"
)

type Attachment struct {
	ID       int
	EmailID  int64 // FK para o id do e-mail no CSV
	Filename string
	MimeType string
	Path     string
}

// AttachmentRepository persiste attachments em CSV.
// Deduplicação por (email_id + filename) em memória.
type AttachmentRepository struct {
	mu       sync.Mutex
	filePath string
	seen     map[string]struct{} // chave: "emailID|filename"
	nextID   int64
}

func NewAttachmentRepository(filePath string) (*AttachmentRepository, error) {
	r := &AttachmentRepository{
		filePath: filePath,
		seen:     make(map[string]struct{}),
	}
	if err := r.load(); err != nil {
		return nil, fmt.Errorf("attachment repository: %w", err)
	}
	return r, nil
}

func (r *AttachmentRepository) load() error {
	if _, err := os.Stat(r.filePath); os.IsNotExist(err) {
		return r.writeHeader()
	}

	f, err := os.Open(r.filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return err
	}

	var maxID int64
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 3 {
			continue
		}
		key := row[1] + "|" + row[2] // email_id | filename
		r.seen[key] = struct{}{}
		if id, err := strconv.ParseInt(row[0], 10, 64); err == nil && id > maxID {
			maxID = id
		}
	}
	r.nextID = maxID + 1
	return nil
}

func (r *AttachmentRepository) writeHeader() error {
	if err := os.MkdirAll(storage.DirOf(r.filePath), os.ModePerm); err != nil {
		return err
	}
	f, err := os.Create(r.filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.Write([]string{"id", "email_id", "filename", "mime_type", "path"}); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}

// Save insere o attachment no CSV, ignorando duplicatas (email_id + filename).
func (r *AttachmentRepository) Save(_ context.Context, a Attachment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := strconv.FormatInt(a.EmailID, 10) + "|" + a.Filename
	if _, exists := r.seen[key]; exists {
		return nil
	}

	f, err := os.OpenFile(r.filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	record := []string{
		strconv.FormatInt(r.nextID, 10),
		strconv.FormatInt(a.EmailID, 10),
		a.Filename,
		a.MimeType,
		a.Path,
	}
	if err := w.Write(record); err != nil {
		return err
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}

	r.seen[key] = struct{}{}
	r.nextID++
	return nil
}
