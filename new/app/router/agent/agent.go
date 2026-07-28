package agent

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"server/core/dist/VueAgent"
	controller "server/new/app/controller/agent"
	"server/new/app/global"
	dbm "server/new/app/models/db"
	mid2 "server/new/app/router/middleware"
	"strings"
)

type AllRouter struct {
}

func (r *AllRouter) InitAgentRouter(router *gin.RouterGroup) {
	agentResponseCamelMiddleware := func(c *gin.Context) {
		c.Set("isAgentResponseCamel", true)
		c.Next()
	}

	agentPrefix := global.GVA_Viper.GetString("代理入口")
	if strings.ToLower(agentPrefix) != agentPrefix {
		router.GET(strings.ToLower(agentPrefix), func(c *gin.Context) {
			c.Redirect(http.StatusFound, "/"+agentPrefix)
		})
	}

	rootRouter := router.Group(agentPrefix)
	rootRouter.Use(mid2.IsAgent是否关闭())
	rootRouter.Use(agentResponseCamelMiddleware)

	htmlHandler := VueAgent.NewHtmlHandler()
	rootRouter.GET("/", htmlHandler.Index)
	rootRouter.GET("/assets/*filepath", func(c *gin.Context) {
		c.FileFromFS("assets/"+c.Param("filepath"), http.FS(VueAgent.Assets))
	})

	baseController := controller.NewAgentBaseController()
	{
		baseRouter := rootRouter.Group("/Base")
		baseRouter.POST("Captcha", baseController.Captcha)
		baseRouter.POST("Login", baseController.Login)
	}

	agentRouter := rootRouter.Group("")
	agentRouter.Use(mid2.IsTokenAgent())

	settingController := controller.NewSettingController()
	{
		agentRouter.POST("Setting/getInfoPay", settingController.GetPayInfo)
		agentRouter.POST("Setting/setInfoPay", settingController.SetPayInfo)
		agentRouter.POST("Setting/setBaseInfo", settingController.SetBaseInfo)
		agentRouter.POST("Setting/getBaseInfo", settingController.GetBaseInfo)
		agentRouter.POST("Setting/getAgentUserConfig", settingController.GetAgentUserConfig)
		agentRouter.POST("Setting/delAgentUserConfig", settingController.DelAgentUserConfig)
		agentRouter.POST("Setting/newAgentUserConfig", settingController.NewAgentUserConfig)
		agentRouter.POST("Setting/saveAgentUserConfig", settingController.SaveAgentUserConfig)
	}

	appUserController := controller.NewAppUserController()
	{
		agentRouter.Group("", mid2.Is代理鉴权([]int{dbm.D代理功能_查看归属软件用户})).POST("AppUser/GetList", appUserController.GetList)
		agentRouter.Group("", mid2.Is代理鉴权([]int{dbm.D代理功能_查看归属软件用户})).POST("AppUser/GetInfo", appUserController.GetAppUserInfo)
		agentRouter.POST("AppUser/SetStatus", appUserController.Set修改状态)
		agentRouter.POST("AppUser/SaveUser", appUserController.Save用户信息)
		agentRouter.Group("", mid2.Is代理鉴权([]int{dbm.D代理功能_修改用户密码})).POST("AppUser/SetPassUser", appUserController.Set用户密码)
	}

	userClassController := controller.NewUserClassController()
	{
		agentRouter.POST("UserClass/GetIdNameList", userClassController.GetIdNameList)
	}

	agentMenuController := controller.NewAgentMenuController()
	{
		agentRouter.GET("Menu/GetAgentInfo", agentMenuController.GetAgentInfo)
		agentRouter.POST("Menu/GetAgentInfo", agentMenuController.GetAgentInfo)
		agentRouter.POST("Menu/OutLogin", agentMenuController.OutLogin)
		agentRouter.POST("Menu/NewPassword", agentMenuController.NewPassword)
		agentRouter.POST("Menu/GetPayStatus", agentMenuController.Q取支付通道状态)
		agentRouter.POST("Menu/GetPayStatus2", agentMenuController.Q取支付通道状态2)
		agentRouter.POST("Menu/GetPayPC", agentMenuController.Y余额充值)
		agentRouter.POST("Menu/GetPayOrderStatus", agentMenuController.Q取余额充值订单状态)
	}

	agentUserController := controller.NewAgentUserController()
	{
		agentRouter.POST("Agent/GetKaSalesStatistics", agentUserController.GetKaSalesStatistics)
		agentRouter.POST("Agent/GetAgentUserList", agentUserController.GetAgentUserList)
		agentRouter.POST("Agent/GetAgentUserInfo", agentUserController.GetAgentUserInfo)
		agentRouter.POST("Agent/SaveAgentUser", agentUserController.Save代理信息)
		agentRouter.POST("Agent/NewAgentUser", agentUserController.New代理信息)
		agentRouter.POST("Agent/SetAgentUserStatus", agentUserController.Set修改状态)
		agentRouter.POST("Agent/DeleteAgentUser", agentUserController.Delete代理用户)
		agentRouter.POST("Agent/GetAgentKaClassAuthority", agentUserController.GetAgentKaClassAuthority)
		agentRouter.POST("Agent/SetAgentKaClassAuthority", agentUserController.SetAgentKaClassAuthority)
		agentRouter.POST("Agent/SendRmbTOAgent", agentUserController.SendRmbTOAgent)
		agentRouter.POST("Agent/ChartAgentLevel", agentUserController.Get代理组织架构图)
	}

	kaClassController := controller.NewKaClassController()
	{
		agentRouter.POST("KaClass/GetList", kaClassController.GetList)
	}

	kaClassUpPriceController := controller.NewKaClassUpPriceController()
	{
		agentRouter.Group("", mid2.Is代理鉴权([]int{dbm.D代理功能_卡类调价})).POST("KaClassUpPrice/Save", kaClassUpPriceController.Save)
		agentRouter.Group("", mid2.Is代理鉴权([]int{dbm.D代理功能_卡类调价})).POST("KaClassUpPrice/Delete", kaClassUpPriceController.Delete)
	}

	kaController := controller.NewAgentKaController()
	{
		agentRouter.POST("Ka/GetList", kaController.GetKaList)
		agentRouter.POST("Ka/GetInfo", kaController.GetInfo)
		agentRouter.POST("Ka/New", kaController.New)
		agentRouter.POST("Ka/InventoryNewKa", kaController.K库存制卡)
		agentRouter.POST("Ka/SetStatus", kaController.Set修改状态)
		agentRouter.POST("Ka/SetAgentNote", kaController.Set修改代理备注)
		agentRouter.POST("Ka/Recover", kaController.Z追回卡号)
		agentRouter.POST("Ka/ReplaceKaName", kaController.G更换卡号)
		agentRouter.POST("Ka/UseKa", kaController.K卡号充值)
		agentRouter.POST("Ka/ChartKaRegister", kaController.Get卡号列表统计制卡)
		agentRouter.POST("Ka/GetKaTemplate", kaController.Q取卡号生成模板)
		agentRouter.POST("Ka/SetKaTemplate", kaController.Set修改卡号生成模板)
	}

	agentInventoryController := controller.NewAgentInventoryController()
	{
		agentRouter.POST("AgentInventory/GetList", agentInventoryController.GetAgentInventoryList)
		agentRouter.POST("AgentInventory/GetInfo", agentInventoryController.GetAgentInventoryInfo)
		agentRouter.POST("AgentInventory/NewBuy", agentInventoryController.New库存购买)
		agentRouter.POST("AgentInventory/Withdraw", agentInventoryController.K库存撤回)
		agentRouter.POST("AgentInventory/Send", agentInventoryController.K库存发送)
		agentRouter.POST("AgentInventory/GetSubordinateAgent", agentInventoryController.Q可发送库存下级代理)
		agentRouter.POST("AgentInventory/SetEndTime", agentInventoryController.K库存延期)
		agentRouter.POST("AgentInventory/GetKaClassTree", agentInventoryController.Get取可创建库存包列表)
		agentRouter.POST("AgentInventory/SetNote", agentInventoryController.K库存修改备注)
	}

	otherFuncController := controller.NewAgentOtherFuncController()
	{
		agentRouter.POST("OtherFunc/SetAppUserKey", otherFuncController.SetAppUserKey)
	}

	appController := controller.NewAgentAppController()
	{
		agentRouter.GET("App/GetAppIdNameList", appController.GetAppIdNameList)
		agentRouter.POST("App/GetAppIdNameList", appController.GetAppIdNameList)
	}

	agentLogMoneyController := controller.NewAgentLogMoneyController()
	{
		agentRouter.POST("LogMoney/GetList", agentLogMoneyController.GetList)
		agentRouter.POST("LogMoney/GetInfo", agentLogMoneyController.Info)
		agentRouter.POST("LogMoney/Delete", agentLogMoneyController.Delete)
	}

	agentLogRegisterKaController := controller.NewAgentLogRegisterKaController()
	{
		agentRouter.POST("LogRegisterKa/GetList", agentLogRegisterKaController.GetList)
		agentRouter.POST("LogRegisterKa/GetInfo", agentLogRegisterKaController.Info)
		agentRouter.POST("LogRegisterKa/Delete", agentLogRegisterKaController.Delete)
	}

	agentLogAgentInventoryController := controller.NewAgentLogAgentInventoryController()
	{
		agentRouter.POST("LogAgentInventory/GetList", agentLogAgentInventoryController.GetList)
		agentRouter.POST("LogAgentInventory/GetInfo", agentLogAgentInventoryController.Info)
		agentRouter.POST("LogAgentInventory/Delete", agentLogAgentInventoryController.Delete)
	}

	withdrawController := controller.NewWithdrawController()
	{
		agentRouter.POST("withdraw/getConfig", withdrawController.GetConfig)
		agentRouter.POST("withdraw/uploadPayeeQr", withdrawController.UploadPayeeQr)
		agentRouter.POST("withdraw/image", withdrawController.Image)
		agentRouter.POST("withdraw/create", withdrawController.Create)
		agentRouter.POST("withdraw/list", withdrawController.List)
		agentRouter.POST("withdraw/detail", withdrawController.Detail)
		agentRouter.POST("withdraw/cancel", withdrawController.Cancel)
	}
}
