package service

import (
	"context"
	"github.com/zeromicro/go-zero/zrpc"
	bffproxyclient "kiyudesign.com/cesnow/light-server/app/bff/proxy/client"
	"kiyudesign.com/cesnow/light-server/app/session/internal/config"
	statusclient "kiyudesign.com/cesnow/light-server/app/status/client"
	"kiyudesign.com/cesnow/light-server/pkg/net/ip"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/metadata"
)

type Service struct {
	*bffproxyclient.BFFProxyClient
	statusclient.StatusClient
	eGateServers map[string]*Gateway
	MyServerId   string
	*RpcShardingManager
}

func New(c config.Config) *Service {
	myServerId := ip.FigureOutListenOn(c.ListenOn)
	d := &Service{
		eGateServers:       make(map[string]*Gateway),
		BFFProxyClient:     bffproxyclient.NewBFFProxyClients(c.BFFProxyClients.Clients, c.BFFProxyClients.IDMap),
		StatusClient:       statusclient.NewStatusClient(zrpc.MustNewClient(c.StatusClient)),
		MyServerId:         myServerId,
		RpcShardingManager: NewRpcShardingManager(myServerId, c.Etcd),
	}

	d.watchGateway(c.GatewayClient)

	return d
}

func (d *Service) InvokeContext(ctx context.Context, rpcMetaData *metadata.RpcMetadata, object tproto.TObject) (tproto.TObject, error) {
	return d.BFFProxyClient.InvokeContext(ctx, rpcMetaData, object)
}
