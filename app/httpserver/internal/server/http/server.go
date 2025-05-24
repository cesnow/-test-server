package http

import (
	"github.com/zeromicro/go-zero/rest"
	"kiyudesign.com/cesnow/light-server/app/httpserver/internal/svc"
	"sync"
)

type Server struct {
	*rest.Server
	svcCtx *svc.ServiceContext
	// sse
	sseClients map[string]chan string
	sseMutex   sync.RWMutex
}

func New(ctx *svc.ServiceContext, c rest.RestConf) *Server {

	s := new(Server)
	s.svcCtx = ctx
	s.sseClients = make(map[string]chan string)

	s.Server = rest.MustNewServer(c, rest.WithCors("*"))

	s.RegisterHandlers()

	go s.SimulateEvents()
	go s.Serve()

	return s
}

func (s *Server) Serve() {
	defer s.Stop()
	s.Start()
}
