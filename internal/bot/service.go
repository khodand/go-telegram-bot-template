package bot

import (
	tele "gopkg.in/telebot.v3"
)

type Service struct {
	client *tele.Bot
}

func New(client *tele.Bot) *Service {
	return &Service{
		client: client,
	}
}

func (s *Service) RegisterRoutes() {
	s.client.Handle("/start", func(c tele.Context) error {
		return c.Send("Hello world!")
	})
}
