package gmail

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	gmailapi "google.golang.org/api/gmail/v1"
)

type PDFResult struct {
	Filename string
	MimeType string
	Path     string
}

func (s *Service) ExtractPDFs(msg *gmailapi.Message) ([]PDFResult, error) {
	return s.walkParts(msg.Payload.Parts, msg.Id)
}

func (s *Service) walkParts(parts []*gmailapi.MessagePart, msgID string) ([]PDFResult, error) {
	var results []PDFResult

	for _, p := range parts {
		if len(p.Parts) > 0 {
			sub, err := s.walkParts(p.Parts, msgID)
			if err != nil {
				return nil, err
			}
			results = append(results, sub...)
			continue
		}

		if p.Filename == "" || p.MimeType != "application/pdf" {
			continue
		}

		if p.Body == nil || p.Body.AttachmentId == "" {
			continue
		}

		fmt.Println("PDF encontrado:", p.Filename)

		path, err := s.downloadAttachment(msgID, p.Body.AttachmentId, p.Filename)
		if err != nil {
			fmt.Println("Erro ao baixar:", err)
			continue
		}

		results = append(results, PDFResult{
			Filename: p.Filename,
			MimeType: p.MimeType,
			Path:     path,
		})
	}

	return results, nil
}

func (s *Service) downloadAttachment(msgID, attachID, filename string) (string, error) {
	att, err := s.srv.Users.Messages.Attachments.Get("me", msgID, attachID).Do()
	if err != nil {
		return "", err
	}

	data, err := base64.URLEncoding.DecodeString(att.Data)
	if err != nil {
		return "", err
	}

	dir := "data/pdfs"
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", err
	}

	path := filepath.Join(dir, msgID+"_"+filename)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", err
	}

	fmt.Println("Salvo em:", path)
	return path, nil
}
