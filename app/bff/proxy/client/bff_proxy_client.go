package bff_proxy_client

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"kiyudesign.com/cesnow/light-server/pkg/net/rpcx"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	metadata2 "kiyudesign.com/cesnow/light-server/pkg/tproto/metadata"
	"reflect"
	"strings"
	"time"
)

type BFFProxyClient struct {
	BFFClients map[string]zrpc.Client
}

func NewBFFProxyClients(cList []zrpc.RpcClientConf, idMap map[string]string) *BFFProxyClient {
	var (
		clients   = make(map[string]zrpc.Client)
		registers = tproto.GetRpcContextRegisters()
	)

	for _, c := range cList {
		cli := rpcx.GetCachedRpcClient(c)
		for k, v := range idMap {
			if v == c.Etcd.Key {
				clients[k] = cli
			}
		}
	}

	bizClients := make(map[string]zrpc.Client)
	for m, ctx := range registers {
		for k := range idMap {
			if strings.HasPrefix(ctx.Method, k) {
				bizClients[m] = clients[k]
				break
			}
		}
	}

	return &BFFProxyClient{
		BFFClients: bizClients,
	}
}

func (c *BFFProxyClient) GetRpcClientByRequest(t interface{}) (zrpc.Client, error) {
	rt := reflect.TypeOf(t)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	if c2, ok := c.BFFClients[rt.Name()]; ok {
		return c2, nil
	} else {
		logx.Errorf("not found method: %s", rt.Name())
	}

	return nil, tproto.ErrNonBusinessMethodFound
}

func (c *BFFProxyClient) Invoke(rpcMetaData *metadata2.RpcMetadata, object tproto.TObject) (tproto.TObject, error) {
	return c.InvokeContext(context.Background(), rpcMetaData, object)
}

func (c *BFFProxyClient) InvokeContext(ctx context.Context, rpcMetaData *metadata2.RpcMetadata, object tproto.TObject) (tproto.TObject, error) {
	logger := logx.WithContext(ctx)

	conn, err := c.GetRpcClientByRequest(object)
	if err != nil {
		return nil, tproto.NewRpcError(err)
	}

	t := tproto.FindRpcContextTuple(object)
	if t == nil {
		err = fmt.Errorf("Invoke error: %v not regist!\n", object)
		logger.Error("FindRPCContextTuple error: %v", err)
		return nil, tproto.NewRpcError(tproto.ErrNonBusinessLogicImplement)
	}

	// logx.Infof("Invoke - method: {%s}", t.Method)
	r := t.NewReplyFunc()
	// logx.Infof("Invoke - NewReplyFunc: {%#v}, t: {%v}", r, reflect.TypeOf(r))

	var (
		header, trailer metadata2.MD
		ctxWithTimeout  context.Context
		ctxCancelFunc   context.CancelFunc
	)

	ctxWithTimeout, ctxCancelFunc = context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancelFunc()

	ctx2, _ := metadata2.RpcMetadataToOutgoing(ctxWithTimeout, rpcMetaData)
	rt := time.Now()

	logger.Debugf("Invoke - NewReplyFunc: {%#v}", r)
	err = conn.Conn().Invoke(ctx2, t.Method, object, r, grpc.Header(&header), grpc.Trailer(&trailer))

	logger.Debugf("rpc Invoke: {method: %s, metadata: %s, result: {%s}, error: {%s}}, cost = %v",
		t.Method,
		rpcMetaData,
		reflect.TypeOf(r),
		err,
		time.Since(rt))

	if err != nil {
		logger.Errorf("RPC Invoke error: {method: %s, metadata: %s, error: %s}", t.Method, rpcMetaData, err)
		return nil, tproto.NewRpcError(err)
	} else {
		if reply, ok := r.(tproto.TObject); !ok {
			logger.Errorf("invalid reply type, maybe server side bug, %v", reply)
			return nil, tproto.NewRpcError(tproto.ErrInternalServerError)
		} else {
			return reply, nil
		}
	}
}
