package logic

import (
	"context"
	"fmt"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
)

type rpcApiMessage struct {
	ctx       context.Context
	sessList  *SessionList
	sessionId int64
	clientIp  string
	reqMsgId  int64
	reqMsg    tproto.TObject
	rpcResult *tproto.RpcResult
}

func (m *rpcApiMessage) MoveRpcResult() *tproto.RpcResult {
	v := m.rpcResult
	m.rpcResult = nil
	return v
}

func (m *rpcApiMessage) TryGetRpcResultError() (*tproto.RpcError, bool) {
	if m.rpcResult != nil && m.rpcResult.Result != nil {
		r := m.rpcResult.Result
		if tproto.IsAnyOfType(r, &tproto.RpcError{}) {
			rpcErr, _ := r.UnmarshalNew()
			return rpcErr.(*tproto.RpcError), true
		}
	}

	return nil, false
}

func (m *rpcApiMessage) DebugString() string {
	if m.rpcResult == nil {
		return fmt.Sprintf("{session_id: %d, req_msg_id: %d, req_msg: %s}",
			m.sessionId,
			m.reqMsgId,
			m.reqMsg)
	} else {
		return fmt.Sprintf("{session_id: %d, req_msg_id: %d, req_msg: %s, rpc_result: %s}",
			m.sessionId,
			m.reqMsgId,
			m.reqMsg,
			m.rpcResult.Result)
	}
}
