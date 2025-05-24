package http

import (
	"github.com/zeromicro/go-zero/rest"
	"net/http"
)

func (s *Server) RegisterHandlers() {
	s.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/api",
				Handler: s.apiHandler(),
			},
		},
	)
	s.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/sse",
		Handler: s.SSEHandler(),
	}, rest.WithSSE())
}
