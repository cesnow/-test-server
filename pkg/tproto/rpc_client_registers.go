package tproto

import (
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"reflect"
	"strings"
	"unicode"
)

type newRpcReplyFunc func() any

type RpcContextTuple struct {
	Method       string
	NewReplyFunc newRpcReplyFunc
}

var rpcContextRegisters = map[string]RpcContextTuple{
	//"THelpGetCountriesList": {"/configuration.RpcConfiguration/help_getCountriesList", func() any { return new(configurationpb.Help_CountriesList) }},
}

func ToPascalCase(s string) string {
	parts := strings.Split(s, "_")
	for i, p := range parts {
		if p == "" {
			continue
		}
		runes := []rune(p)
		runes[0] = unicode.ToUpper(runes[0])
		parts[i] = string(runes)
	}
	return strings.Join(parts, "")
}

func initRpcContextRegisters() {
	rpcContextRegisters = make(map[string]RpcContextTuple)
	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {

		logx.Debugf("read file %s", fd.Path())

		for i := 0; i < fd.Services().Len(); i++ {
			svc := fd.Services().Get(i)
			for j := 0; j < svc.Methods().Len(); j++ {
				m := svc.Methods().Get(j)
				inName := ToPascalCase(string(m.Input().Name())) // e.g. "THelpGetCountriesList"
				//outName := m.Output().FullName()   // e.g. "configuration.Help_CountriesList"
				fullMethod := "/" + string(svc.FullName()) + "/" + string(m.Name())

				//logx.Infof("Registered inName: %s, fullMethod: %s", inName, fullMethod)

				//input := m.Input()
				output := m.Output()

				rpcContextRegisters[inName] = RpcContextTuple{
					Method: fullMethod,
					NewReplyFunc: func() any {
						mt, err := protoregistry.GlobalTypes.FindMessageByName(output.FullName())
						if err != nil {
							panic(err)
						}
						return mt.New().Interface()
					},
				}
			}
		}
		return true
	})
}

func FindRpcContextTuple(t interface{}) *RpcContextTuple {
	rt := reflect.TypeOf(t)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	m, ok := rpcContextRegisters[rt.Name()]
	if !ok {
		// log.Errorf("Can't find name: %s", rt.Name())
		return nil
	}
	return &m
}

func GetRpcContextRegisters() map[string]RpcContextTuple {
	initRpcContextRegisters()
	return rpcContextRegisters
}
