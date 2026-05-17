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
	"net/http"
	"reflect"
	"runtime/debug"
	"server/api/Agent"
	"server/api/UserApi"
	// "server/api/Admin" // [已迁移到新架构]
	// "server/api/WebApi" // [已迁移到新架构]
	"server/api/middleware"
	"server/core/dist/VueAgent"
	"server/global"
	"server/new/app/router"
	mid2 "server/new/app/router/middleware"
	"server/structs/Http/response"
	DB "server/structs/db"
	"strings"
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
	RouterAgent(PublicGroup)   //注册基础功能路由 不做鉴权  初始化token  获取验证码等等
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
// Agent路由 menu 需要鉴权  menu

func RouterAgent(Router *gin.RouterGroup) *gin.RouterGroup {

	局_代理入口 := global.GVA_Viper.GetString("代理入口")
	//客户经常输入错误,单独注册个路由,跳转正确地址
	if strings.ToLower(局_代理入口) != 局_代理入口 {
		Router.GET(strings.ToLower(局_代理入口), func(context *gin.Context) {
			context.Redirect(http.StatusFound, "/"+局_代理入口)
		})
	}

	Router根Agent := Router.Group(局_代理入口) //127.0.0.1:18080/  这个后面第一个不需要 / 符号
	//Router根Agent.Use(middleware.AA())
	//Router根Agent.Use(middleware.IsAgentHost()) //已删除,因为现在支持自定义入口地址了效果更好,这个,精简掉,

	Router根Agent.Use(middleware.IsAgent是否关闭())

	//打包静态VueAgent文件============================
	html := VueAgent.NewHtmlHandler()
	Router根Agent.GET("/", html.Index)
	Router根Agent.GET("/assets/*filepath", func(c *gin.Context) {
		c.FileFromFS("assets/"+c.Param("filepath"), http.FS(VueAgent.Assets))
	})
	// 解决刷新404问题
	//Router.NoRoute(html.RedirectIndex)
	//结束==============================================================

	//Agent基础路由 无数据库  无鉴权 就可以访问
	baseRouter := Router根Agent.Group("/Base")
	{
		BaseApi := Agent.Api.Base
		baseRouter.POST("Captcha", BaseApi.Captcha)
		baseRouter.POST("Login", BaseApi.Login)
	}

	//菜单列表==============================================
	baseRouter = Router根Agent.Group("/Menu")
	baseRouter.Use(middleware.IsTokenAgent()) ///鉴权中间件 检查 token

	MenuApi := Agent.Api.Menu                            //实现路由的 具体方法位置
	baseRouter.GET("GetAgentInfo", MenuApi.GetAgentInfo) // 获取用户信息
	baseRouter.POST("OutLogin", MenuApi.OutLogin)
	baseRouter.POST("GetPayStatus", MenuApi.Q取支付通道状态)
	baseRouter.POST("GetPayStatus2", MenuApi.Q取支付通道状态2)
	// 为需要鉴权的路由单独创建子组
	baseRouter.Group("", mid2.Is代理鉴权([]int{DB.D代理功能_余额充值})).
		POST("GetPayPC", MenuApi.Y余额充值)

	baseRouter.POST("GetPayOrderStatus", MenuApi.Q取余额充值订单状态)

	if !(global.GVA_Viper.GetInt("系统模式") == 1) {
		baseRouter.POST("NewPassword", MenuApi.NewPassword) //修改当前Token密码
	}
	//卡号列表===========================================
	baseRouter = Router根Agent.Group("/Ka")
	baseRouter.Use(middleware.IsTokenAgent()) ///鉴权中间件 检查 token 检查是不是管理员令牌

	{
		App := Agent.Api.Ka                       //实现路由的 具体方法位置
		baseRouter.POST("GetList", App.GetKaList) // 获取列表
		baseRouter.Group("", mid2.Is代理鉴权([]int{DB.D代理功能_制卡})).
			POST("New", App.New)
		baseRouter.POST("InventoryNewKa", App.K库存制卡)   // 新制卡号
		baseRouter.POST("GetInfo", App.GetInfo)        // 获取详细信息
		baseRouter.POST("SetStatus", App.Set修改状态)      // 修改状态
		baseRouter.POST("SetAgentNote", App.Set修改代理备注) // 修改状态
		// 为需要鉴权的路由单独创建子组
		baseRouter.Group("", mid2.Is代理鉴权([]int{DB.D代理功能_卡号追回})).
			POST("Recover", App.Z追回卡号)

		baseRouter.Group("", mid2.Is代理鉴权([]int{DB.D代理功能_更换卡号})).
			POST("ReplaceKaName", App.G更换卡号)

		baseRouter.POST("UseKa", App.K卡号充值)
		baseRouter.POST("ChartKaRegister", App.Get卡号列表统计制卡)
		baseRouter.POST("GetKaTemplate", App.Q取卡号生成模板)
		baseRouter.POST("SetKaTemplate", App.Set修改卡号生成模板)

	}

	//卡号列表===========================================
	baseRouter = Router根Agent.Group("/App")
	baseRouter.Use(middleware.IsTokenAgent()) ///鉴权中间件 检查 token 检查是不是管理员令牌
	{
		App := Agent.Api.Ka //实现路由的 具体方法位置
		baseRouter.GET("GetAppIdNameList", App.GetAppIdNameList)
	}
	//代理账号===========================================
	baseRouter = Router根Agent.Group("/Agent")
	baseRouter.Use(middleware.IsTokenAgent()) ///鉴权中间件 检查 token 检查是不是管理员令牌

	AgentApp := Agent.Api.AgentUser                                //实现路由的 具体方法位置
	baseRouter.POST("GetAgentUserList", AgentApp.GetAgentUserList) // 获取用户列表
	baseRouter.POST("GetAgentUserInfo", AgentApp.GetAgentUserInfo) // 获取用户详细信息
	baseRouter.POST("SaveAgentUser", AgentApp.Save代理信息)            // 保存用户详细信息

	// 为需要鉴权的路由单独创建子组
	baseRouter.Group("", mid2.Is代理鉴权([]int{DB.D代理功能_发展下级代理})).
		POST("NewAgentUser", AgentApp.New代理信息) // 保存用户详细信息
	baseRouter.POST("SetAgentUserStatus", AgentApp.Set修改状态)                        // 保存用户详细信息
	baseRouter.POST("GetAgentKaClassAuthority", AgentApp.GetAgentKaClassAuthority) //取全部可制卡类和已授权卡类
	baseRouter.POST("SetAgentKaClassAuthority", AgentApp.SetAgentKaClassAuthority) //设置代理可制卡类ID
	baseRouter.Group("", mid2.Is代理鉴权([]int{DB.D代理功能_转账})).
		POST("SendRmbTOAgent", AgentApp.SendRmbTOAgent) //转账
	baseRouter.POST("ChartAgentLevel", AgentApp.Get代理组织架构图)

	//代理库存管理===========================================
	baseRouter = Router根Agent.Group("/AgentInventory")
	baseRouter.Use(middleware.IsTokenAgent()) ///鉴权中间件 检查 token 检查是不是管理员令牌

	AgentInventory := Agent.Api.AgentInventory                       //实现路由的 具体方法位置
	baseRouter.POST("GetList", AgentInventory.GetAgentInventoryList) // 获取列表
	baseRouter.POST("GetKaClassTree", AgentInventory.Get取可创建库存包列表)   // 获取列表
	baseRouter.POST("GetInfo", AgentInventory.GetAgentInventoryInfo) // 获取详细信息
	baseRouter.Group("", mid2.Is代理鉴权([]int{DB.D代理功能_制卡})).
		POST("NewBuy", AgentInventory.New库存购买) // 创建库存包
	baseRouter.POST("Send", AgentInventory.K库存发送)                     // 创建库存包
	baseRouter.POST("GetSubordinateAgent", AgentInventory.Q可发送库存下级代理) // 创建库存包
	baseRouter.POST("Withdraw", AgentInventory.K库存撤回)                 // 撤回转出的库存
	baseRouter.POST("SetEndTime", AgentInventory.K库存延期)
	baseRouter.POST("SetNote", AgentInventory.K库存修改备注)
	//余额日志===========================================
	// [已迁移到新架构 new/app/controller/agent/LogMoney.go] 路由由 new/app/router/agent/agent.go 注册
	//baseRouter = Router根Agent.Group("/LogMoney")
	//baseRouter.Use(middleware.IsTokenAgent()) ///鉴权中间件 检查 token 检查是不是管理员令牌
	//
	//{
	//	App := Agent.Api.LogMoney                       //实现路由的 具体方法位置
	//	baseRouter.POST("GetList", App.GetLogMoneyList) // 获取列表
	//}
	//库存日志===========================================
	// [已迁移到新架构 new/app/controller/agent/LogAgentInventory.go] 路由由 new/app/router/agent/agent.go 注册
	//baseRouter = Router根Agent.Group("/LogAgentInventory")
	//baseRouter.Use(middleware.IsTokenAgent()) ///鉴权中间件 检查 token 检查是不是管理员令牌
	//
	//{
	//	App := Agent.Api.LogAgentInventory                   //实现路由的 具体方法位置
	//	baseRouter.POST("GetList", App.GetLogAgentInventory) // 获取列表
	//
	//}
	//制卡日志===========================================
	// [已迁移到新架构 new/app/controller/agent/LogRegisterKa.go] 路由由 new/app/router/agent/agent.go 注册
	//baseRouter = Router根Agent.Group("/LogRegisterKa")
	//baseRouter.Use(middleware.IsTokenAgent()) ///鉴权中间件 检查 token 检查是不是管理员令牌
	//
	//{
	//	App := Agent.Api.LogRegister               //实现路由的 具体方法位置
	//	baseRouter.POST("GetList", App.GetLogList) // 获取列表
	//}
	//其他操作===========================================
	baseRouter = Router根Agent.Group("/OtherFunc")
	baseRouter.Use(middleware.IsTokenAgent()) ///鉴权中间件 检查 token 检查是不是管理员令牌
	{
		App := Agent.Api.OtherFunc //实现路由的 具体方法位置
		baseRouter.Group("", mid2.Is代理鉴权([]int{DB.D代理功能_修改用户绑定})).
			POST("SetAppUserKey", App.SetAppUserKey)
	}

	return baseRouter
}

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
