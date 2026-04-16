package storage

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"sync"
)

// CSVAttachmentRepository persiste attachments em um arquivo CSV.
type CSVAttachmentRepository struct {
	mu       sync.Mutex
	filePath string
}

func NewCSVAttachmentRepository(filePath string) (*CSVAttachmentRepository, error) {
	repo := &CSVAttachmentRepository{filePath: filePath}
	if err := repo.ensureHeader(); err != nil {
		return nil, fmt.Errorf("csv attachment repository: %w", err)
	}
	return repo, nil
}

func (r *CSVAttachmentRepository) ensureHeader() error {
	if _, err := os.Stat(r.filePath); err == nil {
		return nil
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

// Save adiciona um attachment ao CSV.
func (r *CSVAttachmentRepository) Save(_ context.Context, attachment Attachment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	nextID, err := r.nextID()
	if err != nil {
		return err
	}

	f, err := os.OpenFile(r.filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	record := []string{
		strconv.FormatInt(nextID, 10),
		strconv.FormatInt(int64(attachment.EmailID), 10),
		attachment.Filename,
		attachment.MimeType,
		attachment.Path,
	}
	if err := w.Write(record); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}

func (r *CSVAttachmentRepository) nextID() (int64, error) {
	f, err := os.Open(r.filePath)
	if err != nil {
		return 1, nil
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return 0, err
	}

	var maxID int64
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) > 0 {
			if id, err := strconv.ParseInt(row[0], 10, 64); err == nil && id > maxID {
				maxID = id
			}
		}
	}
	return maxID + 1, nil
}
