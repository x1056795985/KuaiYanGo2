package controller

import (
	"EFunc/utils"
	"github.com/gin-gonic/gin"
	App服务 "server/Service/Ser_AppInfo"
	"server/Service/Ser_AppUser"
	"server/Service/Ser_LinkUser"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/logic/webSocket"
	"server/structs/Http/response"
	DB "server/structs/db"
)

type LinkUserCtrl struct {
	Common.Common
}

func NewLinkUserController() *LinkUserCtrl {
	return &LinkUserCtrl{}
}

type 请求_LinkUserGetList struct {
	Page     int    `json:"page"`
	Size     int    `json:"size"`
	Status   int    `json:"status"`
	Tourist  int    `json:"tourist"`
	Type     int    `json:"type"`
	Keywords string `json:"keywords"`
	Order    int    `json:"order"`
	AppId    int    `json:"appId"`
}

type 响应_LinkUserGetList struct {
	List  []DB_LinksToken2 `json:"list"`
	Count int64            `json:"count"`
}

type DB_LinksToken2 struct {
	DB.DB_LinksToken
	AppName string `json:"appName"`
	Note    string `json:"note"`
}

type 请求_LinkUserNewWebApiToken struct {
	DB.DB_LinksToken
}

type 请求_LinkUserSetTokenOutTime struct {
	Id      []int `json:"id"`
	OutTime int   `json:"outTime"`
}

type 请求_LinkUserIDArray struct {
	Id []int `json:"id"`
}

// GetList 获取在线用户列表
func (C *LinkUserCtrl) GetList(c *gin.Context) {
	var 请求 请求_LinkUserGetList
	if !C.ToJSON(c, &请求) {
		return
	}

	var DB_LinksToken []DB_LinksToken2
	var 总数 int64
	局_DB := global.GVA_DB.Model(DB.DB_LinksToken{})

	if 请求.Order == 1 {
		局_DB.Order("Id ASC")
	} else {
		局_DB.Order("Id DESC")
	}

	if 请求.Tourist == 1 {
		局_DB.Where("User != ?", "游客")
	}
	if 请求.Status == 1 || 请求.Status == 2 {
		局_DB.Where("Status = ?", 请求.Status)
	}
	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1:
			局_DB.Where("Id = ?", 请求.Keywords)
		case 2:
			局_文本数组 := utils.Z正则_取全部匹配子文本(请求.Keywords, "([A-Za-z0-9]+)")
			if len(局_文本数组) == 1 {
				局_DB.Where("User  LIKE ?", "%"+请求.Keywords+"%")
			} else {
				局_DB.Where("User IN ? ", 局_文本数组)
			}
		case 3:
			局_DB.Where("LOCATE(?, `Key` )>0 ", 请求.Keywords)
		case 4:
			局_DB.Where("Tab LIKE ?", "%"+请求.Keywords+"%")
		case 5:
			局_DB.Where("AppVer LIKE ?", "%"+请求.Keywords+"%")
		case 6:
			局_DB.Where("AgentUid LIKE ?", "%"+请求.Keywords+"%")
		}
	}
	if 请求.AppId > 0 {
		局_DB.Where("LoginAppid = ?", 请求.AppId)
	}

	err := 局_DB.Count(&总数).Limit(请求.Size).Omit("app_name").Offset((请求.Page - 1) * 请求.Size).Find(&DB_LinksToken).Error
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		global.GVA_LOG.Error("LinkUserGetList:" + err.Error())
		return
	}

	var AppName = App服务.AppInfo取map列表Int(true)
	for 索引 := range DB_LinksToken {
		DB_LinksToken[索引].AppName = AppName[DB_LinksToken[索引].LoginAppid]
		if DB_LinksToken[索引].Uid > 0 {
			DB_LinksToken[索引].Note = Ser_AppUser.Uid取备注(DB_LinksToken[索引].LoginAppid, DB_LinksToken[索引].Uid)
		}
	}

	response.OkWithDetailed(响应_LinkUserGetList{List: DB_LinksToken, Count: 总数}, "获取成功", c)
}

// NewWebApiToken 创建webApi使用的Token
func (C *LinkUserCtrl) NewWebApiToken(c *gin.Context) {
	var 请求 DB.DB_LinksToken
	if !C.ToJSON(c, &请求) {
		return
	}
	if len(请求.Key) > 191 {
		response.FailWithMessage("权限太多了,数据库字段存不下,减少一些吧", c)
		return
	}

	在线信息, err := Ser_LinkUser.NewWebApiToken(请求.OutTime, 请求.Key, 请求.Tab)
	if err != nil {
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithData(在线信息, c)
}

// SetTokenOutTime 修改令牌自动注销时间
func (C *LinkUserCtrl) SetTokenOutTime(c *gin.Context) {
	var 请求 请求_LinkUserSetTokenOutTime
	if !C.ToJSON(c, &请求) {
		return
	}
	if len(请求.Id) == 0 {
		response.FailWithMessage("id数量不能为0", c)
		return
	}
	err := Ser_LinkUser.Set自动注销超时时间(请求.OutTime, 请求.Id)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(c)
}

// Logout 批量注销在线
func (C *LinkUserCtrl) Logout(c *gin.Context) {
	var 请求 请求_LinkUserIDArray
	if !C.ToJSON(c, &请求) {
		return
	}
	if len(请求.Id) == 0 {
		response.FailWithMessage("Id数组为空", c)
		return
	}

	err := Ser_LinkUser.Set批量注销(请求.Id, Ser_LinkUser.Z注销_管理员手动注销)
	if err != nil {
		response.FailWithMessage("注销失败", c)
		global.GVA_LOG.Error("Logout:" + err.Error())
		return
	}
	for _, v := range 请求.Id {
		webSocket.L_webSocket.RemoveConnection(v)
	}
	response.OkWithMessage("注销成功", c)
}

// DeleteLogout 批量删除已注销
func (C *LinkUserCtrl) DeleteLogout(c *gin.Context) {
	var 请求 请求_LinkUserIDArray
	if !C.ToJSON(c, &请求) {
		return
	}
	if len(请求.Id) == 0 {
		response.FailWithMessage("Id数组为空", c)
		return
	}

	var err error
	if 请求.Id[0] == -1 {
		err = global.GVA_DB.Model(DB.DB_LinksToken{}).Where("Status = 2").Delete("").Error
	} else {
		err = global.GVA_DB.Model(DB.DB_LinksToken{}).Where("Id IN ? ", 请求.Id).Where("Status = 2").Delete("").Error
	}
	if err != nil {
		response.FailWithMessage("已注销删除失败", c)
		global.GVA_LOG.Error("DeleteLogout:" + err.Error())
		return
	}
	response.OkWithMessage("已注销删除成功", c)
}
