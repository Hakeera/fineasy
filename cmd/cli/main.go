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

	fmt.Println("Conectado ao banco!")

	client := auth.GetClient(config)

	gService, err := gmailclient.New(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatal(err)
	}

	msgs, err := gService.ListMessages("in:inbox", 5)
	if err != nil {
		log.Fatal(err)
	}

	for _, m := range msgs {
		msg, err := gService.GetMessage(m.Id)
		if err != nil {
			continue
		}

		fmt.Println("------")

		for _, h := range msg.Payload.Headers {
			if h.Name == "Subject" || h.Name == "From" {
				fmt.Println(h.Name+":", h.Value)
			}
		}
	}
}
