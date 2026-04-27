package storage

import (
	"context"
	"encoding/csv"
	storage "fineasy/internal"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
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

// EmailRepository persiste e-mails em CSV.
// Os gmail_id já processados são carregados em memória na inicialização,
// evitando leituras repetidas do disco a cada Save.
type EmailRepository struct {
	mu       sync.Mutex
	filePath string
	seen     map[string]struct{} // gmail_id já existentes
	nextID   int64
}

func NewEmailRepository(filePath string) (*EmailRepository, error) {
	r := &EmailRepository{
		filePath: filePath,
		seen:     make(map[string]struct{}),
	}
	if err := r.load(); err != nil {
		return nil, fmt.Errorf("email repository: %w", err)
	}
	return r, nil
}

// load cria o arquivo (com cabeçalho) se não existir,
// ou lê os registros existentes para popular o cache em memória.
func (r *EmailRepository) load() error {
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
			continue // cabeçalho
		}
		if len(row) < 2 {
			continue
		}
		r.seen[row[1]] = struct{}{} // coluna gmail_id
		if id, err := strconv.ParseInt(row[0], 10, 64); err == nil && id > maxID {
			maxID = id
		}
	}
	r.nextID = maxID + 1
	return nil
}

func (r *EmailRepository) writeHeader() error {
	if err := os.MkdirAll(storage.DirOf(r.filePath), os.ModePerm); err != nil {
		return err
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

// Save insere o e-mail no CSV e retorna (true, id) se inserido,
// (false, 0, nil) se o gmail_id já existia.
func (r *EmailRepository) Save(_ context.Context, e Email) (bool, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.seen[e.GmailID]; exists {
		return false, 0, nil
	}

	// Normaliza a data para exibição legível no CSV/Excel
	receivedAt := e.ReceivedAt
	if t, err := parseEmailDate(e.ReceivedAt); err == nil {
		receivedAt = t.Format("2006-01-02 15:04:05")
	}

	f, err := os.OpenFile(r.filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return false, 0, err
	}
	defer f.Close()

	id := r.nextID
	w := csv.NewWriter(f)
	record := []string{
		strconv.FormatInt(id, 10),
		e.GmailID,
		e.Subject,
		e.From,
		receivedAt,
	}
	if err := w.Write(record); err != nil {
		return false, 0, err
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return false, 0, err
	}

	r.seen[e.GmailID] = struct{}{}
	r.nextID++
	return true, id, nil
}
