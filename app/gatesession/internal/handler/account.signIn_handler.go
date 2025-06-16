package handler

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/zeromicro/go-zero/core/logx"
)

type AccountSignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ssoClaims struct {
	jwt.RegisteredClaims
}

type AccountSignInResponse struct {
	AuthId        int64     `json:"authId"`
	SessionId     int64     `json:"sessionId"`
	Account       string    `json:"account"`
	AccountAuthId int64     `json:"accountAuthId"`
	Jwt           ssoClaims `json:"jwt"`
}

func (h *Handler) AccountSignIn(ctx context.Context, in *AccountSignInRequest) (*AccountSignInResponse, error) {

	logx.Infof("AccountSignIn - %+v", in)

	return &AccountSignInResponse{
		AuthId:        0,
		SessionId:     0,
		Account:       "",
		AccountAuthId: 0,
		Jwt: ssoClaims{
			RegisteredClaims: jwt.RegisteredClaims{},
		},
	}, nil
}
