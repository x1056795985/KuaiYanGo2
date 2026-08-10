package router

import (
	"fmt"
	"io"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"

	"server/app/global"
	"server/app/models/old/response"
	"server/app/monitoring"
	"server/app/router/admin"
	"server/app/router/agent"
	"server/app/router/middleware"
	userSafetyApi2 "server/app/router/userSafetyApi"
	webApi2 "server/app/router/webApi2"
	"server/app/router/webSocket"
	"server/app/router/webUser"
)

func InitRouters() *gin.Engine {
	if !(global.GVA_Viper.GetInt("系统模式") == 1056795985) {
		gin.DefaultWriter = io.Discard
		gin.SetMode(gin.ReleaseMode)
	}

	局_路由 := gin.Default()
	_ = InitTrans("ZH")

	局_路由.Use(middleware.Cors())
	局_路由.Use(middleware.T统一恐慌恢复())
	局_路由.Use(monitoring.Q监控.Q监控中间件())

	局_公开分组 := 局_路由.Group("")
	RouterInit(局_公开分组)

	if global.GVA_Viper.GetInt("系统模式") == 1 {
		局_路由.NoRoute(func(c *gin.Context) {
			response.FailWithMessage("演示模式不可操作，请部署到自己的服务器深度体验", c)
		})
	}

	return 局_路由
}

func InitTrans(locale string) (err error) {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := fld.Tag.Get("zh")
			if name == "" {
				name = fld.Tag.Get("json")
			}
			return name
		})

		局_中文翻译器 := zh.New()
		局_翻译器集合 := ut.New(局_中文翻译器, 局_中文翻译器)

		global.Trans, ok = 局_翻译器集合.GetTranslator(locale)
		if !ok {
			return fmt.Errorf("uni.GetTranslator(%s) failed", locale)
		}

		switch locale {
		case "en":
			err = zhTranslations.RegisterDefaultTranslations(v, global.Trans)
		case "zh":
			err = zhTranslations.RegisterDefaultTranslations(v, global.Trans)
		default:
			err = zhTranslations.RegisterDefaultTranslations(v, global.Trans)
		}
		return
	}
	return
}

func RouterInit(routerGroup *gin.RouterGroup) *gin.RouterGroup {
	局_路由分组 := routerGroup

	局_用户安全接口路由 := userSafetyApi2.AllRouter{}
	局_用户安全接口路由.InitWebApiRouter(局_路由分组)

	局_后台路由 := admin.AllRouter{}
	局_后台路由.InitAdminRouter(局_路由分组)

	局_代理路由 := agent.AllRouter{}
	局_代理路由.InitAgentRouter(局_路由分组)

	局_WebApi路由 := webApi2.AllRouter{}
	局_WebApi路由.InitWebApiRouter(局_路由分组)

	局_WebUser路由 := webUser.AllRouter{}
	局_WebUser路由.InitWebUserRouter(局_路由分组)

	局_WebSocket路由 := webSocket.AllRouter{}
	局_WebSocket路由.InitWebSocketRouter(局_路由分组)

	return routerGroup
}
