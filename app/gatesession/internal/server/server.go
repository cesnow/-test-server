package server

import (
	"flag"
	"kiyudesign.com/cesnow/light-server/app/gatesession/internal/config"
	"kiyudesign.com/cesnow/light-server/app/gatesession/internal/server/gnet"
	"kiyudesign.com/cesnow/light-server/app/gatesession/internal/server/mcp"
	"kiyudesign.com/cesnow/light-server/app/gatesession/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

var configFile = flag.String("f", "etc/httpserver.yaml", "the config file")

type Server struct {
	server *gnet.Server
	mcp    *mcp.Server
}

func (s *Server) Initialize() error {
	var c config.Config
	conf.MustLoad(*configFile, &c)

	logx.Infov(c)
	ctx := svc.NewServiceContext(c)

	s.mcp = mcp.New(ctx, c.MCP)
	s.server = gnet.New(ctx, c)

	return nil
}

func (s *Server) RunLoop() {

}

func (s *Server) Destroy() {
	s.server.Close()
}
