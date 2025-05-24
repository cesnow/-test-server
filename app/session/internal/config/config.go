package config

import (
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	GatewayClient   zrpc.RpcClientConf
	StatusClient    zrpc.RpcClientConf
	BFFProxyClients BFFProxyClients
}

type BFFProxyClients struct {
	Clients []zrpc.RpcClientConf
	IDMap   map[string]string
}
