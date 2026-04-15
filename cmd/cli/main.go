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

	// Conexão DB
	conn, err := storage.NewConnection(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	repo := storage.NewEmailRepository(conn)

	fmt.Println("Conectado ao banco!")

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

		// Salvar email no banco
		inserted, err := repo.Save(ctx, email)
		if err != nil {
			log.Println("Erro ao salvar:", err)
			continue
		}
		if !inserted {
			fmt.Println("Já existe, pulando:", subject)
			continue
		}

		fmt.Println("Novo email:", subject)

		// Salvar Pdf
		err = gService.ExtractPDFs(msg)
		if err != nil {
			log.Println("Erro ao extrair PDFs:", err)
		}
	}
}
