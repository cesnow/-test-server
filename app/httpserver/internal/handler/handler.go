package handler

import (
	"context"
	"encoding/binary"
	"kiyudesign.com/cesnow/light-server/app/httpserver/internal/svc"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/metadata"

	"github.com/zeromicro/go-zero/core/logx"
)

type HttpserverCore struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	MD    *metadata.RpcMetadata
	Token string
}

func New(ctx context.Context, svcCtx *svc.ServiceContext, md *metadata.RpcMetadata) *HttpserverCore {
	return &HttpserverCore{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
		MD:     md,
	}
}

func serializeToBuffer(authId, sessionId int64, reqMsg *tproto.TMessage, obj tproto.TObject) []byte {
	data := tproto.EncodeTObject(obj)

	payload := make([]byte, 32)                                       // class + data
	binary.LittleEndian.PutUint64(payload[0:], uint64(authId))        // auth_id [1, 8]
	binary.LittleEndian.PutUint64(payload[8:], uint64(sessionId))     // session_id [2, 8]
	binary.LittleEndian.PutUint64(payload[16:], uint64(reqMsg.MsgId)) // msg_id [3, 8]
	binary.LittleEndian.PutUint32(payload[24:], uint32(reqMsg.SeqNo)) // seq_no [4, 4]
	binary.LittleEndian.PutUint32(payload[28:], uint32(len(data)))    // obj_length [5, 4]

	return append(payload, data...)
}
