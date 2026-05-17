package controller

import (
	"EFunc/utils"
	"github.com/gin-gonic/gin"
	Db服务 "server/Service/Ser_AppInfo"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/logic/common/publicData"
	"server/structs/Http/response"
	DB "server/structs/db"
	"strconv"
	"time"
)

type PublicDataCtrl struct {
	Common.Common
}

func NewPublicDataController() *PublicDataCtrl {
	return &PublicDataCtrl{}
}

type 请求_PublicDataGetInfo struct {
	AppId int    `json:"appId"`
	Name  string `json:"name"`
}

type 请求_PublicDataGetList struct {
	AppId          int    `json:"appId"`
	Page           int    `json:"page"`
	Size           int    `json:"size"`
	Type           int    `json:"type"`
	Keywords       string `json:"keywords"`
	Order          int    `json:"order"`
	PublicDataType []int  `json:"publicDataType"`
}

type 请求_PublicDataDelete struct {
	Data []DB.DB_PublicData `json:"data"`
}

type 请求_PublicDataSave struct {
	DB.DB_PublicData
}

type 请求_PublicDataNew struct {
	DB.DB_PublicData
}

type 请求_PublicDataSetVip struct {
	Name  []string `json:"name"`
	IsVip int      `json:"isVip"`
	AppID int      `json:"appID"`
}

type 响应_PublicDataGetList struct {
	List  []响应_PublicData扩展 `json:"list"`
	Count int64                `json:"count"`
}

type 响应_PublicData扩展 struct {
	DB.DB_PublicData
	AppName string `json:"appName"`
}

// Info 获取公共变量详情
func (C *PublicDataCtrl) Info(c *gin.Context) {
	var 请求 请求_PublicDataGetInfo
	if !C.ToJSON(c, &请求) {
		return
	}

	var DB_PublicData DB.DB_PublicData
	err := global.GVA_DB.Model(DB.DB_PublicData{}).Where("AppId= ?", 请求.AppId).Where("Name= ?", 请求.Name).First(&DB_PublicData).Error
	if err != nil {
		response.FailWithMessage("获取公共变量失败,可能联合主键不存在", c)
		return
	}
	response.OkWithDetailed(DB_PublicData, "获取成功", c)
}

// GetList 获取公共变量列表
func (C *PublicDataCtrl) GetList(c *gin.Context) {
	var 请求 请求_PublicDataGetList
	if !C.ToJSON(c, &请求) {
		return
	}

	局_DB := global.GVA_DB.Model(DB.DB_PublicData{})
	if 请求.AppId > 0 {
		局_DB.Where("AppId=?", 请求.AppId)
	}
	if 请求.Order == 1 {
		局_DB.Order("Time ASC")
	} else if 请求.Order == 2 {
		局_DB.Order("Time DESC")
	}
	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1:
			局_DB.Where("LOCATE( ?, Name)>0 ", 请求.Keywords)
		}
	}
	if len(请求.PublicDataType) > 0 {
		switch 请求.Type {
		case 1:
			局_DB.Where("Type IN ? ", 请求.PublicDataType)
		}
	}

	var DB_PublicData []响应_PublicData扩展
	var 总数 int64
	err := 局_DB.Count(&总数).Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Omit("AppName").Find(&DB_PublicData).Error
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		global.GVA_LOG.Error("PublicDataGetList:" + err.Error())
		return
	}

	var AppName = Db服务.App取map列表String(false)
	AppName["1"] = "全局"

	for 索引 := range DB_PublicData {
		DB_PublicData[索引].AppName = AppName[strconv.Itoa(DB_PublicData[索引].AppId)]
		if DB_PublicData[索引].Type == 4 && utils.W文本_取长度(DB_PublicData[索引].Value) > 200 {
			DB_PublicData[索引].Value = utils.W文本_取左边(DB_PublicData[索引].Value, 200) + "..."
		}
	}

	response.OkWithDetailed(响应_PublicDataGetList{DB_PublicData, 总数}, "获取成功", c)
}

// Delete 批量删除公共变量
func (C *PublicDataCtrl) Delete(c *gin.Context) {
	var 请求 请求_PublicDataDelete
	if !C.ToJSON(c, &请求) {
		return
	}
	if len(请求.Data) == 0 {
		response.FailWithMessage("数组为空", c)
		return
	}

	var db = global.GVA_DB
	影响行数 := db.Model(DB.DB_PublicData{}).Delete(请求.Data).RowsAffected
	if db.Error != nil {
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
}

// SaveInfo 保存公共变量信息
func (C *PublicDataCtrl) SaveInfo(c *gin.Context) {
	var 请求 DB.DB_PublicData
	if !C.ToJSON(c, &请求) {
		return
	}
	if 请求.Name == "" {
		response.FailWithMessage("变量名不能为空", c)
		return
	}
	if !publicData.L_publicData.Name是否存在(&gin.Context{}, 请求.AppId, 请求.Name) {
		response.FailWithMessage("变量不存在", c)
		return
	}
	请求.Time = time.Now().Unix()
	err := publicData.L_publicData.Z置值_原值(c, 请求)
	if err != nil {
		response.FailWithMessage("保存失败"+err.Error(), c)
		return
	}
	response.OkWithMessage("保存成功", c)
}

// New 新建公共变量
func (C *PublicDataCtrl) New(c *gin.Context) {
	var 请求 DB.DB_PublicData
	if !C.ToJSON(c, &请求) {
		return
	}
	if 请求.Name == "" {
		response.FailWithMessage("变量名不能为空", c)
		return
	}
	if 请求.Type < 1 {
		response.FailWithMessage("变量类型错误", c)
		return
	}
	if publicData.L_publicData.Name是否存在(&gin.Context{}, 请求.AppId, 请求.Name) {
		response.FailWithMessage("变量名已存在", c)
		return
	}
	请求.Time = time.Now().Unix()
	err := publicData.L_publicData.C创建(c, 请求)
	if err != nil {
		response.FailWithMessage("添加失败", c)
		return
	}
	response.OkWithMessage("添加成功", c)
}

// SetVipLimit 批量修改vip限制
func (C *PublicDataCtrl) SetVipLimit(c *gin.Context) {
	var 请求 请求_PublicDataSetVip
	if !C.ToJSON(c, &请求) {
		return
	}
	if 请求.IsVip < 0 {
		response.FailWithMessage("IsVip值错误", c)
		return
	}
	if len(请求.Name) == 0 {
		response.FailWithMessage("公共变量数组为空", c)
		return
	}

	err := publicData.L_publicData.P批量修改IsVip(c, 请求.AppID, 请求.Name, 请求.IsVip)
	if err != nil {
		response.FailWithMessage("修改失败", c)
		global.GVA_LOG.Error("修改失败:" + err.Error())
		return
	}
	response.OkWithMessage("修改成功", c)
}
