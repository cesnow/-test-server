package handler

import "github.com/zeromicro/go-zero/core/logx"

type AccountSignUpRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AccountSignUpResponse struct {
	Ok bool `json:"ok"`
}

func (h *Handler) AccountSignUp(in *AccountSignUpRequest) (*AccountSignUpResponse, error) {

	logx.Infof("AccountSignUp - %+v", in)

	return &AccountSignUpResponse{Ok: true}, nil
}
