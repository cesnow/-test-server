package logic

import (
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	configurationpb "kiyudesign.com/cesnow/light-server/pkg/tproto/bffpb/configuration"
)

func checkRpcWithoutLogin(tl tproto.TObject) bool {
	switch tl.(type) {

	case
		*tproto.Ping,
		*configurationpb.THelpTryNotifyUser,
		*configurationpb.THelpGetCountriesList:
		return true

	default:
		return false
	}
}
