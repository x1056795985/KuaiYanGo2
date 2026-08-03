package controller

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"server/app/controller/Common"
	"server/app/controller/Common/response"
	"server/app/global"
	dbm "server/app/models/db"
	"server/app/models/request"
	. "server/app/models/response"
	"server/app/service"
	"time"
)

type LuckyWheelInfo struct {
	Common.Common
}

func NewLuckyWheelInfoController() *LuckyWheelInfo {
	return &LuckyWheelInfo{}
}

type prizeListItem struct {
	KaClassId   int    `json:"kaClassId"`
	Probability int    `json:"probability"`
	Name        string `json:"name"`
}

// GetList
func (C *LuckyWheelInfo) GetList(c *gin.Context) {
	var 请求 struct {
		request.List
		AppId         int `json:"appId"`
		Status        int `json:"status"`
		PromotionType int `json:"promotionType"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	tx := *global.GVA_DB
	var S = service.NewLuckyWheelInfo(c, &tx)
	var dataList []dbm.DB_LuckyWheelInfo
	var 总数 int64
	var err error
	总数, dataList, err = S.GetList(请求.List)
	if err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}
	response.OkWithDetailed(c, GetList2{List: dataList, Count: 总数}, "操作成功")
	return
}

// Update
func (C *LuckyWheelInfo) Update(c *gin.Context) {
	var 请求 struct {
		request.Id2
		DailyFreeCount  int              `json:"dailyFreeCount" binding:"min=0" zh:"每日免费次数"`
		InviteGiveCount int              `json:"inviteGiveCount" binding:"min=0" zh:"拉新奖励次数"`
		PrizeList       []prizeListItem  `json:"prizeList" binding:"" zh:"奖品列表"`
		ThemeColor      string           `json:"themeColor"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	tx := *global.GVA_DB
	局_奖品列表文本 := "[]"
	marshal, err := json.Marshal(请求.PrizeList)
	if err == nil {
		局_奖品列表文本 = string(marshal)
	}

	_, err = service.NewLuckyWheelInfo(c, &tx).UpdateMap([]int{请求.Id2.Id}, map[string]interface{}{
		"dailyFreeCount":  请求.DailyFreeCount,
		"inviteGiveCount": 请求.InviteGiveCount,
		"prizeList":       局_奖品列表文本,
		"themeColor":      请求.ThemeColor,
		"updateTime":      time.Now().Unix(),
	})

	if err != nil {
		response.FailWithMessage(c, err.Error())
	} else {
		response.OkWithMessage(c, "操作成功")
	}
}

// Info
func (C *LuckyWheelInfo) Info(c *gin.Context) {
	var 请求 request.Id
	if !C.ToJSON(c, &请求) {
		return
	}

	tx := *global.GVA_DB
	var S = service.NewLuckyWheelInfo(c, &tx)
	var info dbm.DB_LuckyWheelInfo
	info, err := S.Info(请求.Id)
	if err != nil {
		response.FailWithMessage(c, err.Error())
	} else {
		var 响应 struct {
			Id              int              `json:"id"`
			CreateTime      int64            `json:"createTime"`
			UpdateTime      int64            `json:"updateTime"`
			DailyFreeCount  int              `json:"dailyFreeCount"`
			InviteGiveCount int              `json:"inviteGiveCount"`
			PrizeList       []prizeListItem  `json:"prizeList"`
			ThemeColor      string           `json:"themeColor"`
		}
		响应.Id = info.Id
		响应.CreateTime = info.CreateTime
		响应.UpdateTime = info.UpdateTime
		响应.DailyFreeCount = info.DailyFreeCount
		响应.InviteGiveCount = info.InviteGiveCount
		响应.ThemeColor = info.ThemeColor
		_ = json.Unmarshal([]byte(info.PrizeList), &响应.PrizeList)

		response.OkWithDetailed(c, 响应, "操作成功")
	}
}
