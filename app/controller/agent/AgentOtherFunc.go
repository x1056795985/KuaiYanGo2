package controller

import (
	"EFunc/utils"
	"github.com/gin-gonic/gin"
	"server/app/global"
	"server/app/logic/common/agent"
	"server/app/logic/common/agentLevel"
	"server/app/logic/common/log"
	"server/app/models/constant"
	dbm "server/app/models/db"
	"server/app/models/old/response"
	"server/app/service"
	"time"
)

type AgentOtherFunc struct{}

func NewAgentOtherFuncController() *AgentOtherFunc {
	return &AgentOtherFunc{}
}

type Agent修改绑定请求 struct {
	AppId int    `json:"appId"`
	User  string `json:"user"`
	Key   string `json:"key"`
}

func (A *AgentOtherFunc) SetAppUserKey(c *gin.Context) {
	var 请求 Agent修改绑定请求
	if err := c.ShouldBindJSON(&请求); err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}
	if !service.NewAppInfo(c, global.GVA_DB).AppId是否存在(请求.AppId) {
		response.FailWithMessage("应用不存在", c)
		return
	}

	局_可操作AppId := agent.L_agent.Id取代理可操作应用AppId列表(c, c.GetInt("Uid"))
	if !utils.S数组_整数是否存在(局_可操作AppId, 请求.AppId) {
		response.FailWithMessage("无该应用操作权限,请联系上级授权该应用任意制卡卡类,获取应用权限", c)
		return
	}

	局_AppUserId := service.NewAppUser(c, global.GVA_DB, 请求.AppId).User或卡号取Id(请求.AppId, 请求.User)
	if 局_AppUserId == 0 {
		response.FailWithMessage("用户不存在", c)
		return
	}

	局_用户详情, err := service.NewAppUser(c, global.GVA_DB, 请求.AppId).Id取详情(请求.AppId, 局_AppUserId)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if 局_用户详情.AgentUid != 0 && 局_用户详情.AgentUid != c.GetInt("Uid") {
		response.FailWithMessage("只能操作自己的归属用户", c)
		return
	}

	if err = service.NewAppUser(c, global.GVA_DB, 请求.AppId).Set绑定信息(请求.AppId, 局_用户详情.Uid, 请求.Key); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	局_用户名 := service.NewAppUser(c, global.GVA_DB, 请求.AppId).Id取User(请求.AppId, 局_AppUserId)
	局_DB := *global.GVA_DB
	_, err = service.NewLogKey(c, &局_DB).Create(&dbm.DB_LogKey{
		Type:   constant.LogKey_换绑,
		User:   局_用户名,
		Uid:    局_用户详情.Uid,
		AppId:  请求.AppId,
		OldKey: 局_用户详情.Key,
		NewKey: 请求.Key,
		Time:   time.Now().Unix(),
		Ip:     c.ClientIP(),
		Note:   "代理:" + c.GetString("User") + ",操作修改绑定信息",
	})
	if err != nil {
		global.GVA_LOG.Println("修改绑定信息日志写入失败:" + err.Error())
	}

	局_信息 := "修改绑定信息 '" + 局_用户详情.Key + "'  ->  '" + 请求.Key + "'"
	log.L_log.Log_写代理操作日志(c.GetInt("Uid"), agentLevel.L_agentLevel.Q取Id代理级别(c, c.GetInt("Uid")), 请求.AppId, 局_AppUserId, 局_用户名, dbm.D代理功能_修改用户绑定, c.ClientIP(), 局_信息)
	response.OkWithMessage("操作成功", c)
}
