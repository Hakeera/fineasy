package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"fineasy/internal/auth"
	gmailclient "fineasy/internal/gmail"
	"fineasy/internal/storage"

	"golang.org/x/oauth2/google"
	gmailapi "google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()

	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatal(err)
	}

	config, err := google.ConfigFromJSON(b, gmailapi.GmailReadonlyScope)
	if err != nil {
		log.Fatal(err)
	}

	// ── Banco de dados ────────────────────────────────────────────────────────
	conn, err := storage.NewConnection(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	emailRepo := storage.NewEmailRepository(conn)
	attachRepo := storage.NewAttachmentRepository(conn)
	fmt.Println("Conectado ao banco!")

	// ── CSV ───────────────────────────────────────────────────────────────────
	csvEmailRepo, err := storage.NewCSVEmailRepository("data/emails.csv")
	if err != nil {
		log.Fatal("Erro ao inicializar CSV de emails:", err)
	}

	csvAttachRepo, err := storage.NewCSVAttachmentRepository("data/attachments.csv")
	if err != nil {
		log.Fatal("Erro ao inicializar CSV de attachments:", err)
	}

	// ── Gmail ─────────────────────────────────────────────────────────────────
	client := auth.GetClient(config)
	gService, err := gmailclient.New(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatal(err)
	}

	msgs, err := gService.ListMessages(
		"in:inbox has:attachment filename:pdf",
		15,
	)
	if err != nil {
		log.Fatal(err)
	}

	for _, m := range msgs {
		msg, err := gService.GetMessage(m.Id)
		if err != nil {
			log.Println("Erro ao buscar mensagem:", err)
			continue
		}

		var subject, from, date string
		for _, h := range msg.Payload.Headers {
			switch h.Name {
			case "Subject":
				subject = h.Value
			case "From":
				from = h.Value
			case "Date":
				date = h.Value
			}
		}

		email := storage.Email{
			GmailID:    msg.Id,
			Subject:    subject,
			From:       from,
			ReceivedAt: date,
		}

		// Salva no banco
		inserted, emailID, err := emailRepo.Save(ctx, email)
		if err != nil {
			log.Println("Erro ao salvar e-mail no banco:", err)
			continue
		}

		// Salva no CSV (usa o mesmo emailID do banco para manter consistência)
		if _, _, csvErr := csvEmailRepo.Save(ctx, email); csvErr != nil {
			log.Println("Erro ao salvar e-mail no CSV:", csvErr)
		}

		if !inserted {
			fmt.Println("Já existe, pulando:", subject)
			continue
		}
		fmt.Println("Novo e-mail:", subject)

		pdfs, err := gService.ExtractPDFs(msg)
		if err != nil {
			log.Println("Erro ao extrair PDFs:", err)
			continue
		}

		for _, pdf := range pdfs {
			attachment := storage.Attachment{
				EmailID:  emailID,
				Filename: pdf.Filename,
				MimeType: pdf.MimeType,
				Path:     pdf.Path,
			}

			// Salva no banco
			if err := attachRepo.Save(ctx, attachment); err != nil {
				log.Println("Erro ao salvar attachment no banco:", err)
			}

			// Salva no CSV
			if err := csvAttachRepo.Save(ctx, attachment); err != nil {
				log.Println("Erro ao salvar attachment no CSV:", err)
			}
		}
	}
}
