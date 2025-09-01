package controller

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/controller/Common/response"
	dbm "server/new/app/models/db"
	"server/new/app/models/request"
	. "server/new/app/models/response"
	"server/new/app/service"
	"time"
)

type CheckInInfo struct {
	Common.Common
}

func NewCheckInInfoController() *CheckInInfo {
	return &CheckInInfo{}
}

// GetList
func (C *CheckInInfo) GetList(c *gin.Context) {
	//{"AppId":2,"Type":2,"Size":10,"Page":1,"Status":1,"keywords":"1"}
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
	var S = service.NewCheckInInfo(c, &tx)
	var dataList []dbm.DB_CheckInInfo
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

type cardClassListItem struct {
	Id     int `json:"id"`
	Points int `json:"p"` //节省占用长度
}

// Update
// @action 更新
// @show  2
func (C *CheckInInfo) Update(c *gin.Context) {
	var 请求 struct {
		request.Id2
		ShareGivePoints  int                 `json:"shareGivePoints" binding:"min=0" zh:"分享任务赠送签到分"` //虽然铜牌居然还需要有数量限制,但是确实是这样的,比如快快网络,前两个推广是不给佣金的,说明有需求,可能是防止自己开小号成为自己的推广者
		InviteGivePoints int                 `json:"inviteGivePoints" binding:"min=0" zh:"邀请任务赠送签到分"`
		CardClassList    []cardClassListItem `json:"cardClassList" binding:"" zh:"兑换卡类列表"`
	}
	//解析失败
	if !C.ToJSON(c, &请求) {
		return
	}
	tx := *global.GVA_DB
	局_临时文本 := "[]"
	marshal, err := json.Marshal(请求.CardClassList)
	if err == nil {
		局_临时文本 = string(marshal)
	}

	_, err = service.NewCheckInInfo(c, &tx).UpdateMap([]int{请求.Id2.Id}, map[string]interface{}{
		"shareGivePoints":  请求.ShareGivePoints,
		"inviteGivePoints": 请求.InviteGivePoints,
		"cardClassList":    局_临时文本,
		"updateTime":       time.Now().Unix(),
	})

	if err != nil {
		response.FailWithMessage(c, err.Error())
	} else {
		response.OkWithMessage(c, "操作成功")
	}

}

// Info
// @action 查询
// @show  2
func (C *CheckInInfo) Info(c *gin.Context) {
	var 请求 request.Id
	if !C.ToJSON(c, &请求) {
		return
	}

	tx := *global.GVA_DB
	var S = service.NewCheckInInfo(c, &tx)
	var info dbm.DB_CheckInInfo
	info, err := S.Info(请求.Id)
	if err != nil {
		response.FailWithMessage(c, err.Error())
	} else {
		var 响应 struct {
			Id               int                 `json:"id"  `
			CreateTime       int64               `json:"createTime" `
			UpdateTime       int64               `json:"updateTime" `
			ShareGivePoints  int                 `json:"shareGivePoints"`
			InviteGivePoints int                 `json:"inviteGivePoints"`
			CardClassList    []cardClassListItem `json:"cardClassList"`
		}
		响应.Id = info.Id
		响应.CreateTime = info.CreateTime
		响应.UpdateTime = info.UpdateTime
		响应.ShareGivePoints = info.ShareGivePoints
		响应.InviteGivePoints = info.InviteGivePoints
		_ = json.Unmarshal([]byte(info.CardClassList), &响应.CardClassList)

		response.OkWithDetailed(c, 响应, "操作成功")
	}

}
