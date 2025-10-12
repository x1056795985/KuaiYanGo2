package controller

import (
	. "EFunc/utils"
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/controller/Common/response"
	dbm "server/new/app/models/db"
	"server/new/app/models/request"
	. "server/new/app/models/response"
	"server/new/app/service"
	DB "server/structs/db"

	"strconv"
)

type CpsPayOrder struct {
	Common.Common
}

func NewCpsPayOrderController() *CpsPayOrder {
	var C = CpsPayOrder{}
	return &C
}

// Delete
// @action 删除
// @show  2
func (C *CpsPayOrder) Delete(c *gin.Context) {
	var 请求 request.Ids
	if !C.ToJSON(c, &请求) {
		return
	}

	var 影响行数 int64
	var err error
	tx := *global.GVA_DB
	影响行数, err = service.NewCpsPayOrder(c, &tx).Delete(请求.Ids)

	if err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}
	response.OkWithMessage(c, "删除成功,数量"+strconv.FormatInt(影响行数, 10))
	return
}

func (C *CpsPayOrder) SerNote(c *gin.Context) {
	var 请求 struct {
		Ids  []int  `json:"ids"`  //用户id数组
		Note string `json:"Note"` //
	}
	//解析失败
	if !C.ToJSON(c, &请求) {
		return
	}
	if len(请求.Ids) <= 0 {
		response.FailWithMessage(c, "id数量必须大于0")
		return
	}

	tx := *global.GVA_DB
	_, err := service.NewCpsPayOrder(c, &tx).UpdateMap(请求.Ids, map[string]interface{}{
		"note": 请求.Note,
	})
	if err != nil {
		response.FailWithMessage(c, err.Error())
	}
	response.OkWithMessage(c, "操作成功")
	return
}

func (C *CpsPayOrder) Info(c *gin.Context) {
	var 请求 request.Id
	if !C.ToJSON(c, &请求) {
		return
	}

	tx := *global.GVA_DB
	var info dbm.DB_CpsPayOrder
	info, err := service.NewCpsPayOrder(c, &tx).Info(请求.Id)
	if err != nil {
		response.FailWithMessage(c, err.Error())
	}
	response.OkWithDetailed(c, info, "操作成功")
	return
}

func (C *CpsPayOrder) GetList(c *gin.Context) {
	var 请求 struct {
		request.List
		RegisterTime []string `json:"RegisterTime"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	tx := *global.GVA_DB
	var dataList []dbm.DB_CpsPayOrder
	var 总数 int64
	var err error
	总数, dataList, err = service.NewCpsPayOrder(c, &tx).GetList(请求.List, 请求.RegisterTime)
	if err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}
	var info struct {
		ids         []int
		userInfo    []DB.DB_User
		map_id_user map[int]string

		appIds         []int
		appInfo        []DB.DB_AppInfo
		map_appid_name map[int]string
	}
	info.ids = make([]int, 0, len(dataList)*3)
	for i := 0; i < len(dataList); i++ {
		info.appIds = append(info.appIds, dataList[i].AppId)
		info.ids = append(info.ids, dataList[i].Uid)
		info.ids = append(info.ids, dataList[i].GrandpaId)
		info.ids = append(info.ids, dataList[i].InviterId)
	}
	info.ids = S数组_去重复(info.ids)
	info.appIds = S数组_去重复(info.appIds)

	info.userInfo, err = service.NewUser(c, &tx).Infos(map[string]interface{}{"Id": info.ids})
	if err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}
	info.map_id_user = make(map[int]string, len(info.userInfo))
	for i := 0; i < len(info.userInfo); i++ {
		info.map_id_user[info.userInfo[i].Id] = info.userInfo[i].User
	}
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
		dbm.DB_CpsPayOrder
		User    string `json:"user"`
		Grandpa string `json:"grandpa"`
		Inviter string `json:"inviter"`
		AppName string `json:"appName"`
	}, len(dataList))

	for i := 0; i < len(dataList); i++ {
		响应数据[i].DB_CpsPayOrder = dataList[i]
		if 临时数据, ok := info.map_id_user[dataList[i].Uid]; ok {
			响应数据[i].User = 临时数据
		}
		if 临时数据, ok := info.map_id_user[dataList[i].GrandpaId]; ok {
			响应数据[i].Grandpa = 临时数据
		}
		if 临时数据, ok := info.map_id_user[dataList[i].InviterId]; ok {
			响应数据[i].Inviter = 临时数据
		}
		if 临时数据, ok := info.map_appid_name[dataList[i].AppId]; ok {
			响应数据[i].AppName = 临时数据
		}
	}

	response.OkWithDetailed(c, GetList{List: 响应数据, Count: 总数}, "操作成功")
	return

}
