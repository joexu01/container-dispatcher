package middleware

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joexu01/container-dispatcher/lib"
	"github.com/joexu01/container-dispatcher/public"
	"net/http"
	"runtime/debug"
	"strings"
)

func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				//先做一下日志记录
				stack := strings.Replace(string(debug.Stack()), `\n`, "\n", -1)

				fmt.Println(stack)
				public.CommonLogNotice(c, "_com_panic", map[string]interface{}{
					"error": fmt.Sprint(err),
					"stack": stack,
				})

				if lib.ConfBase.DebugMode == "debug" {
					ResponseWithCode(c, http.StatusInternalServerError, 2000, errors.New("internal error"), "")
					return
				} else {
					ResponseWithCode(c, http.StatusInternalServerError, 2000, errors.New(fmt.Sprint(err)), "")
					return
				}
			}
		}()
		c.Next()
	}
}
