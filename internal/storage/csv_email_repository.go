package storage

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"sync"
)

// CSVEmailRepository persiste emails em um arquivo CSV.
type CSVEmailRepository struct {
	mu       sync.Mutex
	filePath string
}

func NewCSVEmailRepository(filePath string) (*CSVEmailRepository, error) {
	repo := &CSVEmailRepository{filePath: filePath}
	if err := repo.ensureHeader(); err != nil {
		return nil, fmt.Errorf("csv email repository: %w", err)
	}
	return repo, nil
}

// ensureHeader cria o arquivo com cabeçalho se ainda não existir.
func (r *CSVEmailRepository) ensureHeader() error {
	if _, err := os.Stat(r.filePath); err == nil {
		return nil // arquivo já existe
	}
	f, err := os.Create(r.filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.Write([]string{"id", "gmail_id", "subject", "from", "received_at"}); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}

// Save adiciona um email ao CSV se ele ainda não existir (mesmo comportamento
// do EmailRepository do banco).
// Retorna (inserted bool, id int64, err error).
func (r *CSVEmailRepository) Save(_ context.Context, email Email) (bool, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Verifica duplicata e descobre próximo ID.
	nextID, err := r.nextIDAndCheckDuplicate(email.GmailID)
	if err != nil {
		return false, 0, err
	}
	if nextID == -1 {
		// já existe
		return false, 0, nil
	}

	f, err := os.OpenFile(r.filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return false, 0, err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	record := []string{
		strconv.FormatInt(nextID, 10),
		email.GmailID,
		email.Subject,
		email.From,
		email.ReceivedAt,
	}
	if err := w.Write(record); err != nil {
		return false, 0, err
	}
	w.Flush()
	return true, nextID, w.Error()
}

// nextIDAndCheckDuplicate lê o arquivo e retorna o próximo ID disponível,
// ou -1 caso o gmailID já esteja registrado.
func (r *CSVEmailRepository) nextIDAndCheckDuplicate(gmailID string) (int64, error) {
	f, err := os.Open(r.filePath)
	if err != nil {
		return 1, nil // arquivo não existe ainda
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return 0, err
	}

	var maxID int64
	for i, row := range rows {
		if i == 0 {
			continue // cabeçalho
		}
		if len(row) > 1 && row[1] == gmailID {
			return -1, nil // duplicata
		}
		if len(row) > 0 {
			if id, err := strconv.ParseInt(row[0], 10, 64); err == nil && id > maxID {
				maxID = id
			}
		}
	}
	return maxID + 1, nil
}
