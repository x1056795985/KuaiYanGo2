package core

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
	"io"
	"log"
	"reflect"
	"runtime/debug"
	"server/api/UserApi"
	// "server/api/Admin" // [已迁移到新架构]
	// "server/api/WebApi" // [已迁移到新架构]
	"server/api/middleware"
	"server/global"
	"server/new/app/router"
	"server/structs/Http/response"
)

// InitRouters 初始化总路由
func InitRouters() *gin.Engine {

	if !(global.GVA_Viper.GetInt("系统模式") == 1056795985) {
		gin.DefaultWriter = io.Discard //禁止控制台输出
		gin.SetMode(gin.ReleaseMode)   //设置为生产模式
	}

	Router := gin.Default() //返回路由实例
	_ = InitTrans("ZH")
	// Router.Use(middleware.LoadTls())  // 如果需要使用https 请打开此中间件 然后前往 core/server.go 将启动模式 更变为 Router.RunTLS("端口","你的cre/pem文件","你的key文件")

	// 跨域，如需跨域可以打开下面的注释
	Router.Use(middleware.Cors()) // 直接放行全部跨域请求
	Router.Use(T统一恐慌恢复())         // 直接放行全部跨域请求
	//公共路由器 无需鉴权
	PublicGroup := Router.Group("")
	RouterUserApi(PublicGroup) //先注册用户路由,因为管理员应用设置需要验证码接口需要获取用户api列表
	//RouterAdmin(PublicGroup)   // [已迁移到新架构] Admin路由由 new/app/router/admin/admin.go 注册
	//RouterWebApi(PublicGroup) // [已迁移到新架构] WebApi路由由 new/app/router/webApi2/webApi.go 注册
	router.RouterInit(PublicGroup)
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
		//enT := en.New() // 英文翻译器
		// 第一个参数是备用（fallback）的语言环境
		// 后面的参数是应该支持的语言环境（支持多个）
		uni := ut.New(zhT, zhT) //也是可以的
		//uni := ut.New(enT, zhT, enT)

		// locale 通常取决于 http 请求头的 'Accept-Language'
		// 也可以使用 uni.FindTranslator(...) 传入多个locale进行查找
		global.Trans, ok = uni.GetTranslator(locale)
		if !ok {
			return fmt.Errorf("uni.GetTranslator(%s) failed", locale)
		}

		// 注册翻译器
		switch locale {
		case "en":
			err = enTranslations.RegisterDefaultTranslations(v, global.Trans)
		case "zh":
			err = zhTranslations.RegisterDefaultTranslations(v, global.Trans)
		default:
			err = zhTranslations.RegisterDefaultTranslations(v, global.Trans)
		}
		return
	}
	return
}
func T统一恐慌恢复() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				局_上报错误 := fmt.Sprintln("全局捕获错误:\n", err, "\n堆栈信息:\n", string(debug.Stack()))
				debug.PrintStack()
				log.Println("发生致命错误:", 局_上报错误)
				/*								c.JSON(http.StatusInternalServerError, gin.H{
												"error": "Internal Server Error",
											})*/
				global.Q快验.Z置新用户消息(2, 局_上报错误)
			}
		}()
		c.Next()
	}
}

// admin路由 menu 需要鉴权  menu
// [已迁移到新架构 new/app/router/admin/admin.go] 所有Admin路由由新架构注册

// Agent路由 menu 需要鉴权  menu
// RouterUserApi UserApi路由入口
func RouterUserApi(Router *gin.RouterGroup) *gin.RouterGroup {

	baseRouter := Router.Group("/Api")
	baseRouter.Use(middleware.UserApi检查数据库连接())  //检查数据库是否连接,连接后才可以使用用户Api,不然大量报错
	baseRouter.Use(middleware.UserApi无Token解密()) ///鉴权中间件 检查 token  单独优先处理
	baseRouter.Use(middleware.UserApi解密())       ///鉴权中间件 检查 token
	{
		baseRouter.POST("", UserApi.UserApi_Api不存在)
		//其余的都在中间件内分配
	}

	return baseRouter
}

// WebApi路由入口
// [已迁移到新架构 new/app/router/webApi2/webApi.go] 路由由 new/app/router/webApi2 注册
//func RouterWebApi(Router *gin.RouterGroup) *gin.RouterGroup {
//
//	//===========================================
//	baseRouter := Router.Group("/WebApi/") //WebApi不做任何加密中间件处理
//	baseRouter.Use(middleware.IsWebApiHost())
//	baseRouter.Use(middleware.IsTokenWebApi()) ///鉴权中间件 检查 token  单独优先处理
//	{
//		for 键名, 键值 := range WebApi.J集_UserAPi路由 {
//			baseRouter.GET(键名, 键值.Z指向函数)
//			baseRouter.POST(键名, 键值.Z指向函数)
//		}
//	}
//
//	return baseRouter
//}
