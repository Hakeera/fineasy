package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"golang.org/x/oauth2"
)

const tokFile = "token.json"

func GetClient(config *oauth2.Config) *http.Client {
	tok, err := tokenFromFile(tokFile)
	if err != nil || !tok.Valid() {
		os.Remove(tokFile)
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	codeCh := make(chan string)
	config.RedirectURL = "http://localhost:8080/"

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Código não encontrado", http.StatusBadRequest)
			return
		}
		fmt.Fprintln(w, "Autorização recebida! Pode fechar.")
		codeCh <- code
	})

	server := &http.Server{Addr: ":8080", Handler: mux}
	go server.ListenAndServe()

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	exec.Command("xdg-open", authURL).Start()
	fmt.Println("Ou acesse:", authURL)

	code := <-codeCh
	server.Shutdown(context.Background())

	tok, err := config.Exchange(context.Background(), code)
	if err != nil {
		log.Fatalf("Erro ao trocar código por token: %v", err)
	}
	return tok
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tok := &oauth2.Token{}
	if err := json.NewDecoder(f).Decode(tok); err != nil {
		return nil, err
	}
	return tok, nil
}

func saveToken(path string, token *oauth2.Token) {
	f, err := os.Create(path)
	if err != nil {
		log.Printf("Aviso: não foi possível salvar token: %v", err)
		return
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
