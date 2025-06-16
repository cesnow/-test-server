package main

import (
	"fmt"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/zeromicro/go-zero/core/logx"
	"kiyudesign.com/cesnow/light-server/app/gatesession/internal/server"
	"kiyudesign.com/cesnow/light-server/pkg/commands"
	"strings"
)

func main() {

	catSchema, err := jsonschema.UnmarshalJSON(strings.NewReader(`{
        "type": "object",
        "properties": {
            "speak": { "const": "meow" }
        },
        "required": ["speak"]
    }`))
	if err != nil {
		logx.Infof(err.Error())
	}

	inst, err := jsonschema.UnmarshalJSON(strings.NewReader(`{"speak": "meow"}`))
	if err != nil {
		logx.Infof(err.Error())
	}

	c := jsonschema.NewCompiler()
	if err := c.AddResource("twsp://x/cat.json", catSchema); err != nil {
		logx.Infof(err.Error())
	}
	sch, err := c.Compile("twsp://x/cat.json")
	logx.Infof("%s", sch.Location)
	if err != nil {
		logx.Infof(err.Error())
	}
	err = sch.Validate(inst)
	fmt.Println("valid:", err == nil)

	commands.Run(new(server.Server))
}
