package User

import (
	"github.com/gin-gonic/gin"
	"server/Service/Ser_AppInfo"
	"server/global"
	"server/structs/Http/response"
	DB "server/structs/db"
	"strconv"
)

type Api struct{}

// GetInfo 获取代理信息
func (a *Api) GetInfo(c *gin.Context) {
	Uid := c.GetInt("Uid")
	var DB_user DB.DB_User
	err := global.GVA_DB.Model(DB.DB_User{}).Where("id = ?", Uid).First(&DB_user).Error

	if err != nil {
		response.FailWithMessage("查询失败", c)
		global.GVA_LOG.Error("Uid:" + strconv.Itoa(Uid) + "GetInfo错误:" + err.Error())
		return
	}

	response.OkWithDetailed(结构响应_GetInfo{
		Info:          DB_user,
		UserMsgNoRead: 0,
	}, "获取成功", c)
	return
}

type 结构响应_GetInfo struct {
	Info          DB.DB_User `json:"Info"`
	UserMsgNoRead int64      `json:"UserMsgNoRead"`
}

// GetUserInfo
func (a *Api) GetUserInfo(c *gin.Context) {
	var 请求 结构请求_单id
	//{"Id":2}
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	var DB_user DB_User2

	err = global.GVA_DB.Model(DB.DB_User{}).Omit("PassWord", "SuperPassWord").Where("id = ?", 请求.Id).Find(&DB_user).Error
	// 没查到数据
	if err != nil {
		response.FailWithMessage("查询用户详细信息失败", c)
		return
	}
	if DB_user.UPAgentId == 0 {
		response.FailWithMessage("非代理不可查询", c)
		return
	}

	if agent.L_agent.Q取上级代理的子级代理级别(c, c.GetInt("Uid"), DB_user.UPAgentId) == 0 {
		response.FailWithMessage("只能查询自己下级代理信息", c)
		return
	}
	DB_user.Role = agentLevel.L_agentLevel.Q取Id代理级别(c, DB_user.Id)
	if DB_user.LoginAppid > 0 {
		AppName := ""
		Ser_AppInfo.AppId取应用名称(DB_user.LoginAppid)
		DB_user.LoginAppName = AppName
	}
	response.OkWithDetailed(DB_user, "获取成功", c)
	return
}

type DB_User2 struct {
	DB.DB_User
	LoginAppName string `json:"LoginAppName"` //登录平台App名字
	Role         int    `json:"Role"`         //
}

type 结构请求_单id struct {
	Id int `json:"Id"`
}
