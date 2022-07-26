package public

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/joexu01/container-dispatcher/lib"
)

// ContextWarning 错误日志
func ContextWarning(c context.Context, dLogTag string, m map[string]interface{}) {
	v := c.Value("trace")
	traceContext, ok := v.(*lib.TraceContext)
	if !ok {
		traceContext = lib.NewTrace()
	}
	lib.Log.TagWarn(traceContext, dLogTag, m)
}

// ContextError 错误日志
func ContextError(c context.Context, dLogTag string, m map[string]interface{}) {
	v := c.Value("trace")
	traceContext, ok := v.(*lib.TraceContext)
	if !ok {
		traceContext = lib.NewTrace()
	}
	lib.Log.TagError(traceContext, dLogTag, m)
}

// ContextNotice 普通日志
func ContextNotice(c context.Context, dLogTag string, m map[string]interface{}) {
	v := c.Value("trace")
	traceContext, ok := v.(*lib.TraceContext)
	if !ok {
		traceContext = lib.NewTrace()
	}
	lib.Log.TagInfo(traceContext, dLogTag, m)
}

// CommonLogWarning 错误日志
func CommonLogWarning(c *gin.Context, dLogTag string, m map[string]interface{}) {
	traceContext := GetGinTraceContext(c)
	lib.Log.TagError(traceContext, dLogTag, m)
}

// CommonLogNotice 普通日志
func CommonLogNotice(c *gin.Context, dLogTag string, m map[string]interface{}) {
	traceContext := GetGinTraceContext(c)
	lib.Log.TagInfo(traceContext, dLogTag, m)
}

// GetGinTraceContext 从gin的Context中获取数据
func GetGinTraceContext(c *gin.Context) *lib.TraceContext {
	// 防御
	if c == nil {
		return lib.NewTrace()
	}
	traceContext, exists := c.Get("trace")
	if exists {
		if tc, ok := traceContext.(*lib.TraceContext); ok {
			return tc
		}
	}
	return lib.NewTrace()
}

// GetTraceContext 从Context中获取数据
func GetTraceContext(c context.Context) *lib.TraceContext {
	if c == nil {
		return lib.NewTrace()
	}
	traceContext := c.Value("trace")
	if tc, ok := traceContext.(*lib.TraceContext); ok {
		return tc
	}
	return lib.NewTrace()
}
