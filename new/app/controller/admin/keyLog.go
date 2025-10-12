package controller

import (
	. "EFunc/utils"
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/controller/Common/response"
	dbm "server/new/app/models/db"
	"server/new/app/models/request"
	response2 "server/new/app/models/response"
	"server/new/app/service"
	DB "server/structs/db"
	"strconv"
)

type LogKey struct {
	Common.Common
}

func NewLogKeyController() *LogKey {
	var C = LogKey{}
	return &C
}

// Delete
// @action 删除
// @show  2
func (C *LogKey) Delete(c *gin.Context) {
	var 请求 request.Ids
	if !C.ToJSON(c, &请求) {
		return
	}

	var 影响行数 int64
	var err error
	tx := *global.GVA_DB
	影响行数, err = service.NewLogKey(c, &tx).Delete(请求.Ids)

	if err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}
	response.OkWithMessage(c, "删除成功,数量"+strconv.FormatInt(影响行数, 10))
	return
}

func (C *LogKey) Info(c *gin.Context) {
	var 请求 request.Id
	if !C.ToJSON(c, &请求) {
		return
	}

	tx := *global.GVA_DB
	var info dbm.DB_LogKey
	info, err := service.NewLogKey(c, &tx).Info(请求.Id)
	if err != nil {
		response.FailWithMessage(c, err.Error())
	}
	response.OkWithDetailed(c, info, "操作成功")
	return
}

func (C *LogKey) GetList(c *gin.Context) {
	var 请求 struct {
		request.List
		RegisterTime []string `json:"RegisterTime"` // 制卡开始时间 制卡结束时间
		Appid        int      `json:"Appid"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	tx := *global.GVA_DB
	var dataList []dbm.DB_LogKey
	var 总数 int64
	var err error

	总数, dataList, err = service.NewLogKey(c, &tx).GetList(请求.List, 请求.Appid, 请求.RegisterTime)
	if err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}
	var info struct {
		appIds         []int
		appInfo        []DB.DB_AppInfo
		map_appid_name map[int]string
	}
	for i := 0; i < len(dataList); i++ {
		info.appIds = append(info.appIds, dataList[i].AppId)
	}

	info.appIds = S数组_去重复(info.appIds)
	info.appInfo, err = service.NewAppInfo(c, &tx).Infos(map[string]interface{}{"AppId": info.appIds})
	if err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}

	info.map_appid_name = make(map[int]string, len(info.appInfo))
	for i := 0; i < len(info.appInfo); i++ {
		info.map_appid_name[info.appInfo[i].AppId] = info.appInfo[i].AppName
	}
	var 响应数据 = make([]struct {
		dbm.DB_LogKey
		AppName string `json:"appName"`
	}, len(dataList))

	for i := 0; i < len(dataList); i++ {
		响应数据[i].DB_LogKey = dataList[i]
		if 临时数据, ok := info.map_appid_name[dataList[i].AppId]; ok {
			响应数据[i].AppName = 临时数据
		}
	}

	response.OkWithDetailed(c, response2.GetList2{List: 响应数据, Count: 总数}, "操作成功")
	return
}
