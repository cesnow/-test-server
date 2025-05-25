package mcp

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/mcp"
	"kiyudesign.com/cesnow/light-server/app/gatesession/internal/svc"
)

type Server struct {
	mcpServer mcp.McpServer
	svcCtx    *svc.ServiceContext
	c         mcp.McpConf
}

func New(ctx *svc.ServiceContext, c mcp.McpConf) *Server {
	s := new(Server)

	s.mcpServer = mcp.NewMcpServer(c)
	s.svcCtx = ctx
	s.c = c

	s.registerCalculationTool()
	s.registerEchoTool()
	s.registerTaskTool()

	go func() {
		s.Start()
	}()

	return s
}

func (s *Server) Start() {
	defer s.mcpServer.Stop()
	logx.Infof("mcp server is listening on port %d", s.c.Port)
	s.mcpServer.Start()
}
