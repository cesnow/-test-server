package handler

import (
	"context"
	"encoding/binary"
	"google.golang.org/protobuf/types/known/anypb"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/sessionpb"
	"strconv"
)

func (c *HttpserverCore) HttpserverApi(authId int64, in []byte) ([]byte, error) {
	// authId [8], sessionId [8], msgId [8], seqNo [4], objLength [4], classId [4], data
	if authId == 0 {
		rMsg, err := c.onUnAuthenticatedMessage(in)
		if err != nil {
			c.Logger.Errorf("httpserver.api - error: %v", err)
			return nil, err
		}

		return rMsg, nil
	} else {
		rMsg, err := c.onUserReceivedMessage(authId, in)
		if err != nil {
			c.Logger.Errorf("httpserver.api - error: %v", err)
			return nil, err
		}

		return rMsg, nil
	}
}

func (c *HttpserverCore) onUnAuthenticatedMessage(in []byte) ([]byte, error) {

	tMsg, err := tproto.ConvertBytesToTMessage(in[16:])
	if err != nil {
		return nil, err
	}

	if tproto.IsAnyOfType(tMsg.Object, &tproto.Ping{}) {
		pong, _ := anypb.New(&tproto.Pong{
			MsgId:  tMsg.MsgId,
			PingId: tMsg.MsgId,
		})
		return serializeToBuffer(0, 0, tMsg, pong), nil
	} else if tproto.IsAnyOfType(tMsg.Object, &tproto.AuthIdInfo{}) {
		sessClient, err := c.svcCtx.Service.GetSessionClient("0")
		if err != nil {
			c.Logger.Errorf("onUserReceivedMessage - error: %v", err)
			return nil, err
		}
		newAuth, err := sessClient.SessionQueryAuthId(context.Background(), &sessionpb.TSessionQueryAuthId{})
		if err != nil {
			return nil, err
		}
		return serializeToBuffer(0, 0, tMsg, newAuth), nil
	}

	rpcErr, _ := anypb.New(&tproto.RpcError{
		ErrorCode:    404,
		ErrorMessage: "no handle, only ping and authIdInfo",
	})
	return serializeToBuffer(0, 0, tMsg, rpcErr), nil
}

func (c *HttpserverCore) onUserReceivedMessage(authId int64, msg []byte) ([]byte, error) {

	sessionId := int64(binary.LittleEndian.Uint64(msg[8:]))

	sessClient, err := c.svcCtx.Service.GetSessionClient(strconv.FormatInt(authId, 10))
	if err != nil {
		c.Logger.Errorf("onUserReceivedMessage - error: %v", err)
		return nil, err
	}

	rV, err := sessClient.SessionSendHttpDataToSession(c.ctx, &sessionpb.TSessionSendHttpDataToSession{
		Client: &sessionpb.SessionClientData{
			ServerId:  c.svcCtx.Service.GetGatewayId(),
			AuthId:    authId,
			SessionId: sessionId,
			ClientIp:  c.MD.ClientAddr,
			Payload:   msg[16:],
		},
	})
	if err != nil {
		c.Logger.Errorf("onUserReceivedMessage - error: %v", err)
		return nil, err
	}

	result, err := tproto.ConvertBytesToTMessage(rV.Payload)
	if err != nil {
		return nil, err
	}

	return serializeToBuffer(authId, sessionId, result, result.Object), nil
}
