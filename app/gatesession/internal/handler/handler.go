package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"kiyudesign.com/cesnow/light-server/app/gatesession/internal/svc"
	"kiyudesign.com/cesnow/light-server/pkg/transport"
	"reflect"
)

type RpcMetadata struct {
	ServerId    string
	ClientAddr  string
	AuthId      int64
	SessionId   int64
	ReceiveTime int64
	UserId      int64
	ClientMsgId int64
	IsBot       bool
	Layer       int32
	Client      string
	IsAdmin     bool
	TraceId     string
	SpanId      string
	LangPack    string
}

type Handler struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	MD              *transport.Metadata
	funcRegistry    map[string]reflect.Value
	argTypeRegistry map[string]reflect.Type
}

func New(ctx context.Context, svcCtx *svc.ServiceContext) *Handler {
	h := new(Handler)
	h.ctx = ctx
	h.svcCtx = svcCtx
	h.Logger = logx.WithContext(ctx)
	h.funcRegistry = map[string]reflect.Value{}
	h.argTypeRegistry = map[string]reflect.Type{}

	_ = h.register("t.account.signUp", h.AccountSignIn)

	return h
}

func (h *Handler) register(name string, fn interface{}) error {
	fnVal := reflect.ValueOf(fn)
	fnType := fnVal.Type()

	// Check function has exactly 1 input and 2 output
	if fnType.Kind() != reflect.Func || fnType.NumIn() != 2 || fnType.NumOut() != 2 {
		return errors.New("function must have 1 input and 2 outputs")
	}

	// Validate error as second return
	if fnType.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
		return errors.New("second return value must be error")
	}

	h.funcRegistry[name] = fnVal
	h.argTypeRegistry[name] = fnType.In(1)
	return nil
}

func (h *Handler) EventCall(name string, input interface{}, MD *transport.Metadata) (interface{}, error) {
	fn, ok := h.funcRegistry[name]
	if !ok {
		return nil, errors.New("function not found")
	}

	expectedArgType := h.argTypeRegistry[name]
	inVal := reflect.ValueOf(input)

	if inVal.Type() != expectedArgType {
		return nil, fmt.Errorf("expected argument of type %s but got %s", expectedArgType, inVal.Type())
	}

	h.MD = MD
	ctxVal := reflect.ValueOf(context.Background())
	results := fn.Call([]reflect.Value{ctxVal, inVal})

	if !results[1].IsNil() {
		return nil, results[1].Interface().(error)
	}

	return results[0].Interface(), nil
}
