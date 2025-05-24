package logic

import (
	"context"
	"github.com/zeromicro/go-zero/core/contextx"
	"github.com/zeromicro/go-zero/core/logx"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"reflect"
)

func (c *session) onInitConnection(ctx context.Context, gatewayId, clientIp string, msgId *inboxMsg, request *tproto.InitConnection) {
	logx.WithContext(ctx).Infof("onInitConnection - request data: {sess: %s, conn_id: %s, msg_id: %d, seq_no: %d, request: {%s}}",
		c,
		gatewayId,
		msgId.msgId,
		msgId.msgId,
		reflect.TypeOf(request))

	c.sessList.cb.onUpdateInitConnection(ctx, clientIp, request)

	// more message for processing
	// c.processMsg(ctx, gatewayId, clientIp, msgId, query)
}

func (c *session) onRpcRequest(ctx context.Context, gatewayId, clientIp string, msgId *inboxMsg, query tproto.TObject) bool {
	logx.WithContext(ctx).Infof("onRpcRequest - request data: {sess: %s, gatewayId: %s, msg_id: %d, seq_no: %d, request: {%s}}",
		c,
		gatewayId,
		msgId.msgId,
		msgId.seqNo,
		reflect.TypeOf(query))

	//switch query.(type) {
	//case *tproto.AccountUpdateStatus:
	//	c.sessList.cb.onSetMainUpdatesSession(ctx, c)
	//}

	switch c.sessList.cb.state {
	case AuthStateNormal:
		// state is ok
	default:
		if !checkRpcWithoutLogin(query) {
			c.sendRpcResultToQueue(ctx, gatewayId, msgId.msgId, tproto.NewRpcError(tproto.ErrAuthIdGenerate))
			msgId.state = RECEIVED | RESPONSE_GENERATED
			return false
		}
	}

	msgId.state = RECEIVED | RPC_PROCESSING

	if !c.isHttp {
		c.sessList.cb.tmpRpcApiMessageList = append(
			c.sessList.cb.tmpRpcApiMessageList,
			&rpcApiMessage{
				ctx:       contextx.ValueOnlyFrom(ctx),
				sessList:  c.sessList,
				sessionId: c.sessionId,
				clientIp:  clientIp,
				reqMsgId:  msgId.msgId,
				reqMsg:    query,
			})
	}

	return true
}

func (c *session) onRpcResult(ctx context.Context, rpcResult *rpcApiMessage) {
	rpcErr, _ := rpcResult.TryGetRpcResultError()
	if rpcErr != nil && rpcErr.GetErrorCode() == int32(tproto.ErrNotReturnClient) {
		logx.WithContext(ctx).Debugf("receive not return client")
		c.pendingQueue.Add(rpcResult.reqMsgId)
		return
	}

	// if logout -> c.sessList.cb.changeAuthState(ctx, tproto.AuthStateLogout, 0)

	c.sendRpcResult(ctx, rpcResult.MoveRpcResult())
}

func (c *session) sendRpcResult(ctx context.Context, rpcResult *tproto.RpcResult) {
	msgId := c.inQueue.Lookup(rpcResult.ReqMsgId)
	if msgId == nil {
		logx.WithContext(ctx).Errorf("not found msgId, maybe removed: %d", rpcResult.ReqMsgId)
		return
	}

	gatewayId := c.getGatewayId()
	c.sendRpcResultToQueue(ctx, gatewayId, msgId.msgId, rpcResult.Result)
	msgId.state = RECEIVED | ACKNOWLEDGED

	if gatewayId == "" {
		logx.WithContext(ctx).Errorf("gatewayId is empty, send delay...")
	} else {
		c.sendQueueToGateway(ctx, gatewayId)
	}
}
