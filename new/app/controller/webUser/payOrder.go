package controller

import (
	. "EFunc/utils"
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/encoding/gjson"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/controller/Common/response"
	"server/new/app/models/request"
	. "server/new/app/models/response"
	"server/new/app/service"
	DB "server/structs/db"
	"strconv"
)

type PayOrder struct {
	Common.Common
}

func NewPayOrderController() *PayOrder {
	return &PayOrder{}
}

func (C *PayOrder) List(c *gin.Context) {
	var 请求 struct {
		Page int `json:"page" binding:"required"` // 页
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	var info = struct {
		appInfo  DB.DB_AppInfo
		likeInfo DB.DB_LinksToken
		数组订单     []DB.DB_LogRMBPayOrder
	}{}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	tx := *global.GVA_DB
	var 请求2 = request.List{
		Size:     10,
		Page:     请求.Page,
		Keywords: strconv.Itoa(info.likeInfo.Uid),      //直接固定只能查自己的
		Type:     S三元(info.appInfo.AppType <= 2, 2, 3), // 账号模式 2 卡号模式 3
	}
	var 总数 int64
	var err error
	总数, info.数组订单, err = service.NewPayOrder(c, &tx).GetList(请求2)
	if err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}
	var list = make([]LogRMBPayOrder_简单, 0, len(info.数组订单))
	for _, v := range info.数组订单 {
		list = append(list, LogRMBPayOrder_简单{
			Id:             v.Id,
			PayOrder:       v.PayOrder,
			PayOrder2:      v.PayOrder2,
			Status:         v.Status,
			Type:           v.Type,
			ProcessingType: v.ProcessingType,
			Rmb:            v.Rmb,
			Time:           v.Time,
			KaClassName:    gjson.New(v.Extra).Get("KaClassName").String(),
		})

	}

	response.OkWithData(c, GetList2{List: list, Count: 总数})
}

type LogRMBPayOrder_简单 struct {
	Id             int     `json:"id"`
	PayOrder       string  `json:"payOrder" gorm:"column:PayOrder;size:191;index;comment:余额充值订单id"`
	PayOrder2      string  `json:"payOrder2" gorm:"column:PayOrder2;size:191;comment:第三方订单id"`
	Status         int     `json:"status" gorm:"column:Status;comment:订单状态"`                         // 1  '等待支付'  2  '已付待充' 3 '充值成功' 4 退款中 5 ? 退款失败" : 6退款成功 7 订单关闭
	Type           string  `json:"type" gorm:"column:Type;size:191;comment:支付类型"`                    //  支付宝PC  微信支付 管理员手动充值 小叮当
	ProcessingType int     `json:"processingType" gorm:"column:ProcessingType;size:20;comment:处理类型"` //  0 余额充值 1 购卡直冲
	Rmb            float64 `json:"rmb" gorm:"column:Rmb;type:decimal(10,2);default:0;comment:充值金额"`
	Time           int64   `json:"time" gorm:"column:Time;index;comment:时间"`
	KaClassName    string  `json:"kaClassName"`
}
