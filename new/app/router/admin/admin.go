package admin

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"server/core/dist/VueAdmin"
	"server/global"
	controller "server/new/app/controller/admin"
	"server/new/app/router/middleware"
	"strings"
)

type AllRouter struct {
}

func (r *AllRouter) InitAdminRouter(router *gin.RouterGroup) {
	局_管理入口 := global.GVA_Viper.GetString("管理入口")

	// 客户经常输入错误,单独注册个路由,跳转正确地址
	if strings.ToLower(局_管理入口) != 局_管理入口 {
		router.GET(strings.ToLower(局_管理入口), func(context *gin.Context) {
			context.Redirect(http.StatusFound, "/"+局_管理入口)
		})
	}

	adminRouter := router.Group(局_管理入口)

	// 打包静态VueAdmin文件
	html := VueAdmin.NewHtmlHandler()
	adminRouter.GET("/", html.Index)
	adminRouter.GET("/assets/*filepath", func(c *gin.Context) {
		c.FileFromFS("assets/"+c.Param("filepath"), http.FS(VueAdmin.Assets))
	})

	// ========== 无需鉴权的路由 ==========
	局_Base := controller.NewBaseController()
	{
		adminRouter.POST("base/captcha2", 局_Base.Captcha2)
	}
	局_Login := controller.NewLoginController()
	{
		adminRouter.POST("base/login", 局_Login.Login)
	}
	局_InitDB := controller.NewInitDBController()
	{
		adminRouter.POST("base/checkDB", 局_InitDB.CheckDB)
		adminRouter.POST("base/initDB", 局_InitDB.InitDB)
	}

	// ========== 需要鉴权的路由 ==========
	adminRouter = router.Group(局_管理入口)
	adminRouter.Use(middleware.IsTokenAdmin())
	adminRouter.Use(middleware.IsToken飞鸟快验())

	// 菜单(无需复杂鉴权，仅需token)
	局_Menu := controller.NewMenuController()
	{
		adminRouter.POST("menu/getMenu", 局_Menu.GetMenu)
	}

	// ========== 系统基础 ==========
	局_Setting := controller.NewSettingController()
	{
		adminRouter.POST("setting/info", 局_Setting.Info)
	}

	// ========== 监控面板 ==========
	局_Panel := controller.NewPanelController()
	{
		adminRouter.POST("panel/getServerInfo", 局_Panel.GetServerInfo)
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			adminRouter.POST("panel/reloadSystem", 局_Panel.ReloadSystem)
			adminRouter.POST("panel/stopSystem", 局_Panel.StopSystem)
		}
		//图表分析
		adminRouter.POST("panel/chartLinksUser", 局_Panel.ChartLinksUser)
		adminRouter.POST("panel/chartLinksUserIPCity", 局_Panel.ChartLinksUserIPCity)
		adminRouter.POST("panel/chartLinksUserLoginTime", 局_Panel.ChartLinksUserLoginTime)
		adminRouter.POST("panel/chartEveryHourLinksCount", 局_Panel.ChartEveryHourLinksCount)
		adminRouter.POST("panel/chartAppUserClass", 局_Panel.ChartAppUserClass)
		adminRouter.POST("panel/chartUser", 局_Panel.ChartUser)
		adminRouter.POST("panel/chartRMBAddSub", 局_Panel.ChartRMBAddSub)
		adminRouter.POST("panel/chartVipNumberAddSub", 局_Panel.ChartVipNumberAddSub)
		adminRouter.POST("panel/chartAppUser", 局_Panel.ChartAppUser)
		adminRouter.POST("panel/chartAppKa", 局_Panel.ChartAppKa)
		adminRouter.POST("panel/chartAppKaClass", 局_Panel.ChartAppKaClass)
		adminRouter.POST("panel/chartKaRegister", 局_Panel.ChartKaRegister)
		adminRouter.POST("panel/chartAppUserRegister", 局_Panel.ChartAppUserRegister)
		adminRouter.POST("panel/chartAgentLevel", 局_Panel.ChartAgentLevel)
		adminRouter.POST("panel/chartTidTaskData", 局_Panel.ChartTidTaskData)
	}
	//仪表盘(新)
	局_chart := controller.NewChartController()
	{
		adminRouter.POST("panel/getConsumptionRanking", 局_chart.Q取余额消费排行榜)
		adminRouter.POST("panel/getRmbIncreaseRanking", 局_chart.Q取余额增长排行榜)
		adminRouter.POST("panel/getNumberIncreaseRanking", 局_chart.Q取积分消费排行榜)
		adminRouter.POST("panel/gaodeWeather", 局_chart.G高德取天气)
	}

	// ========== 用户/管理员管理 ==========
	局_User := controller.NewUserController()
	{
		adminRouter.GET("user/getAdminInfo", 局_User.GetAdminInfo)
		adminRouter.POST("user/outLogin", 局_User.OutLogin)
		adminRouter.POST("user/adminNewPassword", 局_User.AdminNewPassword)
		adminRouter.POST("user/getUserList", 局_User.GetUserList)
		adminRouter.POST("user/getUserInfo", 局_User.GetUserInfo)
		adminRouter.POST("user/saveUser", 局_User.SaveUser)
		adminRouter.POST("user/newUser", 局_User.NewUser)
		adminRouter.POST("user/setUserStatus", 局_User.SetUserStatus)
		adminRouter.POST("user/setBatchAddRMB", 局_User.BatchAddRMB)
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			adminRouter.POST("user/deleteUser", 局_User.DeleteUser)
		}
	}
	//在线用户管理
	局_LinkUser := controller.NewLinkUserController()
	{
		adminRouter.POST("user/getLinkUserList", 局_LinkUser.GetList)
		adminRouter.POST("user/newWebApiToken", 局_LinkUser.NewWebApiToken)
		adminRouter.POST("user/setTokenOutTime", 局_LinkUser.SetTokenOutTime)
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			adminRouter.POST("user/logout", 局_LinkUser.Logout)
			adminRouter.POST("user/deleteLogout", 局_LinkUser.DeleteLogout)
		}
	}

	// ========== 应用管理 ==========
	局_App := controller.NewAppController()
	{
		adminRouter.POST("app/getList", 局_App.GetList)
		adminRouter.POST("app/new", 局_App.New)
		adminRouter.POST("app/getInfo", 局_App.GetInfo)
		adminRouter.GET("app/getAppIdNameList", 局_App.GetAppIdNameList)
		adminRouter.GET("app/getAllUserApi", 局_App.GetAllUserApi)
		adminRouter.GET("app/getAllWebApi", 局_App.GetAllWebApi)
		adminRouter.GET("app/getAppIdMax", 局_App.GetAppIdMax)
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			adminRouter.POST("app/delete", 局_App.Delete)
			adminRouter.POST("app/saveInfo", 局_App.SaveInfo)
		}
	}
	//应用排序
	局_AppInfo := controller.NewAppInfoController()
	{
		adminRouter.POST("app/setAppSort", 局_AppInfo.SetAppSort)
	}
	//应用WebUser管理
	局_AppInfoWebUser := controller.NewAppInfoWebUserController()
	{
		adminRouter.POST("appWebUser/getInfo", 局_AppInfoWebUser.GetInfo)
	}

	// ========== 软件用户管理 ==========
	局_AppUser := controller.NewAppUserController()
	{
		adminRouter.POST("appUser/batchAddUser", 局_AppUser.BatchAddUser)
	}
	局_AppUserFull := controller.NewAppUserFullController()
	{
		adminRouter.POST("appUser/getList", 局_AppUserFull.GetList)
		adminRouter.POST("appUser/new", 局_AppUserFull.New)
		adminRouter.POST("appUser/getInfo", 局_AppUserFull.Info)
		adminRouter.POST("appUser/saveInfo", 局_AppUserFull.SaveInfo)
		adminRouter.POST("appUser/setStatus", 局_AppUserFull.SetStatus)
		adminRouter.POST("appUser/setBatchAddVipTime", 局_AppUserFull.SetBatchAddVipTime)
		adminRouter.POST("appUser/setBatchAddVipNumber", 局_AppUserFull.SetBatchAddVipNumber)
		adminRouter.POST("appUser/setBatchSetUserConfig", 局_AppUserFull.SetBatchSetUserConfig)
		adminRouter.POST("appUser/setBatchUserClass", 局_AppUserFull.SetBatchUserClass)
		adminRouter.POST("appUser/setBatchAllUserVipTime", 局_AppUserFull.SetBatchAllUserVipTime)
		adminRouter.POST("appUser/batchSetAppUserKey", 局_AppUserFull.BatchSetAppUserKey)
		adminRouter.POST("appUser/batchSetAppUserNote", 局_AppUserFull.BatchSetAppUserNote)
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			adminRouter.POST("appUser/deleteBatch", 局_AppUserFull.DeleteBatch)
			adminRouter.POST("appUser/delete", 局_AppUserFull.Delete)
		}
	}

	// ========== 用户类型管理 ==========
	局_UserClass := controller.NewUserClassController()
	{
		adminRouter.POST("userClass/getList", 局_UserClass.GetList)
		adminRouter.POST("userClass/new", 局_UserClass.New)
		adminRouter.POST("userClass/getInfo", 局_UserClass.Info)
		adminRouter.POST("userClass/saveInfo", 局_UserClass.SaveInfo)
		adminRouter.POST("userClass/getIdNameList", 局_UserClass.GetIdNameList)
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			adminRouter.POST("userClass/delete", 局_UserClass.Delete)
		}
	}

	// ========== 用户云配置管理 ==========
	局_UserConfig := controller.NewUserConfigController()
	{
		adminRouter.POST("userConfig/getList", 局_UserConfig.GetList)
		adminRouter.POST("userConfig/new", 局_UserConfig.New)
		adminRouter.POST("userConfig/getInfo", 局_UserConfig.Info)
		adminRouter.POST("userConfig/setUserConfig", 局_UserConfig.SetUserConfig)
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			adminRouter.POST("userConfig/delete", 局_UserConfig.Delete)
		}
	}

	// ========== 卡号列表管理 ==========
	局_Ka := controller.NewKaFullController()
	{
		adminRouter.POST("ka/getList", 局_Ka.GetList)
		adminRouter.POST("ka/new", 局_Ka.New)
		adminRouter.POST("ka/batchKaNameNew", 局_Ka.BatchKaNameNew)
		adminRouter.POST("ka/getInfo", 局_Ka.Info)
		adminRouter.POST("ka/saveInfo", 局_Ka.SaveInfo)
		adminRouter.POST("ka/setStatus", 局_Ka.SetStatus)
		adminRouter.POST("ka/setAdminNote", 局_Ka.SetAdminNote)
		adminRouter.POST("ka/getKaTemplate", 局_Ka.GetKaTemplate)
		adminRouter.POST("ka/setKaTemplate", 局_Ka.SetKaTemplate)
		adminRouter.POST("ka/recover", 局_Ka.Recover)
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			adminRouter.POST("ka/delete", 局_Ka.Delete)
			adminRouter.POST("ka/deleteBatch", 局_Ka.DeleteBatch)
		}
	}

	// ========== 卡类列表管理 ==========
	局_KaClass := controller.NewKaClassController()
	{
		adminRouter.POST("kaClass/getList", 局_KaClass.GetList)
		adminRouter.POST("kaClass/getListAll", 局_KaClass.GetListAll)
		adminRouter.POST("kaClass/new", 局_KaClass.New)
		adminRouter.POST("kaClass/getInfo", 局_KaClass.Info)
		adminRouter.POST("kaClass/saveInfo", 局_KaClass.SaveInfo)
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			adminRouter.POST("kaClass/delete", 局_KaClass.Delete)
		}
	}

	// ========== 代理账号管理 ==========
	局_AgentUser := controller.NewAgentUserController()
	{
		adminRouter.POST("agent/getKaSalesStatistics", 局_AgentUser.GetKaSalesStatistics)
		adminRouter.POST("agent/setSort", 局_AgentUser.SetSort)
	}
	局_AgentUserFull := controller.NewAgentUserFullController()
	{
		adminRouter.POST("agent/getAgentUserList", 局_AgentUserFull.GetList)
		adminRouter.POST("agent/getAgentUserInfo", 局_AgentUserFull.Info)
		adminRouter.POST("agent/saveAgentUser", 局_AgentUserFull.Save)
		adminRouter.POST("agent/newAgentUser", 局_AgentUserFull.New)
		adminRouter.POST("agent/setAgentUserStatus", 局_AgentUserFull.SetStatus)
		adminRouter.POST("agent/getAgentKaClassAuthority", 局_AgentUserFull.GetAgentKaClassAuthority)
		adminRouter.POST("agent/setAgentKaClassAuthority", 局_AgentUserFull.SetAgentKaClassAuthority)
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			adminRouter.POST("agent/deleteAgentUser", 局_AgentUserFull.Delete)
		}
	}

	// ========== 代理库存管理 ==========
	局_AgentInventory := controller.NewAgentInventoryController()
	{
		adminRouter.POST("agentInventory/getList", 局_AgentInventory.GetList)
		adminRouter.POST("agentInventory/getAgentTreeAndKaClassTree", 局_AgentInventory.GetAgentTreeAndKaClassTree)
		adminRouter.POST("agentInventory/getInfo", 局_AgentInventory.Info)
		adminRouter.POST("agentInventory/new", 局_AgentInventory.New)
		adminRouter.POST("agentInventory/withdraw", 局_AgentInventory.Withdraw)
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			adminRouter.POST("agentInventory/delete", 局_AgentInventory.Delete)
		}
	}

	// ========== 公共变量管理 ==========
	局_PublicData := controller.NewPublicDataController()
	{
		adminRouter.POST("publicData/getList", 局_PublicData.GetList)
		adminRouter.POST("publicData/new", 局_PublicData.New)
		adminRouter.POST("publicData/getInfo", 局_PublicData.Info)
		adminRouter.POST("publicData/saveInfo", 局_PublicData.SaveInfo)
		adminRouter.POST("publicData/setVipLimit", 局_PublicData.SetVipLimit)
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			adminRouter.POST("publicData/delete", 局_PublicData.Delete)
		}
	}

	// ========== 公共函数管理 ==========
	局_PublicJs := controller.NewPublicJsController()
	{
		adminRouter.POST("publicJs/getList", 局_PublicJs.GetList)
		adminRouter.POST("publicJs/getPublicAppList", 局_PublicJs.GetPublicAppList)
		adminRouter.POST("publicJs/new", 局_PublicJs.New)
		adminRouter.POST("publicJs/getInfo", 局_PublicJs.Info)
		adminRouter.POST("publicJs/saveInfo", 局_PublicJs.SaveInfo)
		adminRouter.POST("publicJs/testRunJs", 局_PublicJs.TestExec)
		adminRouter.POST("publicJs/setVipLimit", 局_PublicJs.SetVipLimit)
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			adminRouter.POST("publicJs/delete", 局_PublicJs.Delete)
		}
	}

	// ========== 系统设置 ==========
	局_SetSystem := controller.NewSetSystemFullController()
	{
		adminRouter.POST("setSystem/getInfoSystem", 局_SetSystem.GetInfoSystem)
		adminRouter.POST("setSystem/generateAPIEncryptedSDK", 局_SetSystem.S生成API加密源码SDK)
		adminRouter.POST("setSystem/getInfoPay", 局_SetSystem.GetInfo在线支付)
		adminRouter.POST("setSystem/getInfoSMS", 局_SetSystem.GetInfo短信平台设置)
		adminRouter.POST("setSystem/getInfoCaptcha2", 局_SetSystem.GetInfo行为验证码平台设置)
		adminRouter.POST("setSystem/getInfoCloudStorage", 局_SetSystem.GetInfo云存储设置)
		adminRouter.POST("setSystem/getUserMsgConfig", 局_SetSystem.Get用户消息配置)
		adminRouter.POST("setSystem/getInfoAiConfig", 局_SetSystem.GetInfoAiConfig)
		adminRouter.POST("setSystem/saveInfoSMS", 局_SetSystem.Save短信平台设置)
		adminRouter.POST("setSystem/testSendSMS", 局_SetSystem.F发送短信平台测试)
		adminRouter.POST("setSystem/saveInfoCaptcha2", 局_SetSystem.Save行为验证码平台设置)
		adminRouter.POST("setSystem/saveInfoCloudStorage", 局_SetSystem.Save云存储设置)
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			adminRouter.POST("setSystem/saveInfoSystem", 局_SetSystem.SaveInfoSystem)
			adminRouter.POST("setSystem/saveInfoPay", 局_SetSystem.SaveInfo在线支付)
			adminRouter.POST("setSystem/saveUserMsgConfig", 局_SetSystem.Save用户消息配置)
			adminRouter.POST("setSystem/saveInfoAiConfig", 局_SetSystem.SaveInfoAiConfig)
		}
	}

	// ========== 任务池管理 ==========
	局_TaskPool := controller.NewTaskPoolFullController()
	{
		adminRouter.POST("taskPool/getList", 局_TaskPool.GetList)
		adminRouter.POST("taskPool/new", 局_TaskPool.New)
		adminRouter.POST("taskPool/getInfo", 局_TaskPool.Info)
		adminRouter.POST("taskPool/saveInfo", 局_TaskPool.Save)
		adminRouter.POST("taskPool/setStatus", 局_TaskPool.SetStatus)
		adminRouter.POST("taskPool/deleteTaskQueueTid", 局_TaskPool.ClearQueue)
		adminRouter.POST("taskPool/uuidAddQueue", 局_TaskPool.UuidAddQueue)
		adminRouter.POST("taskPool/batchUuidAddQueue", 局_TaskPool.BatchUuidAddQueue)
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			adminRouter.POST("taskPool/delete", 局_TaskPool.Delete)
		}
	}
	局_TaskPoolType := controller.NewTaskPoolTypeController()
	{
		adminRouter.POST("taskPool/setSort", 局_TaskPoolType.SetSort)
	}
	//任务数据
	局_TaskPoolData := controller.NewTaskPoolDataController()
	{
		adminRouter.POST("taskPoolData/getList", 局_TaskPoolData.GetList)
		adminRouter.POST("taskPoolData/delete", 局_TaskPoolData.Delete)
	}

	// ========== 余额充值订单 ==========
	局_LogRMBPayOrder := controller.NewLogRMBPayOrderController()
	{
		adminRouter.POST("logRMBPayOrder/getList", 局_LogRMBPayOrder.GetList)
		adminRouter.POST("logRMBPayOrder/getInfo", 局_LogRMBPayOrder.Info)
		adminRouter.POST("logRMBPayOrder/new", 局_LogRMBPayOrder.New)
		adminRouter.POST("logRMBPayOrder/setPayOrderNote", 局_LogRMBPayOrder.SetNote)
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			adminRouter.POST("logRMBPayOrder/delete", 局_LogRMBPayOrder.Delete)
			adminRouter.POST("logRMBPayOrder/out", 局_LogRMBPayOrder.Out)
		}
	}

	// ========== 日志模块 ==========
	局_LogUserMsg := controller.NewLogUserMsgController()
	{
		adminRouter.POST("logUserMsg/getList", 局_LogUserMsg.GetList)
		adminRouter.POST("logUserMsg/getInfo", 局_LogUserMsg.Info)
		adminRouter.POST("logUserMsg/deleteDuplicateMsg", 局_LogUserMsg.S删除重复消息)
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			adminRouter.POST("logUserMsg/setIsRead", 局_LogUserMsg.SetIsRead)
			adminRouter.POST("logUserMsg/delete", 局_LogUserMsg.Delete)
		}
	}
	局_LogLogin := controller.NewLogLoginController()
	{
		adminRouter.POST("logLogin/getList", 局_LogLogin.GetList)
		adminRouter.POST("logLogin/getInfo", 局_LogLogin.Info)
		adminRouter.POST("logLogin/delete", 局_LogLogin.Delete)
	}
	局_LogMoney := controller.NewLogMoneyController()
	{
		adminRouter.POST("logMoney/getList", 局_LogMoney.GetList)
		adminRouter.POST("logMoney/getInfo", 局_LogMoney.Info)
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			adminRouter.POST("logMoney/delete", 局_LogMoney.Delete)
		}
	}
	局_LogVipNumber := controller.NewLogVipNumberController()
	{
		adminRouter.POST("logVipNumber/getList", 局_LogVipNumber.GetList)
		adminRouter.POST("logVipNumber/getInfo", 局_LogVipNumber.Info)
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			adminRouter.POST("logVipNumber/delete", 局_LogVipNumber.Delete)
		}
	}
	局_LogRegisterKa := controller.NewLogRegisterKaController()
	{
		adminRouter.POST("logRegisterKa/getList", 局_LogRegisterKa.GetList)
		adminRouter.POST("logRegisterKa/getInfo", 局_LogRegisterKa.Info)
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			adminRouter.POST("logRegisterKa/delete", 局_LogRegisterKa.Delete)
		}
	}
	局_LogAgentOtherFunc := controller.NewLogAgentOtherFuncController()
	{
		adminRouter.POST("logAgentOtherFunc/getList", 局_LogAgentOtherFunc.GetList)
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			adminRouter.POST("logAgentOtherFunc/delete", 局_LogAgentOtherFunc.Delete)
		}
	}
	局_LogAgentInventory := controller.NewLogAgentInventoryController()
	{
		adminRouter.POST("logAgentInventory/getList", 局_LogAgentInventory.GetList)
		adminRouter.POST("logAgentInventory/getInfo", 局_LogAgentInventory.Info)
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			adminRouter.POST("logAgentInventory/delete", 局_LogAgentInventory.Delete)
		}
	}
	局_CronLog := controller.NewCronLogController()
	{
		adminRouter.POST("logCronTask/getList", 局_CronLog.GetList)
		adminRouter.POST("logCronTask/delete", 局_CronLog.Delete)
		adminRouter.POST("logCronTask/deleteBatch", 局_CronLog.DeleteBatch)
	}
	局_logKey := controller.NewLogKeyController()
	{
		adminRouter.POST("logKey/getList", 局_logKey.GetList)
		adminRouter.POST("logKey/info", 局_logKey.Info)
		adminRouter.POST("logKey/delete", 局_logKey.Delete)
	}

	// ========== 黑名单管理 ==========
	局_Blacklist := controller.NewBlacklistController()
	{
		adminRouter.POST("blacklist/getList", 局_Blacklist.GetList)
		adminRouter.POST("blacklist/create", 局_Blacklist.Create)
		adminRouter.POST("blacklist/delete", 局_Blacklist.Delete)
		adminRouter.POST("blacklist/update", 局_Blacklist.Update)
		adminRouter.POST("blacklist/info", 局_Blacklist.Info)
		adminRouter.POST("blacklist/deleteBatch", 局_Blacklist.DeleteBatch)
	}

	// ========== 定时任务管理 ==========
	局_Cron := controller.NewCronController()
	{
		adminRouter.POST("cron/getList", 局_Cron.GetList)
		adminRouter.POST("cron/create", 局_Cron.Create)
		adminRouter.POST("cron/delete", 局_Cron.Delete)
		adminRouter.POST("cron/update", 局_Cron.Update)
		adminRouter.POST("cron/info", 局_Cron.Info)
		adminRouter.POST("cron/deleteBatch", 局_Cron.DeleteBatch)
		adminRouter.POST("cron/updateStatus", 局_Cron.UpdateStatus)
		adminRouter.POST("cron/testRunId", 局_Cron.Z执行)
	}

	// ========== 云存储 ==========
	局_云存储 := controller.NewCloudStorageController()
	{
		adminRouter.POST("cloudStorage/getBaseInfo", 局_云存储.GetBaseInfo)
		adminRouter.POST("cloudStorage/getList", 局_云存储.GetList)
		adminRouter.POST("cloudStorage/moveTo", 局_云存储.MoveTo)
		adminRouter.POST("cloudStorage/download", 局_云存储.Download)
		adminRouter.POST("cloudStorage/getDownloadUrl", 局_云存储.GetDownloadUrl)
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			adminRouter.POST("cloudStorage/getUpToken", 局_云存储.GetUpToken)
			adminRouter.POST("cloudStorage/delete", 局_云存储.Delete)
		}
	}

	// ========== 工具 ==========
	局_ApkTools := controller.NewApkToolsController()
	{
		adminRouter.POST("apkTools/getList", 局_ApkTools.GetList)
		adminRouter.POST("apkTools/getTaskIdStatus", 局_ApkTools.GetTaskIdStatus)
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			adminRouter.POST("apkTools/getUploadToken", 局_ApkTools.GetUploadToken)
			adminRouter.POST("apkTools/createApkAddFNKYTask", 局_ApkTools.CreateApkAddFNKYTask)
		}
	}
	局_exeTools := controller.NewExeToolsController()
	{
		adminRouter.POST("exeTools/getList", 局_exeTools.GetList)
		adminRouter.POST("exeTools/getTaskIdStatus", 局_exeTools.GetTaskIdStatus)
		adminRouter.POST("exeTools/getUiList", 局_exeTools.GetUiList)
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			adminRouter.POST("exeTools/getUploadToken", 局_exeTools.GetUploadToken)
			adminRouter.POST("exeTools/createExeAddFNKYTask", 局_exeTools.CreateExeAddFNKYTask)
		}
	}

	// ========== 活动管理 ==========
	局_AppPromotionConfig := controller.NewAppPromotionConfigController()
	{
		adminRouter.POST("appPromotionConfig/info", 局_AppPromotionConfig.Info)
		adminRouter.POST("appPromotionConfig/create", 局_AppPromotionConfig.Create)
		adminRouter.POST("appPromotionConfig/delete", 局_AppPromotionConfig.Delete)
		adminRouter.POST("appPromotionConfig/update", 局_AppPromotionConfig.Update)
		adminRouter.POST("appPromotionConfig/getList", 局_AppPromotionConfig.GetList)
		adminRouter.POST("appPromotionConfig/setSort", 局_AppPromotionConfig.Sort)
		adminRouter.POST("appPromotionConfig/reset", 局_AppPromotionConfig.Reset)
	}

	// ========== CPS管理 ==========
	局_CpsInfo := controller.NewCpsInfoController()
	{
		adminRouter.POST("cpsInfo/getList", 局_CpsInfo.GetList)
		adminRouter.POST("cpsInfo/info", 局_CpsInfo.Info)
		adminRouter.POST("cpsInfo/update", 局_CpsInfo.Update)
	}
	局_CpsPayOrder := controller.NewCpsPayOrderController()
	{
		adminRouter.POST("cpsPayOrder/getList", 局_CpsPayOrder.GetList)
		adminRouter.POST("cpsPayOrder/info", 局_CpsPayOrder.Info)
		adminRouter.POST("cpsPayOrder/delete", 局_CpsPayOrder.Delete)
		adminRouter.POST("cpsPayOrder/setNote", 局_CpsPayOrder.SerNote)
	}

	// ========== 签到管理 ==========
	局_CheckIn := controller.NewCheckInInfoController()
	{
		adminRouter.POST("checkInInfo/getList", 局_CheckIn.GetList)
		adminRouter.POST("checkInInfo/info", 局_CheckIn.Info)
		adminRouter.POST("checkInInfo/update", 局_CheckIn.Update)
	}
	局_CheckInScoreLog := controller.NewCheckInScoreLogController()
	{
		adminRouter.POST("checkInScoreLog/getList", 局_CheckInScoreLog.GetList)
		adminRouter.POST("checkInScoreLog/delete", 局_CheckInScoreLog.Delete)
	}

	// ========== 快验个人中心 ==========
	局_KuaiYan := controller.NewKuaiYanController()
	{
		adminRouter.POST("kuaiYan/getCaptchaApiList", 局_KuaiYan.GetCaptchaApiList)
		adminRouter.POST("kuaiYan/getCaptcha", 局_KuaiYan.GetCaptcha)
		adminRouter.POST("kuaiYan/getUserInfo", 局_KuaiYan.GetUserInfo)
		adminRouter.POST("kuaiYan/getSmsCaptcha", 局_KuaiYan.GetSmsCaptcha)
		adminRouter.POST("kuaiYan/setPassword", 局_KuaiYan.SetPassword)
		adminRouter.POST("kuaiYan/register", 局_KuaiYan.Register)
		adminRouter.POST("kuaiYan/login", 局_KuaiYan.Login)
		adminRouter.POST("kuaiYan/getIsBuyKaList", 局_KuaiYan.GetIsBuyKaList)
		adminRouter.POST("kuaiYan/getPurchasedKaList", 局_KuaiYan.GetPurchasedKaList)
		adminRouter.POST("kuaiYan/getPayStatus", 局_KuaiYan.GetPayStatus)
		adminRouter.POST("kuaiYan/updater", 局_KuaiYan.Updater)
		if !(global.GVA_Viper.GetInt("系统模式") == 1) {
			adminRouter.POST("kuaiYan/outLogin", 局_KuaiYan.OutLogin)
			adminRouter.POST("kuaiYan/getPayPC", 局_KuaiYan.GetPayPC)
			adminRouter.POST("kuaiYan/payMoneyToKa", 局_KuaiYan.PayMoneyToKa)
			adminRouter.POST("kuaiYan/useKa", 局_KuaiYan.UseKa)
		}
	}
}
