package Api

import (
	"server/api/Admin/AgentInventory"
	"server/api/Admin/AgentUser"
	"server/api/Admin/App"
	"server/api/Admin/AppUser"
	"server/api/Admin/InitDB"
	"server/api/Admin/Ka"
	"server/api/Admin/KaClass"
	"server/api/Admin/KuaiYan"
	"server/api/Admin/LinkUser"
	"server/api/Admin/LogAgentInventory"
	"server/api/Admin/LogAgentOtherFunc"
	"server/api/Admin/LogLogin"
	"server/api/Admin/LogMoney"
	"server/api/Admin/LogRMBPayOrder"
	"server/api/Admin/LogRegisterka"
	"server/api/Admin/LogUserMsg"
	"server/api/Admin/LogVipNumber"
	"server/api/Admin/Menu"
	"server/api/Admin/Panel"
	"server/api/Admin/PublicData"
	"server/api/Admin/PublicJs"
	"server/api/Admin/SetSystem"
	"server/api/Admin/TaskPool"
	"server/api/Admin/User"
	"server/api/Admin/UserClass"
	"server/api/Admin/UserConfig"
	"server/api/Admin/base"
)

type _Admin struct {
	//成员名     包名.包成员
	Base              base.BaseApi
	InitDb            InitDB.DBApi
	Menu              Menu.MenuApi
	LinkUserApi       LinkUser.LinkUserApi
	User              User.Api
	AgentUser         AgentUser.Api
	AgentInventory    AgentInventory.Api
	App               App.Api
	AppUser           AppUser.Api
	UserClass         UserClass.Api
	KaClass           KaClass.Api
	Ka                Ka.Api
	PublicData        PublicData.Api
	UserConfig        UserConfig.Api
	PublicJs          PublicJs.Api
	TaskPool          TaskPool.Api
	SetSystem         SetSystem.Api
	LogLogin          LogLogin.Api
	LogMoney          LogMoney.Api
	LogAgentInventory LogAgentInventory.Api
	LogVipNumber      LogVipNumber.Api
	LogRegister       LogRegisterKa.Api
	LogUserMsg        LogUserMsg.Api
	LogRMBPayOrder    LogRMBPayOrder.Api
	LogAgentOtherFunc LogAgentOtherFunc.Api
	Panel             Panel.Api
	KuaiYan           KuaiYan.Api
}

// api实例化 路由内可以调用
var Admin = new(_Admin)
