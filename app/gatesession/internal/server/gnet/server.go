package gnet

import (
	"context"
	"github.com/panjf2000/gnet/v2"
	"github.com/panjf2000/gnet/v2/pkg/logging"
	"github.com/panjf2000/gnet/v2/pkg/pool/goroutine"
	"github.com/zeromicro/go-zero/core/logx"
	"kiyudesign.com/cesnow/light-server/app/gatesession/internal/config"
	"kiyudesign.com/cesnow/light-server/app/gatesession/internal/svc"
	"kiyudesign.com/cesnow/light-server/pkg/cache"
	"runtime"
	"sync"
	"time"
)

type Server struct {
	gnet.BuiltinEventEngine
	etcdEndpoints  string
	eng            gnet.Engine
	numEventLoop   int
	pool           *goroutine.Pool
	cache          *cache.LRUCache
	c              *config.Config
	authSessionMgr *AuthSessionManager
	svcCtx         *svc.ServiceContext
	tickNumber     int64
	connections    sync.Map
}

func New(svcCtx *svc.ServiceContext, c config.Config) *Server {

	s := new(Server)

	s.cache = cache.NewLRUCache(10 * 1024 * 1024) // cache capacity: 10MB
	s.pool = goroutine.Default()
	s.authSessionMgr = NewAuthSessionManager()

	s.numEventLoop = 1
	if c.GNet.Multicore {
		s.numEventLoop = runtime.NumCPU()
	}

	s.c = &c
	s.svcCtx = svcCtx

	go func() {
		s.Serve()
	}()
	return s
}

func (s *Server) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	logx.Debugf("stop engine... error: %v", s.eng.Stop(ctx))
}

func (s *Server) Serve() {
	logx.Debugf("addrs: %s", s.c.GNet.ToAddresses())

	err := gnet.Rotate(
		s,
		s.c.GNet.ToAddresses(),
		gnet.WithMulticore(s.c.GNet.Multicore),
		gnet.WithLoadBalancing(gnet.SourceAddrHash),
		gnet.WithSocketRecvBuffer(s.c.GNet.ReceiveBuf),
		gnet.WithSocketSendBuffer(s.c.GNet.SendBuf),
		gnet.WithLockOSThread(true),
		gnet.WithReuseAddr(true),
		gnet.WithTicker(true),
		gnet.WithLogLevel(logging.DebugLevel),
		gnet.WithLogger(NewLogger()))
	if err != nil {
		logx.Error(err)
		panic(err)
	}
}
