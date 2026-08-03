package controller

import (
	"EFunc/utils"
	"github.com/gin-gonic/gin"
	"server/app/controller/Common"
	"server/app/controller/Common/response"
	"server/app/global"
	dbm "server/app/models/db"
	"server/app/models/request"
	. "server/app/models/response"
	"server/app/service"
	"strconv"
)

type LuckyWheelLog struct {
	Common.Common
}

func NewLuckyWheelLogController() *LuckyWheelLog {
	return &LuckyWheelLog{}
}

func (C *LuckyWheelLog) GetList(c *gin.Context) {
	var 请求 struct {
		request.List2
		AppId        int     `json:"appId"`
		UserId       int     `json:"userId"`
		RegisterTime []int64 `json:"registerTime"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	var 开始时间, 结束时间 int64
	if 请求.RegisterTime != nil && len(请求.RegisterTime) == 2 && 请求.RegisterTime[0] > 0 && 请求.RegisterTime[1] > 0 {
		开始时间 = 请求.RegisterTime[0]
		结束时间 = 请求.RegisterTime[1]
	}

	tx := *global.GVA_DB
	var dataList []dbm.DB_LuckyWheelLog
	var 总数 int64
	var err error
	总数, dataList, err = service.NewLuckyWheelLog(c, &tx).GetList(请求.List2, 请求.AppId, 请求.UserId, 开始时间, 结束时间)
	if err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}

	if len(dataList) == 0 {
		response.OkWithDetailed(c, GetList2{List: dataList, Count: 总数}, "操作成功")
		return
	}

	//获取用户的name
	var 缓存 map[int]string
	var ids []int
	ids = make([]int, 0, len(dataList))
	for _, v := range dataList {
		ids = append(ids, v.UserId)
	}
	ids = utils.S数组_去重复(ids)

	userInfos, err := service.NewUser(c, &tx).Infos(map[string]interface{}{"Id": ids})
	if err != nil {
		response.FailWithMessage(c, "获取用户信息失败")
		return
	}
	缓存 = make(map[int]string, len(userInfos))
	for _, v := range userInfos {
		缓存[v.Id] = v.User
	}

	type item struct {
		dbm.DB_LuckyWheelLog
		Name string `json:"name"`
	}
	var 响应 = make([]item, 0, len(dataList))
	for _, v := range dataList {
		局_用户名 := "已删除"
		if 临时, ok := 缓存[v.UserId]; ok {
			局_用户名 = 临时
		}
		响应 = append(响应, item{
			DB_LuckyWheelLog: v,
			Name:             局_用户名,
		})
	}

	response.OkWithDetailed(c, GetList2{List: 响应, Count: 总数}, "操作成功")
	return
}

// Delete
func (C *LuckyWheelLog) Delete(c *gin.Context) {
	var 请求 struct {
		Ids []int `json:"ids" binding:"required,min=1"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	tx := *global.GVA_DB
	var err error
	var 影响行数 int64
	影响行数, err = service.NewLuckyWheelLog(c, &tx).Delete(请求.Ids)

	if err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}
	response.OkWithMessage(c, "删除成功,数量"+strconv.FormatInt(影响行数, 10))
}
