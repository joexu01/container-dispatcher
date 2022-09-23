package middleware

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joexu01/container-dispatcher/lib"
	"net/http"
	"strings"
)

type ResponseCode int

// 1000以下为通用码，1000以上为用户自定义码
const (
	SuccessCode ResponseCode = iota
	UndefErrorCode
	ValidErrorCode
	InternalErrorCode

	CustomizeCode ResponseCode = 1000
)

type Response struct {
	ErrorCode ResponseCode `json:"errno"`
	ErrorMsg  string       `json:"err_msg"`
	Data      interface{}  `json:"data"`
	TraceId   interface{}  `json:"trace_id"`
	Stack     interface{}  `json:"stack"`
}

func ResponseWithCode(c *gin.Context, status int, errorCode ResponseCode, err error, data interface{}) {
	trace, _ := c.Get("trace")
	traceContext, _ := trace.(*lib.TraceContext)
	traceId := ""
	if traceContext != nil {
		traceId = traceContext.TraceId
	}
	stack := ""
	if status != http.StatusOK {
		if c.Query("is_debug") == "1" || lib.GetConfEnv() == "dev" {
			stack = strings.Replace(fmt.Sprintf("%+v", err), err.Error()+"\n", "", -1)
		}
	}

	resp := &Response{
		ErrorCode: errorCode,
		ErrorMsg:  err.Error(),
		Data:      data,
		TraceId:   traceId,
		Stack:     stack,
	}
	c.JSON(status, resp)
	response, _ := json.Marshal(resp)
	c.Set("response", string(response))
	//if status != http.StatusOK {
	//	_ = c.AbortWithError(status, err)
	//}
}

func ResponseSuccess(c *gin.Context, data interface{}) {
	trace, _ := c.Get("trace")
	traceContext, _ := trace.(*lib.TraceContext)
	traceId := ""
	if traceContext != nil {
		traceId = traceContext.TraceId
	}

	resp := &Response{ErrorCode: SuccessCode, ErrorMsg: "", Data: data, TraceId: traceId}
	c.JSON(200, resp)
	response, _ := json.Marshal(resp)
	c.Set("response", string(response))
}

func ResponseError(c *gin.Context, code ResponseCode, err error) {
	trace, _ := c.Get("trace")
	traceContext, _ := trace.(*lib.TraceContext)
	traceId := ""
	if traceContext != nil {
		traceId = traceContext.TraceId
	}

	stack := ""
	if c.Query("is_debug") == "1" || lib.GetConfEnv() == "dev" {
		stack = strings.Replace(fmt.Sprintf("%+v", err), err.Error()+"\n", "", -1)
	}

	resp := &Response{ErrorCode: code, ErrorMsg: err.Error(), Data: "", TraceId: traceId, Stack: stack}
	c.JSON(200, resp)
	response, _ := json.Marshal(resp)
	c.Set("response", string(response))
	_ = c.AbortWithError(http.StatusInternalServerError, err)
}
