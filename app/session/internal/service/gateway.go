package service

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
)

func (d *Service) watchGateway(c zrpc.RpcClientConf) {
	sub, _ := discov.NewSubscriber(c.Etcd.Hosts, c.Etcd.Key)
	update := func() {
		values := sub.Values()
		if len(values) == 0 {
			return
		}

		clients := map[string]*Gateway{}
		for _, v := range values {
			if old, ok := d.eGateServers[v]; ok {
				clients[v] = old
				continue
			}
			c.Endpoints = []string{v}
			cli, err := NewGateway(c)
			if err != nil {
				logx.Errorf("watchComet NewClient(%v) error(%v)", values, err)
				return
			}
			clients[v] = cli
		}

		for key, old := range d.eGateServers {
			if _, ok := clients[key]; !ok {
				old.cancel()
				logx.Infof("watchComet DelComet:%s", key)
			}
		}

		d.eGateServers = clients
	}

	sub.AddListener(update)
	update()
}

func (d *Service) SendDataToGateway(ctx context.Context, gatewayId string, authId, sessionId int64, msg *tproto.MsgRawData) (bool, error) {
	if c, ok := d.eGateServers[gatewayId]; ok {
		return c.SendDataToGate(ctx, authId, sessionId, msg.Encode())
	} else {
		logx.WithContext(ctx).Errorf("not found k: %s, %v", gatewayId, d.eGateServers)
		return false, fmt.Errorf("not found k: %s", gatewayId)
	}
}

func (d *Service) SendHttpDataToGateway(ctx context.Context, ch chan interface{}, authKeyId, sessionId int64, msg *tproto.MsgRawData) (bool, error) {
	select {
	case ch <- msg.Encode():
		// maybe add session to msg encode
		close(ch)
		return true, nil
	default:
		logx.WithContext(ctx).Errorf("Default fail !!!!! ch closed")
		return false, fmt.Errorf("ch closed")
	}
}
