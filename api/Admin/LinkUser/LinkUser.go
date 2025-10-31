// Package LinkUser web在线用户列表api
package LinkUser

import (
	"EFunc/utils"
	"github.com/gin-gonic/gin"
	App服务 "server/Service/Ser_AppInfo"
	"server/Service/Ser_AppUser"
	"server/Service/Ser_LinkUser"
	"server/global"
	"server/new/app/logic/common/log"
	"server/structs/Http/response"
	DB "server/structs/db"
)

type LinkUserApi struct{}

// GetLinkUserList
// 获取在线用户列表
func (a *LinkUserApi) GetLinkUserList(c *gin.Context) {
	var 请求 结构请求_GetLinkUserList
	//{"Type":"2","Size":10,"Page":1,"Status":"1","keywords":"1"}
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
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
		case 1: //在线id
			局_DB.Where("Id = ?", 请求.Keywords)
		case 2: //用户名
			局_文本数组 := utils.Z正则_取全部匹配子文本(请求.Keywords, "([A-Za-z0-9]+)")
			if len(局_文本数组) == 1 {
				局_DB.Where("User  LIKE ?", "%"+请求.Keywords+"%")
			} else {
				局_DB.Where("User IN ? ", 局_文本数组)
			}
		case 3: //绑定信息
			局_DB.Where("LOCATE(?, `Key` )>0 ", 请求.Keywords)
		case 4: //动态标签
			局_DB.Where("Tab LIKE ?", "%"+请求.Keywords+"%")
		case 5: //AppVer  软件版本
			局_DB.Where("AppVer LIKE ?", "%"+请求.Keywords+"%")
		case 6: //代理标识Uid
			局_DB.Where("AgentUid LIKE ?", "%"+请求.Keywords+"%")
		}
	}

	if 请求.AppId > 0 {
		局_DB.Where("LoginAppid = ?", 请求.AppId)
	}

	//Count(&总数) 必须放在where 后面 不然值会被清0
	err = 局_DB.Count(&总数).Limit(请求.Size).Omit("app_name").Offset((请求.Page - 1) * 请求.Size).Find(&DB_LinksToken).Error
	//fmt.Println("在线用户总数%d", 总数, DB_LinksToken)
	// 没查到数据  或  取反(密码正确)

	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		global.GVA_LOG.Error("GetLinkUserList:" + err.Error())
		return
	}

	var AppName = App服务.AppInfo取map列表Int()

	for 索引 := range DB_LinksToken {
		DB_LinksToken[索引].AppName = AppName[DB_LinksToken[索引].LoginAppid]
		if DB_LinksToken[索引].Uid > 0 {
			//过于繁琐,以后有时间考虑优化一下,暂时这样
			DB_LinksToken[索引].Note = Ser_AppUser.Uid取备注(DB_LinksToken[索引].LoginAppid, DB_LinksToken[索引].Uid)
		}
	}

	response.OkWithDetailed(结构响应_GetLinkUserList{
		List:  DB_LinksToken,
		Count: 总数,
	}, "获取成功", c)
	return
}

type 结构请求_GetLinkUserList struct {
	Page     int    `json:"Page"`     // 页
	Size     int    `json:"Size"`     // 页数量
	Status   int    `json:"Status"`   // 状态id
	Tourist  int    `json:"Tourist"`  // 游客  0 包含 1排除
	Type     int    `json:"Type"`     // 关键字类型  1 id 2 用户名 3绑定信息 4 动态标签
	Keywords string `json:"Keywords"` // 关键字
	Order    int    `json:"Order"`    // 0 倒序 1 正序
	AppId    int    `json:"AppId"`    //
}
type 结构响应_GetLinkUserList struct {
	List  interface{} `json:"List"`  // 列表
	Count int64       `json:"Count"` // 总数
}

type DB_LinksToken2 struct {
	DB.DB_LinksToken
	AppName string `json:"AppName"` //登录平台App名字
	Note    string `json:"Note"`    //软件用户备注
}

// 创建webApi使用的Token
func (a *LinkUserApi) NewWebApiToken(c *gin.Context) {
	var 请求 DB.DB_LinksToken
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	在线信息, err := Ser_LinkUser.NewWebApiToken(请求.OutTime, 请求.Key, 请求.Tab)
	if err != nil {
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithData(在线信息, c)
	return
}

type 结构响应_NewWebApiToken struct {
	Id      []int `json:"Id"`
	OutTime int   `json:"OutTime"`
}

// 修改令牌自动注销时间
func (a *LinkUserApi) SetTokenOutTime(c *gin.Context) {
	var 请求 结构响应_NewWebApiToken
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if len(请求.Id) == 0 {
		response.FailWithMessage("id数量不能为0", c)
		return
	}
	err = Ser_LinkUser.Set自动注销超时时间(请求.OutTime, 请求.Id)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(c)
	return
}

// Del批量注销
func (a *LinkUserApi) Del批量注销(c *gin.Context) {
	var 请求 结构请求_ID数组
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if len(请求.Id) == 0 {
		response.FailWithMessage("Id数组为空", c)
		return
	}

	err = Ser_LinkUser.Set批量注销(请求.Id, Ser_LinkUser.Z注销_管理员手动注销)
	if err != nil {
		response.FailWithMessage("注销失败", c)
		global.GVA_LOG.Error("Del批量注销:" + err.Error())
		return
	}

	response.OkWithMessage("注销成功", c)
	return
}

type 结构请求_ID数组 struct {
	Id []int `json:"Id"` //要注销的id数组
}

// Del批量删除已注销
func (a *LinkUserApi) Del批量删除已注销(c *gin.Context) {
	var 请求 结构请求_ID数组
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if len(请求.Id) == 0 {
		response.FailWithMessage("Id数组为空", c)
		return
	}

	if 请求.Id[0] == -1 {
		err = global.GVA_DB.Model(DB.DB_LinksToken{}).Where("Status = 2").Delete("").Error
	} else {
		err = global.GVA_DB.Model(DB.DB_LinksToken{}).Where("Id IN ? ", 请求.Id).Where("Status = 2").Delete("").Error
	}

	if err != nil {
		response.FailWithMessage("已注销删除失败", c)
		global.GVA_LOG.Error("Del批量删除已注销:" + err.Error())
		return
	}

	response.OkWithMessage("已注销删除成功", c)
	return
}
