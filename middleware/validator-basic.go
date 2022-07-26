package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/joexu01/container-dispatcher/public"
)

// 使用此中间件时必须搭配 GetValidParamsDefault() 方法来验证

func ValidatorBasicMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		validate := validator.New()

		//自定义验证方法
		//https://github.com/go-playground/validator/blob/master/_examples/custom-validation/main.go

		//自定义验证器
		//https://github.com/go-playground/validator/blob/master/_examples/translations/main.go

		var validateUsername validator.Func = func(fl validator.FieldLevel) bool {
			if username := fl.Field().String(); len(username) > 5 {
				return true
			}
			return false
		}

		_ = validate.RegisterValidation("validate_username", validateUsername)

		c.Set(public.ValidatorKey, validate)
		c.Next()
	}
}
