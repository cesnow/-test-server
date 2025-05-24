package gnet

import (
	"fmt"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
)

type CacheV struct {
	V *tproto.AuthIdInfo
}

func (c CacheV) Size() int {
	return 1
}

func (s *Server) GetAuth(authKeyId int64) *tproto.AuthIdInfo {
	var (
		cacheK = fmt.Sprintf("%d", authKeyId)
		value  *CacheV
	)

	if v, ok := s.cache.Get(cacheK); ok {
		value = v.(*CacheV)
	}

	if value == nil {
		return nil
	} else {
		return value.V
	}
}

func (s *Server) PutAuth(keyInfo *tproto.AuthIdInfo) {
	var (
		cacheK = fmt.Sprintf("%d", keyInfo.AuthId)
	)

	s.cache.Set(cacheK, &CacheV{V: keyInfo})
}
