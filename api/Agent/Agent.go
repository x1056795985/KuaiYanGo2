package Agent

import (
	"server/api/Agent/AgentInventory"
	"server/api/Agent/AgentUser"
	"server/api/Agent/Ka"
	"server/api/Agent/LogAgentInventory"
	"server/api/Agent/LogMoney"
	LogRegisterKa "server/api/Agent/LogRegisterka"
	"server/api/Agent/Menu"
	"server/api/Agent/OtherFunc"
	"server/api/Agent/base"
)

type _Agent struct {
	//成员名     包名.包成员
	Base              base.BaseApi
	Menu              Menu.Api
	Ka                Ka.Api
	AgentUser         AgentUser.Api
	AgentInventory    AgentInventory.Api
	LogMoney          LogMoney.Api
	LogAgentInventory LogAgentInventory.Api
	LogRegister       LogRegisterKa.Api
	OtherFunc         OtherFunc.Api
	/*LinkUserApi       LinkUser.LinkUserApi
	User              User.Api


	App               App.Api
	AppUser           AppUser.Api
	UserClass         UserClass.Api
	KaClass           KaClass.Api
	Ka                Ka.Api
	PublicData        PublicData.Api
	PublicJs          PublicJs.Api
	TaskPool          TaskPool.Api
	SetSystem         SetSystem.Api
	LogLogin          LogLogin.Api


	LogVipNumber      LogVipNumber.Api

	LogUserMsg        LogUserMsg.Api
	LogRMBPayOrder    LogRMBPayOrder.Api
	Panel             Panel.Api
	KuaiYan           KuaiYan.Api*/
}

// api实例化 路由内可以调用
var Api = new(_Agent)
