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
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"runtime/debug"
	"server/api/Admin"
	"server/api/Agent"
	"server/api/UserApi"
	"server/api/WebApi"
	"server/api/middleware"
	"server/core/dist/VueAdmin"
	VueAdminAssets "server/core/dist/VueAdmin/assets"
	"server/core/dist/VueAgent"
	VueAgentAssets "server/core/dist/VueAgent/assets"
	"server/global"
	"server/new/app/router"
	"server/structs/Http/response"
)

// InitRouters 初始化总路由
func InitRouters() *gin.Engine {
	//gin.SetMode(gin.ReleaseMode) //设置为生产模式

	if !(global.GVA_Viper.GetInt("系统模式") == 1056795985) {
		gin.DefaultWriter = ioutil.Discard //禁止控制台输出

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
	RouterAdmin(PublicGroup)   //注册基础功能路由 不做鉴权  初始化token  获取验证码等等
	RouterAgent(PublicGroup)   //注册基础功能路由 不做鉴权  初始化token  获取验证码等等
	RouterWebApi(PublicGroup)
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
func RouterAdmin(Router *gin.RouterGroup) *gin.RouterGroup {

	Router.GET("admin", func(context *gin.Context) { //客户经常输入错误,单独注册个路由,跳转正确地址
		context.Redirect(http.StatusFound, "/Admin")
	})
	Router根Admin := Router.Group("Admin") //127.0.0.1:18080/  这个后面第一个不需要 / 符号
	Router根Admin.Use(middleware.IsAdminHost())

	//打包静态VueAdmin文件============================
	html := VueAdmin.NewHtmlHandler()
	Router根Admin.StaticFS("/assets", http.FS(VueAdminAssets.Assets))
	Router根Admin.GET("/", html.Index)

	// 解决刷新404问题
	//Router.NoRoute(html.RedirectIndex)
	//结束==============================================================

	// admin基础路由 无数据库  无鉴权 就可以访问
	baseRouter := Router根Admin.Group("/base")
	{
		base := Api.Admin.Base
		baseRouter.POST("Captcha", base.Captcha)
		baseRouter.POST("Login", base.Login)
		/*		baseRouter.POST("SetTableWidth", base.Table宽度保存)
				baseRouter.POST("GetTableWidth", base.Table宽度读取)*/

		initDB := Api.Admin.InitDb                 //实现路由的 具体方法位置
		baseRouter.POST("InitDB", initDB.InitDB)   // 初始化数据库
		baseRouter.POST("CheckDB", initDB.CheckDB) // 检测是否需要初始化数据库
	}

	baseRouter = Router根Admin.Group("/User")
	baseRouter.Use(middleware.IsTokenAdmin()) ///鉴权中间件 检查 token 检查是不是管理员令牌

	menu := Api.Admin.Menu                            //实现路由的 具体方法位置
	LinkUser := Api.Admin.LinkUserApi                 //实现路由的 具体方法位置
	User := Api.Admin.User                            //实现路由的 具体方法位置
	baseRouter.POST("GetMenu", menu.GetMenu)          // 初始化菜单信息
	baseRouter.GET("GetAdminInfo", User.GetAdminInfo) // 获取用户信息
	baseRouter.POST("OutLogin", User.OutLogin)

	if !(global.GVA_Viper.GetInt("系统模式") == 1) {
		baseRouter.POST("AdminNewPassword", User.AdminNewPassword) //修改当前Token密码
	}

	//在线列表==============================================
	baseRouter.POST("GetLinkUserList", LinkUser.GetLinkUserList) // 获取在线列表
	baseRouter.POST("NewWebApiToken", LinkUser.NewWebApiToken)   // 获取在线列表
	baseRouter.POST("SetTokenOutTime", LinkUser.SetTokenOutTime) // 获取在线列表
	if !(global.GVA_Viper.GetInt("系统模式") == 1) {
		baseRouter.POST("logout", LinkUser.Del批量注销)          // 批量注销在线
		baseRouter.POST("DeleteLogout", LinkUser.Del批量删除已注销) // 批量删除已注销
	}
	//用户账号===========================================
	baseRouter.POST("GetUserList", User.GetUserList) // 获取用户列表
	baseRouter.POST("GetUserInfo", User.GetUserInfo) // 获取用户详细信息
	baseRouter.POST("SaveUser", User.Save用户信息)       // 保存用户详细信息
	baseRouter.POST("NewUser", User.New用户信息)         // 保存用户详细信息
	baseRouter.POST("SetUserStatus", User.Set修改状态)   // 保存用户详细信息
	baseRouter.POST("SetBatchAddRMB", User.P批量_增减余额) // 保存用户详细信息
	if !(global.GVA_Viper.GetInt("系统模式") == 1) {
		baseRouter.POST("DeleteUser", User.Del批量删除用户) // 获取用户详细信息
	}
	//代理账号===========================================
	baseRouter = Router根Admin.Group("/Agent")
	baseRouter.Use(middleware.IsTokenAdmin()) ///鉴权中间件 检查 token 检查是不是管理员令牌

	AgentApp := Api.Admin.AgentUser                                                //实现路由的 具体方法位置
	baseRouter.POST("GetAgentUserList", AgentApp.GetAgentUserList)                 // 获取用户列表
	baseRouter.POST("GetAgentUserInfo", AgentApp.GetAgentUserInfo)                 // 获取用户详细信息
	baseRouter.POST("SaveAgentUser", AgentApp.Save代理信息)                            // 保存用户详细信息
	baseRouter.POST("NewAgentUser", AgentApp.New代理信息)                              // 保存用户详细信息
	baseRouter.POST("SetAgentUserStatus", AgentApp.Set修改状态)                        // 保存用户详细信息
	baseRouter.POST("GetAgentKaClassAuthority", AgentApp.GetAgentKaClassAuthority) //取全部可制卡类和已授权卡类
	baseRouter.POST("SetAgentKaClassAuthority", AgentApp.SetAgentKaClassAuthority) //设置代理可制卡类ID
	if !(global.GVA_Viper.GetInt("系统模式") == 1) {
		baseRouter.POST("DeleteAgentUser", AgentApp.Del批量删除代理) // 获取用户详细信息
	}
	//代理库存管理===========================================
	baseRouter = Router根Admin.Group("/AgentInventory")
	baseRouter.Use(middleware.IsTokenAdmin()) ///鉴权中间件 检查 token 检查是不是管理员令牌

	AgentInventory := Api.Admin.AgentInventory                                        //实现路由的 具体方法位置
	baseRouter.POST("GetList", AgentInventory.GetAgentInventoryList)                  // 获取列表
	baseRouter.POST("GetAgentTreeAndKaClassTree", AgentInventory.Get取下级代理列表和可创建库存包列表) // 获取列表
	baseRouter.POST("GetInfo", AgentInventory.GetAgentInventoryInfo)                  // 获取详细信息
	baseRouter.POST("New", AgentInventory.New库存包信息)                                   // 创建库存包
	baseRouter.POST("Withdraw", AgentInventory.K库存撤回)                                 // 撤回转出的库存
	if !(global.GVA_Viper.GetInt("系统模式") == 1) {
		baseRouter.POST("Delete", AgentInventory.Del批量删除库存)
	}
	//应用管理===========================================
	baseRouter = Router根Admin.Group("/App")
	baseRouter.Use(middleware.IsTokenAdmin()) ///鉴权中间件 检查 token 检查是不是管理员令牌

	{
		App := Api.Admin.App                                     //实现路由的 具体方法位置
		baseRouter.POST("GetList", App.GetAppList)               // 获取列表
		baseRouter.POST("New", App.NewApp信息)                     // 新建信息
		baseRouter.POST("GetInfo", App.GetAppInfo)               // 获取详细信息
		baseRouter.GET("GetAppIdNameList", App.GetAppIdNameList) // 取appid和名字数组
		baseRouter.GET("GetAllUserApi", App.Get全部用户APi)          // Get全部用户APi
		baseRouter.GET("GetAllWebApi", App.Get全部WebAPi)          // Get全部用户APi
		baseRouter.GET("GetAppIdMax", App.GetAppIdMax)           // 取AppId最大值
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			baseRouter.POST("Delete", App.Del批量删除App)  // 删除信息
			baseRouter.POST("SaveInfo", App.SaveApp信息) // 保存详细信息
		}
	}
	//软件用户===========================================
	baseRouter = Router根Admin.Group("/AppUser")
	baseRouter.Use(middleware.IsTokenAdmin()) ///鉴权中间件 检查 token 检查是不是管理员令牌

	{
		App := Api.Admin.AppUser //实现路由的 具体方法位置

		baseRouter.POST("GetList", App.GetAppUserList) // 获取列表
		baseRouter.POST("New", App.New用户信息)            // 新建信息
		baseRouter.POST("GetInfo", App.GetAppUserInfo) // 获取详细信息
		baseRouter.POST("SaveInfo", App.Save用户信息)      // 保存详细信息
		baseRouter.POST("SetStatus", App.Set修改状态)      // 修改状态
		baseRouter.POST("SetBatchAddVipTime", App.Set批量维护_增减时间点数)
		baseRouter.POST("SetBatchAddVipNumber", App.Set批量维护_增减积分)
		baseRouter.POST("SetBatchSetUserConfig", App.Set批量维护_置云配置)
		baseRouter.POST("SetBatchUserClass", App.Set批量维护_修改用户类型)
		baseRouter.POST("SetBatchAllUserVipTime", App.P批量_全部用户增减时间点数)

		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			baseRouter.POST("DeleteBatch", App.Set批量维护_删除用户)
			baseRouter.POST("Delete", App.Del批量删除软件用户) // 删除信息
		}

	}
	//软件用户类型===========================================
	baseRouter = Router根Admin.Group("/UserClass")
	baseRouter.Use(middleware.IsTokenAdmin()) ///鉴权中间件 检查 token 检查是不是管理员令牌

	{
		App := Api.Admin.UserClass                          //实现路由的 具体方法位置
		baseRouter.POST("GetList", App.GetUserClassList)    // 获取列表
		baseRouter.POST("New", App.NewUserClass信息)          // 新建信息
		baseRouter.POST("GetInfo", App.GetUserClassInfo)    // 获取详细信息
		baseRouter.POST("SaveInfo", App.SetUserClass信息)     // 保存详细信息
		baseRouter.POST("GetIdNameList", App.GetIdNameList) // 取id和名字数组
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			baseRouter.POST("Delete", App.Del批量删除用户类型) // 删除信息
		}
	}
	//卡类列表===========================================
	baseRouter = Router根Admin.Group("/KaClass")
	baseRouter.Use(middleware.IsTokenAdmin()) ///鉴权中间件 检查 token 检查是不是管理员令牌

	{
		App := Api.Admin.KaClass                       //实现路由的 具体方法位置
		baseRouter.POST("GetList", App.GetKaClassList) // 获取列表
		baseRouter.POST("New", App.New)                // 新建信息
		baseRouter.POST("GetInfo", App.GetInfo)        // 获取详细信息
		baseRouter.POST("SaveInfo", App.SaveInfo)      // 保存详细信息
		//baseRouter.GET("GetIdNameList", App.GetIdNameList) // 取id和名字数组
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			baseRouter.POST("Delete", App.Delete) // 删除信息
		}
	}

	//卡号列表===========================================
	baseRouter = Router根Admin.Group("/Ka")
	baseRouter.Use(middleware.IsTokenAdmin()) ///鉴权中间件 检查 token 检查是不是管理员令牌

	{
		App := Api.Admin.Ka                                   //实现路由的 具体方法位置
		baseRouter.POST("GetList", App.GetKaList)             // 获取列表
		baseRouter.POST("New", App.New)                       // 新制卡号
		baseRouter.POST("batchKaNameNew", App.BatchKaNameNew) // 新制卡号,指定卡号
		baseRouter.POST("GetInfo", App.GetInfo)               // 获取详细信息
		baseRouter.POST("SaveInfo", App.SaveKa信息)             // 保存详细信息

		baseRouter.POST("SetStatus", App.Set修改状态) // 修改状态
		baseRouter.POST("SetAdminNote", App.Set修改管理员备注)
		baseRouter.POST("GetKaTemplate", App.Q取卡号生成模板)
		baseRouter.POST("SetKaTemplate", App.Set修改卡号生成模板)
		baseRouter.POST("Recover", App.Z追回卡号)
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			baseRouter.POST("Delete", App.Delete)            // 删除信息
			baseRouter.POST("DeleteBatch", App.Set批量维护_删除用户) // 删除信息

		}
	}
	//用户云配置===========================================
	baseRouter = Router根Admin.Group("/UserConfig")
	baseRouter.Use(middleware.IsTokenAdmin()) ///鉴权中间件 检查 token 检查是不是管理员令牌
	{
		App := Api.Admin.UserConfig                         //实现路由的 具体方法位置
		baseRouter.POST("GetList", App.GetList)             // 获取列表
		baseRouter.POST("New", App.New)                     // 新建信息
		baseRouter.POST("GetInfo", App.GetInfo)             // 获取详细信息
		baseRouter.POST("SetUserConfig", App.SetUserConfig) // 保存详细信息

		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			baseRouter.POST("Delete", App.Delete) // 删除信息
		}
	}
	//公共变量===========================================
	baseRouter = Router根Admin.Group("/PublicData")
	baseRouter.Use(middleware.IsTokenAdmin()) ///鉴权中间件 检查 token 检查是不是管理员令牌

	{
		App := Api.Admin.PublicData                          //实现路由的 具体方法位置
		baseRouter.POST("GetList", App.GetPublicDataList)    // 获取列表
		baseRouter.POST("New", App.New)                      // 新建信息
		baseRouter.POST("GetInfo", App.GetInfo)              // 获取详细信息
		baseRouter.POST("SaveInfo", App.SaveDB_PublicData信息) // 保存详细信息

		baseRouter.POST("SetIsVip", App.Set修改vip限制)
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			baseRouter.POST("Delete", App.Delete) // 删除信息
		}
	}
	//公共函数===========================================
	baseRouter = Router根Admin.Group("/PublicJs")
	baseRouter.Use(middleware.IsTokenAdmin()) ///鉴权中间件 检查 token 检查是不是管理员令牌

	{
		App := Api.Admin.PublicJs                       //实现路由的 具体方法位置
		baseRouter.POST("GetList", App.GetPublicJsList) // 获取列表
		baseRouter.POST("GetInfo", App.GetInfo)         // 获取详细信息
		baseRouter.POST("SetIsVip", App.Set修改vip限制)     // 删除信息
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			baseRouter.POST("New", App.New)                    // 新建信息
			baseRouter.POST("TestRunJs", App.C测试执行)            // 新建信息
			baseRouter.POST("SaveInfo", App.SaveDB_PublicJs信息) // 保存详细信息
			baseRouter.POST("Delete", App.Delete)              // 删除信息
		}
	}
	//任务池===========================================
	baseRouter = Router根Admin.Group("/TaskPool")
	baseRouter.Use(middleware.IsTokenAdmin()) ///鉴权中间件 检查 token 检查是不是管理员令牌

	{
		App := Api.Admin.TaskPool                //实现路由的 具体方法位置
		baseRouter.POST("GetList", App.GetList2) // 获取列表
		baseRouter.POST("New", App.New)          // 新建信息
		baseRouter.POST("GetInfo", App.GetInfo)  // 获取详细信息
		baseRouter.POST("SaveInfo", App.Save)    // 保存详细信息
		baseRouter.POST("SetStatus", App.Set修改状态)
		baseRouter.POST("DeleteTaskQueueTid", App.Q清空队列)
		baseRouter.POST("UuidAddQueue", App.Uuid重新加入队列)

		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			baseRouter.POST("Delete", App.Del批量删除) // 删除信息
		}
	}
	//系统设置===========================================
	baseRouter = Router根Admin.Group("/SetSystem")
	baseRouter.Use(middleware.IsTokenAdmin()) ///鉴权中间件 检查 token 检查是不是管理员令牌

	{
		App := Api.Admin.SetSystem //实现路由的 具体方法位置
		baseRouter.POST("GetInfoSystem", App.GetInfoSystem)

		baseRouter.POST("GenerateAPIEncryptedSDK", App.S生成API加密源码SDK)
		//在线支付
		baseRouter.POST("GetInfoPay", App.GetInfo在线支付) // 获取详细信息
		///mqtt配置
		baseRouter.POST("GetInfoMQTT", App.GetInfoMQTT配置)
		baseRouter.POST("SaveInfoMQTT", App.SaveMQTT配置)
		baseRouter.POST("mqttSendMsg", App.Mqtt发送测试)

		//短信平台配置
		baseRouter.POST("GetInfoSMS", App.GetInfo短信平台设置)
		baseRouter.POST("SaveInfoSMS", App.Save短信平台设置)
		baseRouter.POST("TestSendSMS", App.F发送短信平台测试)
		baseRouter.POST("GetInfoCaptcha2", App.GetInfo行为验证码平台设置)
		baseRouter.POST("SaveInfoCaptcha2", App.Save行为验证码平台设置)
		baseRouter.POST("GetInfoCloudStorage", App.GetInfo云存储设置)
		baseRouter.POST("SaveInfoCloudStorage", App.Save云存储设置)
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			baseRouter.POST("SaveInfoSystem", App.Save信息System)
			baseRouter.POST("SaveInfoPay", App.Save信息在线支付) // 保存详细信息
		}
	}
	//登录日志===========================================
	baseRouter = Router根Admin.Group("/LogLogin")
	baseRouter.Use(middleware.IsTokenAdmin()) ///鉴权中间件 检查 token 检查是不是管理员令牌

	{
		App := Api.Admin.LogLogin                       //实现路由的 具体方法位置
		baseRouter.POST("GetList", App.GetLogLoginList) // 获取列表
		baseRouter.POST("GetInfo", App.GetInfo)         // 获取详细信息
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			baseRouter.POST("Delete", App.Delete) // 删除信息
		}

	}
	//余额日志===========================================
	baseRouter = Router根Admin.Group("/LogMoney")
	baseRouter.Use(middleware.IsTokenAdmin()) ///鉴权中间件 检查 token 检查是不是管理员令牌

	{
		App := Api.Admin.LogMoney                       //实现路由的 具体方法位置
		baseRouter.POST("GetList", App.GetLogMoneyList) // 获取列表
		baseRouter.POST("GetInfo", App.GetInfo)         // 获取详细信息

		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			baseRouter.POST("Delete", App.Delete) // 删除信息
		}
	}
	//库存日志===========================================
	baseRouter = Router根Admin.Group("/LogAgentInventory")
	baseRouter.Use(middleware.IsTokenAdmin()) ///鉴权中间件 检查 token 检查是不是管理员令牌

	{
		App := Api.Admin.LogAgentInventory                   //实现路由的 具体方法位置
		baseRouter.POST("GetList", App.GetLogAgentInventory) // 获取列表
		baseRouter.POST("GetInfo", App.GetInfo)              // 获取详细信息

		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			baseRouter.POST("Delete", App.Delete) // 删除信息
		}
	}

	//积分点数日志===========================================
	baseRouter = Router根Admin.Group("/LogVipNumber")
	baseRouter.Use(middleware.IsTokenAdmin()) ///鉴权中间件 检查 token 检查是不是管理员令牌

	{
		App := Api.Admin.LogVipNumber              //实现路由的 具体方法位置
		baseRouter.POST("GetList", App.GetLogList) // 获取列表
		baseRouter.POST("GetInfo", App.GetInfo)    // 获取详细信息
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			baseRouter.POST("Delete", App.Delete) // 删除信息
		}

	}
	//制卡日志===========================================
	baseRouter = Router根Admin.Group("/LogRegisterKa")
	baseRouter.Use(middleware.IsTokenAdmin()) ///鉴权中间件 检查 token 检查是不是管理员令牌

	{
		App := Api.Admin.LogRegister               //实现路由的 具体方法位置
		baseRouter.POST("GetList", App.GetLogList) // 获取列表
		baseRouter.POST("GetInfo", App.GetInfo)    // 获取详细信息
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			baseRouter.POST("Delete", App.Delete) // 删除信息
		}

	}
	//代理操作日志===========================================
	baseRouter = Router根Admin.Group("/LogAgentOtherFunc")
	baseRouter.Use(middleware.IsTokenAdmin()) ///鉴权中间件 检查 token 检查是不是管理员令牌
	{
		App := Api.Admin.LogAgentOtherFunc         //实现路由的 具体方法位置
		baseRouter.POST("GetList", App.GetLogList) // 获取列表
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			baseRouter.POST("Delete", App.Delete) // 删除信息
		}

	}
	//用户消息===========================================
	baseRouter = Router根Admin.Group("/LogUserMsg")
	baseRouter.Use(middleware.IsTokenAdmin()) ///鉴权中间件 检查 token 检查是不是管理员令牌

	{
		App := Api.Admin.LogUserMsg                //实现路由的 具体方法位置
		baseRouter.POST("GetList", App.GetLogList) // 获取列表
		baseRouter.POST("GetInfo", App.GetInfo)    // 获取详细信息
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			baseRouter.POST("SetIsRead", App.Set修改IsRead) // 设置已读状态
			baseRouter.POST("Delete", App.Delete)         // 删除信息
		}
	}
	//余额充值订单===========================================
	baseRouter = Router根Admin.Group("/LogRMBPayOrder")
	baseRouter.Use(middleware.IsTokenAdmin()) ///鉴权中间件 检查 token 检查是不是管理员令牌

	{
		App := Api.Admin.LogRMBPayOrder             //实现路由的 具体方法位置
		baseRouter.POST("GetList", App.GetLogList2) // 获取列表
		baseRouter.POST("GetInfo", App.GetInfo)     // 获取详细信息
		baseRouter.POST("New", App.New手动充值)
		baseRouter.POST("SetPayOrderNote", App.Set修改备注)
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			baseRouter.POST("Delete", App.Delete) // 删除信息
			baseRouter.POST("Out", App.Out退款)     // 退款
		}
	}
	//控制面板===========================================
	baseRouter = Router根Admin.Group("/Panel")
	baseRouter.Use(middleware.IsTokenAdmin()) ///鉴权中间件 检查 token 检查是不是管理员令牌

	{
		App := Api.Admin.Panel //实现路由的 具体方法位置
		//监控也
		baseRouter.POST("getServerInfo", App.GetServerInfo) // 获取服务器信息
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			baseRouter.POST("reloadSystem", App.ReloadSystem) // 重启服务
			baseRouter.POST("StopSystem", App.StopSystem)     // 停止系统
		}

		//分析页
		baseRouter.POST("ChartLinksUser", App.Get在线统计)
		baseRouter.POST("ChartLinksUserIPCity", App.Get在线用户Ip地图分布统计)
		baseRouter.POST("ChartLinksUserLoginTime", App.Get在线用户统计登录活动时间)
		baseRouter.POST("ChartAppUserClass", App.Get应用用户类型统计)
		baseRouter.POST("ChartUser", App.Get用户账号登录注册统计)
		baseRouter.POST("ChartRMBAddSub", App.Get余额充值消费统计)
		baseRouter.POST("ChartVipNumberAddSub", App.Get积分点数消费统计)
		baseRouter.POST("ChartAppUser", App.Get应用用户统计)
		baseRouter.POST("ChartAppKa", App.Get卡号列表统计应用卡可用已用)
		baseRouter.POST("ChartAppKaClass", App.Get卡号列表统计应用卡类可用已用)
		baseRouter.POST("ChartKaRegister", App.Get卡号列表统计制卡)
		baseRouter.POST("ChartAppUserRegister", App.Get应用用户账号注册统计)
		baseRouter.POST("ChartAgentLevel", App.Get代理组织架构图)
	}
	//快验个人中心===========================================
	baseRouter = Router根Admin.Group("/KuaiYan")
	baseRouter.Use(middleware.IsTokenAdmin()) ///鉴权中间件 检查 token 检查是不是管理员令牌

	baseRouter.Use(middleware.IsToken飞鸟快验())
	{
		App := Api.Admin.KuaiYan                              //实现路由的 具体方法位置
		baseRouter.POST("GetCaptchaApiList", App.Q取开启验证码接口列表) // 获取列表
		baseRouter.POST("GetCaptcha", App.Q取英数验证码)
		baseRouter.POST("GetUserInfo", App.Q快验个人信息更新)
		baseRouter.POST("GetSmsCaptcha", App.Q取短信验证码)
		baseRouter.POST("SetPassword", App.Z快验找回密码)
		baseRouter.POST("Register", App.Z快验注册)
		baseRouter.POST("Login", App.D登录)
		baseRouter.POST("GetIsBuyKaList", App.Q取可购买充值卡列表)
		baseRouter.POST("GetPurchasedKaList", App.Q购买充值卡记录)
		baseRouter.POST("GetPayStatus", App.Q取支付通道状态)
		baseRouter.POST("Updater", App.G更新程序)

		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			baseRouter.POST("OutLogin", App.Q注销)
			baseRouter.POST("GetPayPC", App.Y余额充值)
			baseRouter.POST("PayMoneyToKa", App.Y余额购买充值卡)
			baseRouter.POST("UseKa", App.K卡号充值)
		}

	}
	return baseRouter
}

// Agent路由 menu 需要鉴权  menu
func RouterAgent(Router *gin.RouterGroup) *gin.RouterGroup {

	Router.GET("agent", func(context *gin.Context) { //客户经常输入错误,单独注册个路由,跳转正确地址
		context.Redirect(http.StatusFound, "/Agent")
	})

	Router根Agent := Router.Group("Agent") //127.0.0.1:18080/  这个后面第一个不需要 / 符号
	//Router根Agent.Use(middleware.AA())
	Router根Agent.Use(middleware.IsAgentHost())

	Router根Agent.Use(middleware.IsAgent是否关闭())

	//打包静态VueAdmin文件============================
	html := VueAgent.NewHtmlHandler()
	Router根Agent.StaticFS("/assets", http.FS(VueAgentAssets.Assets))
	Router根Agent.GET("/", html.Index)

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
	baseRouter.POST("GetPayPC", MenuApi.Y余额充值)
	baseRouter.POST("GetPayOrderStatus", MenuApi.Q取余额充值订单状态)

	if !(global.GVA_Viper.GetInt("系统模式") == 1) {
		baseRouter.POST("NewPassword", MenuApi.NewPassword) //修改当前Token密码
	}
	//卡号列表===========================================
	baseRouter = Router根Agent.Group("/Ka")
	baseRouter.Use(middleware.IsTokenAgent()) ///鉴权中间件 检查 token 检查是不是管理员令牌

	{
		App := Agent.Api.Ka                            //实现路由的 具体方法位置
		baseRouter.POST("GetList", App.GetKaList)      // 获取列表
		baseRouter.POST("New", App.New)                // 新制卡号
		baseRouter.POST("InventoryNewKa", App.K库存制卡)   // 新制卡号
		baseRouter.POST("GetInfo", App.GetInfo)        // 获取详细信息
		baseRouter.POST("SetStatus", App.Set修改状态)      // 修改状态
		baseRouter.POST("SetAgentNote", App.Set修改代理备注) // 修改状态
		baseRouter.POST("Recover", App.Z追回卡号)
		baseRouter.POST("UseKa", App.K卡号充值)
		baseRouter.POST("ReplaceKaName", App.G更换卡号)
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

	AgentApp := Agent.Api.AgentUser                                                //实现路由的 具体方法位置
	baseRouter.POST("GetAgentUserList", AgentApp.GetAgentUserList)                 // 获取用户列表
	baseRouter.POST("GetAgentUserInfo", AgentApp.GetAgentUserInfo)                 // 获取用户详细信息
	baseRouter.POST("SaveAgentUser", AgentApp.Save代理信息)                            // 保存用户详细信息
	baseRouter.POST("NewAgentUser", AgentApp.New代理信息)                              // 保存用户详细信息
	baseRouter.POST("SetAgentUserStatus", AgentApp.Set修改状态)                        // 保存用户详细信息
	baseRouter.POST("GetAgentKaClassAuthority", AgentApp.GetAgentKaClassAuthority) //取全部可制卡类和已授权卡类
	baseRouter.POST("SetAgentKaClassAuthority", AgentApp.SetAgentKaClassAuthority) //设置代理可制卡类ID
	baseRouter.POST("SendRmbTOAgent", AgentApp.SendRmbTOAgent)                     //转账
	baseRouter.POST("ChartAgentLevel", AgentApp.Get代理组织架构图)

	//代理库存管理===========================================
	baseRouter = Router根Agent.Group("/AgentInventory")
	baseRouter.Use(middleware.IsTokenAgent()) ///鉴权中间件 检查 token 检查是不是管理员令牌

	AgentInventory := Agent.Api.AgentInventory                        //实现路由的 具体方法位置
	baseRouter.POST("GetList", AgentInventory.GetAgentInventoryList)  // 获取列表
	baseRouter.POST("GetKaClassTree", AgentInventory.Get取可创建库存包列表)    // 获取列表
	baseRouter.POST("GetInfo", AgentInventory.GetAgentInventoryInfo)  // 获取详细信息
	baseRouter.POST("NewBuy", AgentInventory.New库存购买)                 // 创建库存包
	baseRouter.POST("Send", AgentInventory.K库存发送)                     // 创建库存包
	baseRouter.POST("GetSubordinateAgent", AgentInventory.Q可发送库存下级代理) // 创建库存包
	baseRouter.POST("Withdraw", AgentInventory.K库存撤回)                 // 撤回转出的库存
	baseRouter.POST("SetEndTime", AgentInventory.K库存延期)
	baseRouter.POST("SetNote", AgentInventory.K库存修改备注)
	//余额日志===========================================
	baseRouter = Router根Agent.Group("/LogMoney")
	baseRouter.Use(middleware.IsTokenAgent()) ///鉴权中间件 检查 token 检查是不是管理员令牌

	{
		App := Agent.Api.LogMoney                       //实现路由的 具体方法位置
		baseRouter.POST("GetList", App.GetLogMoneyList) // 获取列表
	}
	//库存日志===========================================
	baseRouter = Router根Agent.Group("/LogAgentInventory")
	baseRouter.Use(middleware.IsTokenAgent()) ///鉴权中间件 检查 token 检查是不是管理员令牌

	{
		App := Agent.Api.LogAgentInventory                   //实现路由的 具体方法位置
		baseRouter.POST("GetList", App.GetLogAgentInventory) // 获取列表

	}
	//制卡日志===========================================
	baseRouter = Router根Agent.Group("/LogRegisterKa")
	baseRouter.Use(middleware.IsTokenAgent()) ///鉴权中间件 检查 token 检查是不是管理员令牌

	{
		App := Agent.Api.LogRegister               //实现路由的 具体方法位置
		baseRouter.POST("GetList", App.GetLogList) // 获取列表
	}
	//其他操作===========================================
	baseRouter = Router根Agent.Group("/OtherFunc")
	baseRouter.Use(middleware.IsTokenAgent()) ///鉴权中间件 检查 token 检查是不是管理员令牌
	{
		App := Agent.Api.OtherFunc                          //实现路由的 具体方法位置
		baseRouter.POST("SetAppUserKey", App.SetAppUserKey) // 获取列表
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
func RouterWebApi(Router *gin.RouterGroup) *gin.RouterGroup {

	//===========================================
	baseRouter := Router.Group("/WebApi/") //WebApi不做任何加密中间件处理
	baseRouter.Use(middleware.IsWebApiHost())
	baseRouter.Use(middleware.IsTokenWebApi()) ///鉴权中间件 检查 token  单独优先处理
	{
		for 键名, 键值 := range WebApi.J集_UserAPi路由 {
			baseRouter.GET(键名, 键值.Z指向函数)
			baseRouter.POST(键名, 键值.Z指向函数)
		}
	}

	return baseRouter
}
