package client

import (
	"errors"
	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/core/hash"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stringx"
	"github.com/zeromicro/go-zero/zrpc"
	"kiyudesign.com/cesnow/light-server/app/gateway/internal/config"
	sessionclient "kiyudesign.com/cesnow/light-server/app/session/client"
	"strings"
)

var (
	ErrSessionNotFound = errors.New("not found session")
)

type ShardingSession struct {
	dispatcher *hash.ConsistentHash
	sessions   map[string]sessionclient.SessionClient
}

func NewShardingSession(c config.Config) *ShardingSession {
	sess := &ShardingSession{
		dispatcher: hash.NewConsistentHash(),
		sessions:   make(map[string]sessionclient.SessionClient),
	}
	sess.watch(c.Session)

	return sess
}

func (sess *ShardingSession) watch(c zrpc.RpcClientConf) {
	sub, _ := discov.NewSubscriber(c.Etcd.Hosts, c.Etcd.Key)
	update := func() {
		var (
			addClients    []string
			removeClients []string
		)

		values := sub.Values()
		sessions := map[string]sessionclient.SessionClient{}
		for _, v := range values {
			if old, ok := sess.sessions[v]; ok {
				sessions[v] = old
				continue
			}
			c.Endpoints = []string{v}
			cli, err := zrpc.NewClient(c)
			if err != nil {
				logx.Errorf("watchComet NewClient(%s) error(%s)", strings.Join(values, ", "), err.Error())
				return
			}
			sessionCli := sessionclient.NewSessionClient(cli)
			sessions[v] = sessionCli

			addClients = append(addClients, v)
		}

		for key := range sess.sessions {
			if !stringx.Contains(values, key) {
				removeClients = append(removeClients, key)
			}
		}

		for _, n := range addClients {
			sess.dispatcher.Add(n)
		}

		for _, n := range removeClients {
			sess.dispatcher.Remove(n)
		}

		sess.sessions = sessions
	}

	sub.AddListener(update)
	update()
}

func (sess *ShardingSession) InvokeByKey(key string, cb func(client sessionclient.SessionClient) (err error)) error {
	val, ok := sess.dispatcher.Get(key)
	if !ok {
		return ErrSessionNotFound
	}

	cli, ok := sess.sessions[val.(string)]
	if !ok {
		return ErrSessionNotFound
	}

	if cb == nil {
		return nil
	}

	return cb(cli)
}
