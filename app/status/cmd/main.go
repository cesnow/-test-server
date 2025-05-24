package main

import (
	"kiyudesign.com/cesnow/light-server/app/status/internal/server"
	"kiyudesign.com/cesnow/light-server/pkg/commands"
)

func main() {
	commands.Run(new(server.Server))
}
