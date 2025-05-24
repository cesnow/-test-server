package sessionclient

import (
	"context"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/sessionpb"

	"github.com/zeromicro/go-zero/zrpc"
)

type SessionClient interface {
	SessionQueryAuthId(ctx context.Context, in *sessionpb.TSessionQueryAuthId) (*tproto.AuthIdInfo, error)
	SessionCreateSession(ctx context.Context, in *sessionpb.TSessionCreateSession) (*tproto.Bool, error)
	SessionSendDataToSession(ctx context.Context, in *sessionpb.TSessionSendDataToSession) (*tproto.Bool, error)
	SessionSendHttpDataToSession(ctx context.Context, in *sessionpb.TSessionSendHttpDataToSession) (*sessionpb.HttpSessionData, error)
	SessionCloseSession(ctx context.Context, in *sessionpb.TSessionCloseSession) (*tproto.Bool, error)
	SessionPushUpdatesData(ctx context.Context, in *sessionpb.TSessionPushUpdatesData) (*tproto.Bool, error)
	SessionPushSessionUpdatesData(ctx context.Context, in *sessionpb.TSessionPushSessionUpdatesData) (*tproto.Bool, error)
	SessionPushRpcResultData(ctx context.Context, in *sessionpb.TSessionPushRpcResultData) (*tproto.Bool, error)
}

type defaultSessionClient struct {
	cli zrpc.Client
}

func NewSessionClient(cli zrpc.Client) SessionClient {
	return &defaultSessionClient{
		cli: cli,
	}
}

func (m *defaultSessionClient) SessionQueryAuthId(ctx context.Context, in *sessionpb.TSessionQueryAuthId) (*tproto.AuthIdInfo, error) {
	client := sessionpb.NewRpcSessionClient(m.cli.Conn())
	return client.SessionQueryAuthId(ctx, in)
}

func (m *defaultSessionClient) SessionCreateSession(ctx context.Context, in *sessionpb.TSessionCreateSession) (*tproto.Bool, error) {
	client := sessionpb.NewRpcSessionClient(m.cli.Conn())
	return client.SessionCreateSession(ctx, in)
}

func (m *defaultSessionClient) SessionSendDataToSession(ctx context.Context, in *sessionpb.TSessionSendDataToSession) (*tproto.Bool, error) {
	client := sessionpb.NewRpcSessionClient(m.cli.Conn())
	return client.SessionSendDataToSession(ctx, in)
}

func (m *defaultSessionClient) SessionSendHttpDataToSession(ctx context.Context, in *sessionpb.TSessionSendHttpDataToSession) (*sessionpb.HttpSessionData, error) {
	client := sessionpb.NewRpcSessionClient(m.cli.Conn())
	return client.SessionSendHttpDataToSession(ctx, in)
}

func (m *defaultSessionClient) SessionCloseSession(ctx context.Context, in *sessionpb.TSessionCloseSession) (*tproto.Bool, error) {
	client := sessionpb.NewRpcSessionClient(m.cli.Conn())
	return client.SessionCloseSession(ctx, in)
}

func (m *defaultSessionClient) SessionPushUpdatesData(ctx context.Context, in *sessionpb.TSessionPushUpdatesData) (*tproto.Bool, error) {
	client := sessionpb.NewRpcSessionClient(m.cli.Conn())
	return client.SessionPushUpdatesData(ctx, in)
}

func (m *defaultSessionClient) SessionPushSessionUpdatesData(ctx context.Context, in *sessionpb.TSessionPushSessionUpdatesData) (*tproto.Bool, error) {
	client := sessionpb.NewRpcSessionClient(m.cli.Conn())
	return client.SessionPushSessionUpdatesData(ctx, in)
}

func (m *defaultSessionClient) SessionPushRpcResultData(ctx context.Context, in *sessionpb.TSessionPushRpcResultData) (*tproto.Bool, error) {
	client := sessionpb.NewRpcSessionClient(m.cli.Conn())
	return client.SessionPushRpcResultData(ctx, in)
}
