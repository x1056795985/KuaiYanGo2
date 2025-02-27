package controller

import (
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/service"
	"server/structs/Http/response"
	"strconv"
)

type AppInfo struct {
	Common.Common
}

func NewAppInfoController() *AppInfo {
	return &AppInfo{}
}

// 修改app排序
func (C *AppInfo) SetAppSort(c *gin.Context) {
	var 请求 struct {
		AppId int   `json:"AppId"`
		Sort  int64 `json:"Sort"`
	}
	//解析失败
	if !C.ToJSON(c, &请求) {
		return
	}
	tx := *global.GVA_DB
	var S = service.NewAppInfo(c, &tx)

	row, err := S.Update(请求.AppId, map[string]interface{}{"Sort": 请求.Sort})
	if err != nil {
		response.FailWithMessage(err.Error(), c)
	}
	response.OkWithMessage("操作成功,数量:"+strconv.Itoa(int(row)), c)
	return
}
