package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/joexu01/container-dispatcher/public"

	en_translations "github.com/go-playground/validator/v10/translations/en"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"

	"reflect"
)

// 使用此中间件时必须搭配 GetValidParamsI18n() 方法来验证

func TranslationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//参照：https://github.com/go-playground/validator/blob/v9/_examples/translations/main.go

		//设置支持语言
		enLang := en.New()
		zhLang := zh.New()

		//设置国际化翻译器
		uni := ut.New(zhLang, zhLang, enLang)
		validate := validator.New()

		//根据参数取翻译器实例
		locale := c.DefaultQuery("locale", "zh")
		trans, _ := uni.GetTranslator(locale)

		//翻译器注册到validator
		switch locale {
		case "en":
			_ = en_translations.RegisterDefaultTranslations(validate, trans)
			validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
				return fld.Tag.Get("en_comment")
			})
			break
		default:
			_ = zh_translations.RegisterDefaultTranslations(validate, trans)
			validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
				return fld.Tag.Get("comment")
			})

			//自定义验证方法
			//https://github.com/go-playground/validator/blob/master/_examples/custom-validation/main.go
			_ = validate.RegisterValidation("is-valid-user", func(fl validator.FieldLevel) bool {
				return fl.Field().String() == "admin"
			})

			//自定义验证器
			//https://github.com/go-playground/validator/blob/master/_examples/translations/main.go
			_ = validate.RegisterTranslation("is-valid-user", trans, func(ut ut.Translator) error {
				return ut.Add("is-valid-user", "{0} 非有效用户", true)
			}, func(ut ut.Translator, fe validator.FieldError) string {
				t, _ := ut.T("is-valid-user", fe.Field())
				return t
			})
			break
		}
		c.Set(public.TranslatorKey, trans)
		c.Set(public.ValidatorKey, validate)
		c.Next()
	}
}
