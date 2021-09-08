package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	"github.com/go-playground/locales/zh_Hant_TW"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

func Translations() gin.HandlerFunc {
	return func(c *gin.Context) {
		if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
			uni := ut.New(en.New(), zh.New(), zh_Hant_TW.New())
			locale := c.GetHeader("locale")
			trans, _ := uni.GetTranslator(locale)

			switch locale {
			case "zh":
				_ = zh_translations.RegisterDefaultTranslations(v, trans)
			case "en":
				_ = en_translations.RegisterDefaultTranslations(v, trans)
			default:
				_ = en_translations.RegisterDefaultTranslations(v, trans)
			}
			c.Set("trans", trans)
		}

		c.Next()
	}
}