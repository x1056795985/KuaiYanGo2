package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
	"io"
	"reflect"
	"server/global"
	"server/new/app/router/admin"
	"server/new/app/router/agent"
	"server/new/app/router/middleware"
	userSafetyApi2 "server/new/app/router/userSafetyApi"
	"server/new/app/router/webApi2"
	"server/new/app/router/webSocket"
	"server/new/app/router/webUser"
	"server/structs/Http/response"
)

// 初始化总路由
func InitRouters() *gin.Engine {

	if !(global.GVA_Viper.GetInt("系统模式") == 1056795985) {
		gin.DefaultWriter = io.Discard //禁止控制台输出
		gin.SetMode(gin.ReleaseMode)   //设置为生产模式
	}

	Router := gin.Default() //返回路由实例
	_ = InitTrans("ZH")
	// 跨域，如需跨域可以打开下面的注释
	Router.Use(middleware.Cors())    // 直接放行全部跨域请求
	Router.Use(middleware.T统一恐慌恢复()) // 全局恐慌恢复中间件
	//公共路由器 无需鉴权
	PublicGroup := Router.Group("")

	RouterInit(PublicGroup)
	if global.GVA_Viper.GetInt("系统模式") == 1 {
		Router.NoRoute(func(c *gin.Context) {
			response.FailWithMessage("演示模式不可操作,请部署到自己服务器深度体验", c)
			return
		})
	}

	//global.GVA_LOG.Info("router register success(路由注册成功)")
	return Router
}

// InitTrans 初始化控制器翻译器
func InitTrans(locale string) (err error) {
	// 修改gin框架中的Validator引擎属性，实现自定制
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {

		// 注册一个获取json tag的自定义方法 将字段名改为中文,
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := fld.Tag.Get("zh") //如果有这个信息,就是用这个
			if name == "" {
				name = fld.Tag.Get("json") //没有就用json
			}
			return name
		})

		zhT := zh.New() // 中文翻译器
		// 第一个参数是备用（fallback）的语言环境
		// 后面的参数是应该支持的语言环境（支持多个）
		uni := ut.New(zhT, zhT) //也是可以的

		// locale 通常取决于 http 请求头的 'Accept-Language'
		// 也可以使用 uni.FindTranslator(...) 传入多个locale进行查找
		global.Trans, ok = uni.GetTranslator(locale)
		if !ok {
			return fmt.Errorf("uni.GetTranslator(%s) failed", locale)
		}

		// 注册翻译器
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

func RouterInit(gin *gin.RouterGroup) *gin.RouterGroup {

	Router := gin //返回路由实例

	routerUserSafetyApi2 := userSafetyApi2.AllRouter{}
	routerUserSafetyApi2.InitWebApiRouter(Router) //先注册用户路由,因为管理员应用设置需要验证码接口需要获取用户api列表

	router := admin.AllRouter{}
	router.InitAdminRouter(Router) //初始化管理员路由

	routerAgent := agent.AllRouter{}
	routerAgent.InitAgentRouter(Router) //初始化Agent路由

	routerWebApi := webApi2.AllRouter{}
	routerWebApi.InitWebApiRouter(Router) //初始化WEBAPi路由

	routerWebUser := webUser.AllRouter{}
	routerWebUser.InitWebUserRouter(Router) //初始化WEBUser路由

	routerWebSocket := webSocket.AllRouter{}
	routerWebSocket.InitWebSocketRouter(Router) //初始化WebSocket路由
	return gin
}
