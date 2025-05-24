package http

import (
	"encoding/binary"
	"encoding/hex"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"io"
	"kiyudesign.com/cesnow/light-server/app/httpserver/internal/handler"
	"kiyudesign.com/cesnow/light-server/pkg/tproto/metadata"
	"net/http"
	"time"
)

func (s *Server) apiHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength <= 0 || r.Body == nil {
			http.NotFound(w, r)
		} else {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.NotFound(w, r)
			} else {
				authKeyId := int64(binary.LittleEndian.Uint64(body))

				c := handler.New(
					r.Context(),
					s.svcCtx,
					&metadata.RpcMetadata{
						ServerId:    s.svcCtx.Service.GetGatewayId(),
						ClientAddr:  httpx.GetRemoteAddr(r),
						AuthId:      authKeyId,
						SessionId:   0,
						ReceiveTime: time.Now().Unix(),
						UserId:      0,
						ClientMsgId: 0,
						IsBot:       false,
						Layer:       0,
						Client:      "",
						IsAdmin:     false,
						Takeout:     nil,
						LangPack:    "",
					})
				logx.Infof("api - authKeyId: %d, body: %s", authKeyId, hex.EncodeToString(body))
				rData, err2 := c.HttpserverApi(authKeyId, body)
				if err2 != nil {
					http.Error(w, err2.Error(), http.StatusInternalServerError)
				} else {
					w.Header().Set("Access-Control-Allow-Headers", "origin, content-type")
					w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
					w.Header().Set("Access-Control-Allow-Origin", "*")
					w.Header().Set("Access-Control-Max-Age", "1728000")
					w.Header().Set("Cache-control", "no-store")
					w.Header().Set("Connection", "keep-alive")
					w.Header().Set("Content-type", "application/octet-stream")
					w.Header().Set("Pragma", "no-cache")
					w.Header().Set("Strict-Transport-Security", "max-age=15768000")
					w.WriteHeader(200)

					_, _ = w.Write(rData)
				}
			}
		}
	}
}
