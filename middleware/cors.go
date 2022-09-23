package middleware

import (
	"github.com/gin-gonic/gin"
	"log"
)

func CORSMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		//method := ctx.Request.Method

		// set response header
		c.Header("Access-Control-Allow-Origin", "http://127.0.0.1:5173")
		log.Println("Origin:", c.Request.Header.Get("Origin"))
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers",
			"Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")

		// 默认过滤options和head这两个请求，使用204返回
		//if method == http.MethodOptions || method == http.MethodHead {
		//	ctx.AbortWithStatus(http.StatusNoContent)
		//	return
		//}

		c.Next()

	}
}
