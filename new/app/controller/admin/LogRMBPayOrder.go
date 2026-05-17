package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"server/Service/Ser_Log"
	"server/Service/Ser_RMBPayOrder"
	"server/Service/Ser_User"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/logic/common/rmbPay"
	"server/new/app/models/common"
	"server/new/app/models/constant"
	"server/structs/Http/response"
	DB "server/structs/db"
	"strconv"
	"strings"
	"time"
)

type LogRMBPayOrderCtrl struct {
	Common.Common
}

func NewLogRMBPayOrderController() *LogRMBPayOrderCtrl {
	return &LogRMBPayOrderCtrl{}
}

type 请求_LogRMBPayOrderGetInfo struct {
	Id int `json:"id"`
}

type 请求_LogRMBPayOrderGetList struct {
	Page         int      `json:"page"`
	Size         int      `json:"size"`
	Type         int      `json:"type"`
	Status       int      `json:"status"`
	Keywords     string   `json:"keywords"`
	Order        int      `json:"order"`
	RegisterTime []string `json:"registerTime"`
	Count        int64    `json:"count"`
}

type 请求_LogRMBPayOrderDelete struct {
	Id       []int  `json:"id"`
	Type     int    `json:"type"`
	Keywords string `json:"keywords"`
}

type 请求_LogRMBPayOrderNew struct {
	User string  `json:"user"`
	RMB  float64 `json:"rMB"`
	Note string  `json:"note"`
}

type 请求_LogRMBPayOrderOut struct {
	PayOrder string `json:"payOrder"`
	IsOutRMB bool   `json:"isOutRMB"`
	Note     string `json:"note"`
}

type 请求_LogRMBPayOrderSetNote struct {
	PayOrder []string `json:"payOrder"`
	Note     string   `json:"note"`
}

type 响应_LogRMBPayOrderGetList struct {
	List  []响应_LogRMBPayOrderGetListItem `json:"list"`
	Count int64                           `json:"count"`
}

type 响应_LogRMBPayOrderGetListItem struct {
	DB.DB_LogRMBPayOrder
	Processing string `json:"processing"`
}

// Info 获取余额充值订单详情
func (C *LogRMBPayOrderCtrl) Info(c *gin.Context) {
	var 请求 请求_LogRMBPayOrderGetInfo
	if !C.ToJSON(c, &请求) {
		return
	}

	var DB_LogRMBPayOrder DB.DB_LogRMBPayOrder
	err := global.GVA_DB.Model(DB.DB_LogRMBPayOrder{}).Where("Id= ?", 请求.Id).First(&DB_LogRMBPayOrder).Error
	if err != nil {
		response.FailWithMessage("获取失败,可能不存在", c)
		return
	}
	response.OkWithDetailed(DB_LogRMBPayOrder, "获取成功", c)
}

// GetList 获取余额充值订单列表
func (C *LogRMBPayOrderCtrl) GetList(c *gin.Context) {
	var 请求 请求_LogRMBPayOrderGetList
	if !C.ToJSON(c, &请求) {
		return
	}

	局_DB := global.GVA_DB.Model(DB.DB_LogRMBPayOrder{})
	if 请求.Order == 1 {
		局_DB.Order("db_Log_RMBPayOrder.Id ASC")
	} else {
		局_DB.Order("db_Log_RMBPayOrder.Id DESC")
	}
	if 请求.RegisterTime != nil && len(请求.RegisterTime) == 2 && 请求.RegisterTime[0] != "" && 请求.RegisterTime[1] != "" {
		制卡开始时间, _ := strconv.ParseInt(请求.RegisterTime[0], 10, 64)
		制卡结束时间, _ := strconv.ParseInt(请求.RegisterTime[1], 10, 64)
		局_DB.Where("db_Log_RMBPayOrder.Time > ?", 制卡开始时间).Where("db_Log_RMBPayOrder.Time < ?", 制卡结束时间+86400)
	}
	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1:
			局_DB.Where("LOCATE( ?, db_Log_RMBPayOrder.User)>0 ", 请求.Keywords)
		case 2:
			局_DB.Where("LOCATE( ?, db_Log_RMBPayOrder.Note)>0 ", 请求.Keywords)
		case 3:
			局_DB.Where("db_Log_RMBPayOrder.Ip = ? ", 请求.Keywords)
		case 4:
			局_DB.Where("LOCATE( ?, db_Log_RMBPayOrder.PayOrder)>0 ", 请求.Keywords)
		case 5:
			局_DB.Where("LOCATE( ?, db_Log_RMBPayOrder.PayOrder2)>0 ", 请求.Keywords)
		case 6:
			局_DB.Where("db_Log_RMBPayOrder.Rmb = ? ", 请求.Keywords)
		}
	}
	if 请求.Status > 0 {
		局_DB.Where("db_Log_RMBPayOrder.Status  = ? ", 请求.Status)
	}

	var DB_LogRMBPayOrder []DB.DB_LogRMBPayOrder
	var 总数 int64
	if 请求.Count > 500000 {
		总数 = 请求.Count
	} else {
		局_DB.Count(&总数)
	}
	局_DB = 局_DB.Select("db_Log_RMBPayOrder.*")
	err := 局_DB.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&DB_LogRMBPayOrder).Error
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		return
	}

	var 局_响应 []响应_LogRMBPayOrderGetListItem
	局_响应 = make([]响应_LogRMBPayOrderGetListItem, len(DB_LogRMBPayOrder))
	for 索引 := range DB_LogRMBPayOrder {
		局_响应[索引].DB_LogRMBPayOrder = DB_LogRMBPayOrder[索引]
		局_响应[索引].Processing = Ser_RMBPayOrder.C处理类型[DB_LogRMBPayOrder[索引].ProcessingType]
		if 局_响应[索引].Processing == "" {
			局_响应[索引].Processing = "未知原因" + strconv.Itoa(DB_LogRMBPayOrder[索引].ProcessingType)
		}
		var tmp interface{}
		if err = json.Unmarshal([]byte(局_响应[索引].Extra), &tmp); err == nil {
			if formatted, err2 := json.MarshalIndent(tmp, "", "    "); err2 == nil {
				局_响应[索引].Extra = string(formatted)
			}
		}
		局_响应[索引].Extra = "<pre>" + strings.ReplaceAll(局_响应[索引].Extra, "\n", "<br />") + "<pre />"
	}
	response.OkWithDetailed(响应_LogRMBPayOrderGetList{局_响应, 总数}, "获取成功", c)
}

// Delete 批量删除余额充值订单
func (C *LogRMBPayOrderCtrl) Delete(c *gin.Context) {
	var 请求 请求_LogRMBPayOrderDelete
	if !C.ToJSON(c, &请求) {
		return
	}
	var 影响行数 int64
	var db = global.GVA_DB.Model(DB.DB_LogRMBPayOrder{})

	switch 请求.Type {
	default:
		response.FailWithMessage("Type错误", c)
		return
	case 1:
		if len(请求.Id) == 0 {
			response.FailWithMessage("Id数组没有要删除的ID", c)
			return
		}
		影响行数 = db.Where("Id IN ? ", 请求.Id).Delete(DB.DB_LogRMBPayOrder{}).RowsAffected
	case 2:
		影响行数 = db.Where("User = ? ", 请求.Keywords).Delete(DB.DB_LogRMBPayOrder{}).RowsAffected
	case 3:
		影响行数 = db.Where("1=1").Delete(DB.DB_LogRMBPayOrder{}).RowsAffected
	case 4:
		影响行数 = db.Where("Time <  ?", time.Now().Unix()-604800).Delete(DB.DB_LogRMBPayOrder{}).RowsAffected
	case 5:
		影响行数 = db.Where("Time <  ?", time.Now().Unix()-2592000).Delete(DB.DB_LogRMBPayOrder{}).RowsAffected
	case 6:
		影响行数 = db.Where("Time <  ?", time.Now().Unix()-7776000).Delete(DB.DB_LogRMBPayOrder{}).RowsAffected
	case 7:
		if len(请求.Keywords) == 0 {
			response.FailWithMessage("关键字不能为空", c)
			return
		}
		影响行数 = db.Where("LOCATE( ?, Note)>0 ", 请求.Keywords).Delete(请求.Id).RowsAffected
	case 8:
		状态id := []int{
			constant.D订单状态_等待支付,
			constant.D订单状态_已关闭,
		}
		影响行数 = db.Where("Time <  ?", time.Now().Unix()-3600).Where("Status IN ?", 状态id).Delete(DB.DB_LogRMBPayOrder{}).RowsAffected
		if db.Error == nil {
			response.OkWithMessage("删除过期(1小时前)待支付和已关闭成功,数量"+strconv.FormatInt(影响行数, 10), c)
			return
		}
	}
	if db.Error != nil {
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
}

// New 手动充值
func (C *LogRMBPayOrderCtrl) New(c *gin.Context) {
	var 请求 请求_LogRMBPayOrderNew
	if !C.ToJSON(c, &请求) {
		return
	}
	局_Uid := Ser_User.User用户名取id(请求.User)
	if 请求.User == "" || 局_Uid == 0 {
		response.FailWithMessage("用户不存在", c)
		return
	}
	if 请求.RMB > 1000000000 || 请求.RMB < -1000000000 {
		response.FailWithMessage("增减金额不能超过10亿(11位)", c)
	}

	var 新订单 DB.DB_LogRMBPayOrder
	新订单.Id = 0
	新订单.Uid = 局_Uid
	新订单.User = 请求.User
	新订单.Status = 2
	新订单.Time = time.Now().Unix()
	新订单.Ip = c.ClientIP()
	新订单.Type = "管理员手动充值"
	新订单.Rmb = 请求.RMB
	新订单.Note = 请求.Note
	新订单.PayOrder = Ser_RMBPayOrder.Get获取新订单号()
	新订单.UidType = 1

	err := global.GVA_DB.Model(DB.DB_LogRMBPayOrder{}).Create(&新订单).Error
	if err != nil {
		response.FailWithMessage("订单创建失败", c)
		return
	}
	var 新余额 float64
	if 新余额, err = Ser_User.Id余额增减(新订单.Uid, 新订单.Rmb, true); err != nil {
		response.FailWithMessage("订单创建成功充值用户失败", c)
		return
	}
	Ser_Log.Log_写余额日志(新订单.User, c.ClientIP(), fmt.Sprintf("管理员手动创建支付订单:%s|新余额≈%.2f", 新订单.PayOrder, 新余额), 新订单.Rmb)

	if !Ser_RMBPayOrder.Order更新订单状态(新订单.PayOrder, Ser_RMBPayOrder.D订单状态_成功) {
		response.FailWithMessage("用户充值成功订单状态更新失败", c)
		return
	}
	response.OkWithMessage("成功,为保证规范该接口后续将删除,后续请到用户列表,勾选->更多->on批量增减余额", c)
}

// Out 退款
func (C *LogRMBPayOrderCtrl) Out(c *gin.Context) {
	var 请求 请求_LogRMBPayOrderOut
	if !C.ToJSON(c, &请求) {
		return
	}
	var 参数 common.PayParams
	参数.PayOrder = 请求.PayOrder
	err := rmbPay.L_rmbPay.D订单退款(c, 参数, 请求.IsOutRMB, 请求.Note)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("操作成功", c)
}

// SetNote 批量修改备注
func (C *LogRMBPayOrderCtrl) SetNote(c *gin.Context) {
	var 请求 请求_LogRMBPayOrderSetNote
	if !C.ToJSON(c, &请求) {
		return
	}
	if len(请求.PayOrder) == 0 {
		response.FailWithMessage("订单数组为空", c)
		return
	}
	err := Ser_RMBPayOrder.Order更新订单备注_批量(请求.PayOrder, 请求.Note)
	if err != nil {
		response.FailWithMessage("修改失败", c)
		global.GVA_LOG.Error("修改失败:" + err.Error())
		return
	}
	response.OkWithMessage("修改成功", c)
}
