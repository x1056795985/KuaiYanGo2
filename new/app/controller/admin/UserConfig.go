package controller

import (
	"github.com/gin-gonic/gin"
	"server/Service/Ser_Admin"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_AppUser"
	"server/Service/Ser_UserConfig"
	"server/global"
	"server/new/app/controller/Common"
	"server/structs/Http/response"
	DB "server/structs/db"
	"strconv"
	"time"
)

type UserConfig struct {
	Common.Common
}

func NewUserConfigController() *UserConfig {
	return &UserConfig{}
}

// 请求结构体（与旧架构完全一致）
type 请求_UserConfigGetInfo struct {
	AppId int    `json:"appId"`
	Uid   int    `json:"uid"`
	Name  string `json:"name"`
}

type 请求_UserConfigGetList struct {
	AppId          int    `json:"appId"`
	Page           int    `json:"page"`
	Size           int    `json:"size"`
	Type           int    `json:"type"`
	Keywords       string `json:"keywords"`
	Order          int    `json:"order"`
	PublicDataType []int  `json:"publicDataType"`
}

type 请求_UserConfigDelete struct {
	Data []DB.DB_UserConfig `json:"data"`
}

type 请求_UserConfigNew struct {
	DB.DB_UserConfig
}

type 请求_UserConfigSet struct {
	DB.DB_UserConfig
}

// 响应结构体
type 响应_UserConfigGetList struct {
	List  []响应_UserConfig扩展 `json:"list"`
	Count int64               `json:"count"`
}

type 响应_UserConfig扩展 struct {
	DB.DB_UserConfig
	AppName string `json:"appName"`
}

// Info 获取用户云配置详情
func (C *UserConfig) Info(c *gin.Context) {
	var 请求 请求_UserConfigGetInfo
	if !C.ToJSON(c, &请求) {
		return
	}

	var DB_UserConfig DB.DB_UserConfig
	Ser_UserConfig.Q取值(请求.AppId, 请求.Uid, 请求.Name)
	err := global.GVA_DB.Model(DB.DB_UserConfig{}).Where("AppId= ?", 请求.AppId).Where("Name= ?", 请求.Name).First(&DB_UserConfig).Error
	if err != nil {
		response.FailWithMessage("获取公共变量失败,可能联合主键不存在", c)
		return
	}
	response.OkWithDetailed(DB_UserConfig, "获取成功", c)
}

// GetList 获取用户云配置列表
func (C *UserConfig) GetList(c *gin.Context) {
	var 请求 请求_UserConfigGetList
	if !C.ToJSON(c, &请求) {
		return
	}

	局_DB := global.GVA_DB.Model(&DB.DB_UserConfig{})
	局_DB = 局_DB.Where("Uid>?", 0)
	if 请求.AppId > 0 {
		局_DB = 局_DB.Where("AppId=?", 请求.AppId)
	}

	if 请求.Order == 1 {
		局_DB = 局_DB.Order("Time ASC")
	} else if 请求.Order == 2 {
		局_DB = 局_DB.Order("Time DESC")
	}

	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1:
			局_DB = 局_DB.Where("LOCATE( ?, Name)>0 ", 请求.Keywords)
		case 2:
			局_DB = 局_DB.Where("LOCATE( ?, User)>0 ", 请求.Keywords)
		case 3:
			局_DB = 局_DB.Where("Uid = ?", 请求.Keywords)
		}
	}

	var DB_PublicData []响应_UserConfig扩展
	var 总数 int64
	err := 局_DB.Count(&总数).Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Omit("AppName").Find(&DB_PublicData).Error
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		global.GVA_LOG.Error("UserConfigGetList:" + err.Error())
		return
	}

	var AppName = Ser_AppInfo.App取map列表String(true)
	AppName["50"] = "代理云配置"

	var AdminIdNameMap = make(map[int]string)
	for 索引 := range DB_PublicData {
		DB_PublicData[索引].AppName = AppName[strconv.Itoa(DB_PublicData[索引].AppId)]
		if DB_PublicData[索引].AppId == 1 {
			if AdminIdNameMap[DB_PublicData[索引].Uid] == "" {
				AdminIdNameMap[DB_PublicData[索引].Uid] = Ser_Admin.Id取User(DB_PublicData[索引].Uid)
			}
			DB_PublicData[索引].User = AdminIdNameMap[DB_PublicData[索引].Uid]
			DB_PublicData[索引].Uid = -DB_PublicData[索引].Uid
		}
	}

	response.OkWithDetailed(响应_UserConfigGetList{DB_PublicData, 总数}, "获取成功", c)
}

// Delete 批量删除用户云配置
func (C *UserConfig) Delete(c *gin.Context) {
	var 请求 请求_UserConfigDelete
	if !C.ToJSON(c, &请求) {
		return
	}

	if len(请求.Data) == 0 {
		response.FailWithMessage("数组为空", c)
		return
	}

	var db = global.GVA_DB
	影响行数 := db.Model(DB.DB_UserConfig{}).Delete(请求.Data).RowsAffected
	if db.Error != nil {
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
}

// New 新建用户云配置
func (C *UserConfig) New(c *gin.Context) {
	var 请求 DB.DB_UserConfig
	if !C.ToJSON(c, &请求) {
		return
	}
	if 请求.Name == "" {
		response.FailWithMessage("变量名不能为空", c)
		return
	}
	if 请求.AppId <= 0 {
		response.FailWithMessage("AppId错误", c)
		return
	}
	if !Ser_AppUser.Uid是否存在(请求.AppId, 请求.Uid) {
		response.FailWithMessage("软件用户不存在", c)
		return
	}
	if Ser_UserConfig.Name是否存在(请求.AppId, 请求.Uid, 请求.Name) {
		response.FailWithMessage("变量名已存在", c)
		return
	}

	请求.Time = time.Now().Unix()
	请求.UpdateTime = time.Now().Unix()
	请求.User = Ser_AppUser.Uid取User(请求.AppId, 请求.Uid)

	err := Ser_UserConfig.C创建(请求)
	if err != nil {
		response.FailWithMessage("添加失败", c)
		return
	}
	response.OkWithMessage("添加成功", c)
}

// SetUserConfig 修改用户云配置值
func (C *UserConfig) SetUserConfig(c *gin.Context) {
	var 请求 DB.DB_UserConfig
	if !C.ToJSON(c, &请求) {
		return
	}
	if 请求.Uid <= 0 {
		response.FailWithMessage("管理平台配置禁止修改", c)
		return
	}
	if 请求.Name == "" {
		response.FailWithMessage("变量名不能为空", c)
		return
	}
	if !Ser_UserConfig.Name是否存在(请求.AppId, 请求.Uid, 请求.Name) {
		response.FailWithMessage("配置不存在", c)
		return
	}

	err := Ser_UserConfig.Z置值(请求.AppId, 请求.Uid, 请求.Name, 请求.Value)
	if err != nil {
		response.FailWithMessage("保存失败", c)
		return
	}
	response.OkWithMessage("保存成功", c)
}
