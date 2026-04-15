package gmail

import (
	"context"

	gmailapi "google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type Service struct {
	srv *gmailapi.Service
}

func New(ctx context.Context, client option.ClientOption) (*Service, error) {
	srv, err := gmailapi.NewService(ctx, client)
	if err != nil {
		return nil, err
	}

	return &Service{srv: srv}, nil
}

func (s *Service) ListMessages(query string, max int64) ([]*gmailapi.Message, error) {
	res, err := s.srv.Users.Messages.
		List("me").
		Q(query).
		MaxResults(max).
		Do()

	if err != nil {
		return nil, err
	}

	return res.Messages, nil
}

func (s *Service) GetMessage(id string) (*gmailapi.Message, error) {
	return s.srv.Users.Messages.Get("me", id).Do()
}
