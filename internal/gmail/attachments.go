package gmail

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	gmailapi "google.golang.org/api/gmail/v1"
)

// percorre recursivamente as partes
func (s *Service) ExtractPDFs(msg *gmailapi.Message) error {
	return s.walkParts(msg.Payload.Parts, msg.Id)
}

func (s *Service) walkParts(parts []*gmailapi.MessagePart, msgID string) error {
	for _, p := range parts {

		// se tiver subpartes → recursão
		if len(p.Parts) > 0 {
			if err := s.walkParts(p.Parts, msgID); err != nil {
				return err
			}
			continue
		}

		// filtrar PDF
		if p.Filename == "" || p.MimeType != "application/pdf" {
			continue
		}

		fmt.Println("PDF encontrado:", p.Filename)

		if p.Body == nil || p.Body.AttachmentId == "" {
			continue
		}

		err := s.downloadAttachment(msgID, p.Body.AttachmentId, p.Filename)
		if err != nil {
			fmt.Println("Erro ao baixar:", err)
		}
	}

	return nil
}

func (s *Service) downloadAttachment(msgID, attachID, filename string) error {
	att, err := s.srv.Users.Messages.Attachments.Get("me", msgID, attachID).Do()
	if err != nil {
		return err
	}

	data, err := base64.URLEncoding.DecodeString(att.Data)
	if err != nil {
		return err
	}

	// diretório pdfs
	dir := "data/pdfs"
	os.MkdirAll(dir, os.ModePerm)

	path := filepath.Join(dir, msgID+"_"+filename)

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}

	fmt.Println("Salvo em:", path)

	return nil
}
