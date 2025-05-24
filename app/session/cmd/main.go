package main

import (
	"kiyudesign.com/cesnow/light-server/app/session/internal/server"
	"kiyudesign.com/cesnow/light-server/pkg/commands"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	_ "kiyudesign.com/cesnow/light-server/pkg/tproto/bffpb/configuration"
	_ "kiyudesign.com/cesnow/light-server/pkg/tproto/gatewaypb"
	_ "kiyudesign.com/cesnow/light-server/pkg/tproto/sessionpb"
	_ "kiyudesign.com/cesnow/light-server/pkg/tproto/statuspb"
	_ "kiyudesign.com/cesnow/light-server/pkg/tproto/syncpb"
)

func main() {
	tproto.InitChecksum()
	commands.Run(new(server.Server))
}
