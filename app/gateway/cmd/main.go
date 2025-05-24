package main

import (
	"kiyudesign.com/cesnow/light-server/app/gateway/internal/server"
	"kiyudesign.com/cesnow/light-server/pkg/commands"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
)

func main() {
	tproto.InitChecksum()
	commands.Run(new(server.Server))
}
