package http_util

import (
	"bytes"
	"fmt"
	"github.com/zeromicro/go-zero/core/jsonx"
	"github.com/zeromicro/go-zero/core/logx"
	"io"
	"kiyudesign.com/cesnow/light-server/pkg/net/http/binding"
	"net/http"
	"strings"
)

type HttpApiRequest interface {
	Method() string
}

type HttpApiMethod interface {
	NewRequest() HttpApiRequest
	Decode2(c *http.Request, contentType string, r HttpApiRequest) (err error)
}

func BindWithApiRequest(r *http.Request, req HttpApiMethod) error {
	var (
		b           binding.Binding
		contentType = r.Header.Get("Content-Type")
		logger      = logx.WithContext(r.Context())
	)

	if r.Method == "GET" {
		b = binding.Form
	} else {
		var stripContentTypeParam = func(contentType string) string {
			i := strings.Index(contentType, ";")
			if i != -1 {
				contentType = contentType[:i]
			}
			return contentType
		}

		contentType = stripContentTypeParam(contentType)
		switch contentType {
		case binding.MIMEJSON:
			// b = binding.FASTJSON
			b = binding.JSON
		case binding.MIMEMultipartPOSTForm:
			b = binding.FormMultipart
		case binding.MIMEPOSTForm:
			b = binding.FormPost
		default:
			return fmt.Errorf("not support contentType: %s", contentType)
		}
	}

	bindingReq := req.NewRequest()
	err := b.Bind(r, bindingReq)
	if err != nil {
		logger.Errorf("bind form error: %v", err)
		return err
	} else {
		bindingReqJson, _ := jsonx.MarshalToString(bindingReq)
		logger.Debugf("bindingReq(%s): %s", bindingReq.Method(), bindingReqJson)
	}

	if err = req.Decode2(r, contentType, bindingReq); err != nil {
		logger.Errorf("decode(%s) error: %v", bindingReq.Method(), err)
		return err
	}
	reqJson, _ := jsonx.MarshalToString(req)
	logger.Infof("req(%s): %s", bindingReq.Method(), reqJson)

	return nil
}

func GetUploadFile(c *http.Request, key string) (fileName string, file []byte, err error) {
	file2, fileHeader, err2 := c.FormFile(key)
	if err2 != nil {
		err = err2
		return
	}
	defer file2.Close()

	buf := new(bytes.Buffer)
	if _, err = io.Copy(buf, file2); err != nil {
		// log.Errorf("upload.%s.file.illegal,err::%v", key, err.Error())
		return
	}

	fileName = fileHeader.Filename
	file = buf.Bytes()

	return
}
