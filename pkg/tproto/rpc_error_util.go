package tproto

import (
	"errors"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func toTProtoErrorCode(code codes.Code) codes.Code {
	switch code {
	case ErrSeeOther:
		return code
	case ErrBadRequest:
		return code
	case ErrUnauthorized:
		return code
	case ErrForbidden:
		return code
	case ErrNotFound:
		return code
	case ErrNotAcceptable:
		return code
	case ErrFlood:
		return code
	case ErrInternal:
		return code
	case ErrNotReturnClient:
		return code
	default:
		return ErrInternal
	}
}

func NewRpcError(e error) *RpcError {
	if e == nil {
		return nil
	}

	if rErr, ok := status.FromError(e); ok {
		return &RpcError{
			ErrorCode:    int32(toTProtoErrorCode(rErr.Code())),
			ErrorMessage: rErr.Message(),
		}
	} else {
		var err *RpcError
		switch {
		case errors.As(e, &err):
			return err
		default:
			return &RpcError{
				ErrorCode:    int32(toTProtoErrorCode(codes.Internal)),
				ErrorMessage: "INTERNAL_SERVER_ERROR",
			}
		}
	}
}

func (m *RpcError) IsOK() bool {
	if m == nil {
		return true
	}
	return m.GetErrorCode() == int32(codes.OK)
}

func (m *RpcError) Error() string {
	if m == nil {
		return ""
	}

	return fmt.Sprintf("rpc(RpcError) error: code = %d desc = %s", m.Code(), m.Message())
}

func (m *RpcError) Code() int {
	return int(m.GetErrorCode())
}

func (m *RpcError) Message() string {
	return m.GetErrorMessage()
}

func (m *RpcError) Details() []interface{} {
	return nil
}
