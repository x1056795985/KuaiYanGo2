package controller

import (
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/controller/Common"
	dbm "server/new/app/models/db"
	"server/new/app/models/request"
	. "server/new/app/models/response"
	"server/new/app/service"
	"server/structs/Http/response"
	"time"
)

type CpsInfo struct {
	Common.Common
}

func NewCpsInfoController() *CpsInfo {
	return &CpsInfo{}
}

// GetList
func (C *CpsInfo) GetList(c *gin.Context) {
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
	var S = service.NewCpsInfo(c, &tx)
	var dataList []dbm.DB_CpsInfo
	var 总数 int64
	var err error
	总数, dataList, err = S.GetList(请求.List)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(GetList2{List: dataList, Count: 总数}, "操作成功", c)
	return
}

// Update
// @action 更新
// @show  2
func (C *CpsInfo) Update(c *gin.Context) {
	var 请求 struct {
		request.Id
		BronzeThreshold    int `json:"bronzeThreshold" binding:"" zh:"成为铜牌推广数量阈值"` //虽然铜牌居然还需要有数量限制,但是确实是这样的,比如快快网络,前两个推广是不给佣金的,说明有需求,可能是防止自己开小号成为自己的推广者
		BronzeKickback     int `json:"bronzeKickback" binding:"min=0,max=100" zh:"铜牌分成比例"`
		SilverThreshold    int `json:"silverThreshold" binding:"" zh:"成为银牌推广数量阈值"`
		SilverKickback     int `json:"silverKickback" binding:"min=0,max=100" zh:"银牌分成比例"`
		GoldMedalThreshold int `json:"goldMedalThreshold" binding:"" zh:"成为金牌推广数量阈值"`
		GoldMedalKickback  int `json:"goldMedalKickback" binding:"min=0,max=100" zh:"金牌分成比例"`
		GrandsonKickback   int `json:"grandsonKickback" binding:"min=0,max=100" zh:"徒孙分成比例"`
		//NarrowPic          string `json:"widePic" binding:"required" zh:"素材_窄图"`
		//DetailPic          string `json:"detailPic" binding:"required" zh:"素材_详情图"`
	}
	//解析失败
	if !C.ToJSON(c, &请求) {
		return
	}
	tx := *global.GVA_DB

	_, err := service.NewCpsInfo(c, &tx).UpdateMap([]int{请求.Id.Id}, map[string]interface{}{
		"bronzeThreshold":    请求.BronzeThreshold,
		"bronzeKickback":     请求.BronzeKickback,
		"silverThreshold":    请求.SilverThreshold,
		"silverKickback":     请求.SilverKickback,
		"goldMedalThreshold": 请求.GoldMedalThreshold,
		"goldMedalKickback":  请求.GoldMedalKickback,
		"grandsonKickback":   请求.GrandsonKickback,
		//"narrowPic":          请求.NarrowPic,
		//"detailPic":          请求.DetailPic,
		"updateTime": time.Now().Unix(),
	})

	if err != nil {
		response.FailWithMessage(err.Error(), c)
	} else {
		response.OkWithMessage("操作成功", c)
	}

}

// Info
// @action 查询
// @show  2
func (C *CpsInfo) Info(c *gin.Context) {
	var 请求 request.Id
	if !C.ToJSON(c, &请求) {
		return
	}

	tx := *global.GVA_DB
	var S = service.NewCpsInfo(c, &tx)
	var info dbm.DB_CpsInfo
	info, err := S.Info(请求.Id)
	if err != nil {
		response.FailWithMessage(err.Error(), c)

	} else {
		response.OkWithDetailed(info, "操作成功", c)
	}

}
