package authsession

import (
	"context"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/zeromicro/go-zero/core/logx"
	"reflect"
)

func (c *session) onEventRequest(ctx context.Context, gatewayId, clientIp string, msgId *inboxMsg, event string, query []byte) bool {
	logx.WithContext(ctx).Infof("onEventRequest - request data: {sess: %s, gatewayId: %s, msg_id: %d, seq_no: %d, request: {%s}}",
		c,
		gatewayId,
		msgId.msgId,
		msgId.seqNo,
		reflect.TypeOf(query))

	c.onEventRequestBeforeProcess(ctx, gatewayId, clientIp, msgId, event, query)

	// check auth state
	switch c.sessList.cb.state {
	case AuthStateNormal:
		// state is ok
	default:
		if !checkEventWithoutLogin(event) {
			// send
			msgId.state = RECEIVED | RESPONSE_GENERATED
			return false
		}
	}

	//msgId.state = RECEIVED | DATA_PROCESSING
	msgId.state = RECEIVED | NO_NEED_ACK

	var payload any
	_ = msgpack.Unmarshal(query, payload)

	// TODO: process request
	result, err := c.sessList.cb.eventHandler(event, payload)
	if err != nil {
		logx.WithContext(ctx).Errorf("onEventRequest - error: {sess: %s, gatewayId: %s, msg_id: %d, seq_no: %d, request: {%s}}",
			c,
			gatewayId,
			msgId.msgId,
			msgId.seqNo,
			reflect.TypeOf(query))
		return false
	}

	c.onEventRequestAfterProcess(ctx, gatewayId, clientIp, msgId, event, payload, result)

	resultBytes, err := msgpack.Marshal(result)
	if err != nil {
		logx.WithContext(ctx).Errorf("onEventRequest - error: {sess: %s, gatewayId: %s, msg_id: %d, seq_no: %d, request: {%s}}",
			c,
			gatewayId,
			msgId.msgId,
			msgId.seqNo,
			reflect.TypeOf(query))
		return false
	}

	// send
	_, _ = c.sendDirectToGateway(ctx, gatewayId, false, resultBytes)

	return true
}

func checkEventWithoutLogin(event string) bool {
	switch event {
	// account auth helper
	default:
		return false
	}
}

func (c *session) onEventRequestBeforeProcess(ctx context.Context, id string, ip string, inboxMsg *inboxMsg, event string, query []byte) {
	// update status
	c.sessList.cb.onSetMainUpdatesSession(ctx, c)
}

func (c *session) onEventRequestAfterProcess(ctx context.Context, id string, ip string, inboxMsg *inboxMsg, event string, payload any, result any) {
	// about auth
	switch event {
	case "account.auth.logout":
		c.sessList.cb.changeAuthState(ctx, AuthStateLogout, 0)
	case "account.auth.signIn":
		//accountSignIn := result.(*handler.AccountSignInResponse)
		c.sessList.cb.changeAuthState(ctx, AuthStateNormal, 1010)
	default:
	}
}
