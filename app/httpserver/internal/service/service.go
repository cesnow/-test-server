package service

import (
	"kiyudesign.com/cesnow/light-server/app/httpserver/internal/config"
)

type Service struct {
	session *Session
}

func New(c config.Config) (s *Service) {
	s = new(Service)
	s.session = NewSession(c)
	//s.cache = cache.NewLRUCache(10 * 1024 * 1024) // cache capacity: 10MB

	return s
}
