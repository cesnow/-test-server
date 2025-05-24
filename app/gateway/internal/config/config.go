package config

import (
	"github.com/zeromicro/go-zero/zrpc"
	compare "kiyudesign.com/cesnow/light-server/pkg/contains"
)

type Config struct {
	zrpc.RpcServerConf
	GNet    *GNetConfig
	Session zrpc.RpcClientConf
}

type GNetConfig struct {
	Address    []string
	Multicore  bool
	SendBuf    int
	ReceiveBuf int
}

func (c GNetConfig) ToAddresses() []string {
	var addresses []string
	for _, address := range c.Address {
		if ok := compare.ContainsString(addresses, address); !ok {
			addresses = append(addresses, "tcp://"+address)
		}
	}
	return addresses
}
