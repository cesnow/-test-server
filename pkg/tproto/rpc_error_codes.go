package tproto

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	ErrSeeOther        codes.Code = 303
	ErrBadRequest      codes.Code = 400
	ErrUnauthorized    codes.Code = 401
	ErrForbidden       codes.Code = 403
	ErrNotFound        codes.Code = 404
	ErrNotAcceptable   codes.Code = 406
	ErrFlood           codes.Code = 420
	ErrInternal        codes.Code = 500
	ErrTimeOut503      codes.Code = 5030000
	ErrNotReturnClient codes.Code = 700
)

var (
	ErrNonBusinessLogicImplement = status.Error(ErrBadRequest, "ERR_NON_BUSINESS_LOGIC_IMPL")
	ErrNonBusinessServiceFound   = status.Error(ErrBadRequest, "ERR_NON_BUSINESS_SERVICE_NOT_FOUND")
	ErrNonBusinessMethodFound    = status.Error(ErrNotFound, "ERR_NON_BUSINESS_METHOD_NOT_FOUND")
)

var (
	ErrAuthIdGenerate      = status.Error(ErrUnauthorized, "AUTH_ID_UNREGISTERED")
	ErrInternalServerError = status.Error(ErrInternal, "INTERNAL_SERVER_ERROR")
	ErrInputRequestInvalid = status.Error(ErrBadRequest, "INPUT_REQUEST_INVALID")
)

func NewErrRedirectToX(v string) error {
	return status.Errorf(ErrNotReturnClient, "REDIRECT_TO_%s", v)
}
