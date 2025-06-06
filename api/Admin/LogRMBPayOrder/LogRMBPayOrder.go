package LogRMBPayOrder

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"server/Service/Ser_Log"
	"server/Service/Ser_RMBPayOrder"
	"server/Service/Ser_User"
	"server/global"
	"server/new/app/logic/common/rmbPay"
	"server/new/app/models/common"
	"server/new/app/models/constant"
	"server/structs/Http/response"
	DB "server/structs/db"
	"strconv"
	"strings"
	"time"
)

type Api struct{}
type 结构请求_单id struct {
	Id int `json:"Id"`
}

// GetInfo
func (a *Api) GetInfo(c *gin.Context) {
	var 请求 结构请求_单id
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	var DB_LogRMBPayOrder DB.DB_LogRMBPayOrder

	err = global.GVA_DB.Model(DB.DB_LogRMBPayOrder{}).Where("Id= ?", 请求.Id).First(&DB_LogRMBPayOrder).Error
	// 没查到数据
	if err != nil {
		response.FailWithMessage("获取失败,可能不存在", c)
		return
	}
	response.OkWithDetailed(DB_LogRMBPayOrder, "获取成功", c)
	return
}

type 结构请求_GetDB_LogRMBPayOrderList struct {
	Page         int      `json:"Page"` // 页
	Size         int      `json:"Size"` // 页数量
	Type         int      `json:"Type"` // 关键字类型
	Status       int      `json:"Status"`
	Keywords     string   `json:"Keywords"`     // 关键字
	Order        int      `json:"Order"`        // 0 倒序 1 正序
	RegisterTime []string `json:"RegisterTime"` // 开始时间 结束时间
	Count        int64    `json:"Count"`        // 总数
}

// GetDB_LogRMBPayOrderList

func (a *Api) GetLogList2(c *gin.Context) {
	var 请求 结构请求_GetDB_LogRMBPayOrderList
	//{"Type":2,"Size":10,"Page":1,"Status":1,"keywords":"1"}
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	局_DB := global.GVA_DB.Model(DB.DB_LogRMBPayOrder{})
	if 请求.Order == 1 {
		局_DB.Order("db_Log_RMBPayOrder.Id ASC")
	} else {
		局_DB.Order("db_Log_RMBPayOrder.Id DESC")
	}
	//时间筛选=========
	if 请求.RegisterTime != nil && len(请求.RegisterTime) == 2 && 请求.RegisterTime[0] != "" && 请求.RegisterTime[1] != "" {
		制卡开始时间, _ := strconv.Atoi(请求.RegisterTime[0])
		制卡结束时间, _ := strconv.Atoi(请求.RegisterTime[1])
		局_DB.Where("db_Log_RMBPayOrder.Time > ?", 制卡开始时间).Where("db_Log_RMBPayOrder.Time < ?", 制卡结束时间+86400)
	}
	//关键字筛选
	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //订单ID
			局_DB.Where("LOCATE( ?, db_Log_RMBPayOrder.User)>0 ", 请求.Keywords)
		case 2: //消息
			局_DB.Where("LOCATE( ?, db_Log_RMBPayOrder.Note)>0 ", 请求.Keywords)
		case 3: //ip
			局_DB.Where("db_Log_RMBPayOrder.Ip = ? ", 请求.Keywords)
		case 4: //订单编号
			局_DB.Where("LOCATE( ?, db_Log_RMBPayOrder.PayOrder)>0 ", 请求.Keywords)
		case 5: //支付通道订单编号
			局_DB.Where("LOCATE( ?, db_Log_RMBPayOrder.PayOrder2)>0 ", 请求.Keywords)
		case 6: //金额
			局_DB.Where("db_Log_RMBPayOrder.Rmb = ? ", 请求.Keywords)
		}
	}

	if 请求.Status > 0 {
		局_DB.Where("db_Log_RMBPayOrder.Status  = ? ", 请求.Status)
	}

	var DB_LogRMBPayOrder []DB.DB_LogRMBPayOrder
	var 总数 int64
	//Count(&总数) 必须放在where 后面 不然值会被清0

	if 请求.Count > 500000 { //如果数据大于50万 直接使用,不重新查询了 优化速度
		总数 = 请求.Count
	} else {
		局_DB.Count(&总数)
	}
	局_DB = 局_DB.Select("db_Log_RMBPayOrder.*")
	err = 局_DB.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&DB_LogRMBPayOrder).Error
	//fmt.Println("用户总数%d", 总数, DB_LinksToken)
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		return
	}
	var 局_响应 []结构响应_GetDB_LogRMBPayOrderList_item
	局_响应 = make([]结构响应_GetDB_LogRMBPayOrderList_item, len(DB_LogRMBPayOrder))
	for 索引 := range DB_LogRMBPayOrder {
		局_响应[索引].DB_LogRMBPayOrder = DB_LogRMBPayOrder[索引]
		局_响应[索引].Processing = Ser_RMBPayOrder.C处理类型[DB_LogRMBPayOrder[索引].ProcessingType]
		if 局_响应[索引].Processing == "" {
			局_响应[索引].Processing = "未知原因" + strconv.Itoa(DB_LogRMBPayOrder[索引].ProcessingType)
		}
		// 如果Extra是JSON字符串
		var tmp interface{}
		if err = json.Unmarshal([]byte(局_响应[索引].Extra), &tmp); err == nil {
			if formatted, err2 := json.MarshalIndent(tmp, "", "    "); err2 == nil {
				局_响应[索引].Extra = string(formatted)
			}
		}
		//局_响应[索引].Extra = strings.ReplaceAll(局_响应[索引].Extra, "{\n", "{")
		//局_响应[索引].Extra = strings.ReplaceAll(局_响应[索引].Extra, "}\n", "}")
		//局_响应[索引].Extra = strings.ReplaceAll(局_响应[索引].Extra, "},\n", "}")
		局_响应[索引].Extra = "<pre>" + strings.ReplaceAll(局_响应[索引].Extra, "\n", "<br />") + "<pre />"
	}
	response.OkWithDetailed(结构响应_GetDB_LogRMBPayOrderList{局_响应, 总数}, "获取成功", c)
	return
}

type 结构响应_GetDB_LogRMBPayOrderList_item struct {
	DB.DB_LogRMBPayOrder
	Processing string `json:"Processing"`
}
type 结构响应_GetDB_LogRMBPayOrderList struct {
	List  interface{} `json:"List"`  // 列表
	Count int64       `json:"Count"` // 总数
}

type 结构响应_DB_LogRMBPayOrder_扩展 struct {
	DB.DB_LogRMBPayOrder
	User string `json:"User" gorm:"column:User;index;comment:用户名"`
}
type 结构请求_批量Delete struct {
	Id       []int  `json:"Id"`       //用户id数组
	Type     int    `json:"Type"`     //  1删除用户数组 2删除指定关键字 3清空 4删除7天前 5删除30天前 6删除90天前 7 关键字  8删除过期待支付
	Keywords string `json:"Keywords"` //
}

// Del批量删除
func (a *Api) Delete(c *gin.Context) {
	var 请求 结构请求_批量Delete
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	var 影响行数 int64
	var db = global.GVA_DB.Model(DB.DB_LogRMBPayOrder{})

	//1删除用户数组 2删除指定用户 3清空 4删除7天前 5删除30天前 6删除90天前  7 关键字
	switch 请求.Type {
	default:
		response.FailWithMessage("Type错误", c)
		return
	case 1:
		if 请求.Type == 1 && len(请求.Id) == 0 {
			response.FailWithMessage("Id数组没有要删除的ID", c)
			return
		}
		影响行数 = db.Where("Id IN ? ", 请求.Id).Delete(DB.DB_LogRMBPayOrder{}).RowsAffected
	case 2:
		影响行数 = db.Where("User = ? ", 请求.Keywords).Delete(DB.DB_LogRMBPayOrder{}).RowsAffected
	case 3: //清空
		影响行数 = db.Where("1=1").Delete(DB.DB_LogRMBPayOrder{}).RowsAffected
	case 4: //删7天前
		影响行数 = db.Where("Time <  ?", time.Now().Unix()-604800).Delete(DB.DB_LogRMBPayOrder{}).RowsAffected
	case 5: //删除30天前
		影响行数 = db.Where("Time <  ?", time.Now().Unix()-2592000).Delete(DB.DB_LogRMBPayOrder{}).RowsAffected
	case 6: //删除90天前
		影响行数 = db.Where("Time <  ?", time.Now().Unix()-7776000).Delete(DB.DB_LogRMBPayOrder{}).RowsAffected
	case 7: //删除关键字
		if len(请求.Keywords) == 0 {
			response.FailWithMessage("关键字不能为空", c)
			return
		}
		影响行数 = db.Where("LOCATE( ?, Note)>0 ", 请求.Keywords).Delete(请求.Id).RowsAffected
	case 8: //一小时前待支付
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
	return
}

// 手动充值
func (a *Api) New手动充值(c *gin.Context) {
	var 请求 结构请求_手动充值
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
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

	err = global.GVA_DB.Model(DB.DB_LogRMBPayOrder{}).Create(&新订单).Error
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

	response.OkWithMessage("操作成功", c)
	return
}

type 结构请求_手动充值 struct {
	User string  `json:"User"`
	RMB  float64 `json:"RMB"`
	Note string  `json:"Note"`
}

// 退款
func (a *Api) Out退款(c *gin.Context) {
	var 请求 结构请求_退款
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	var 参数 common.PayParams
	参数.PayOrder = 请求.PayOrder
	err = rmbPay.L_rmbPay.D订单退款(c, 参数, 请求.IsOutRMB, 请求.Note)
	// 根据支付类型从映射中获取对应的退款函数

	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("操作成功", c)
	return
}

type 结构请求_退款 struct {
	PayOrder string `json:"PayOrder"`
	IsOutRMB bool   `json:"IsOutRMB"`
	Note     string `json:"Note"`
}

type 结构请求_批量修改备注 struct {
	PayOrder []string `json:"PayOrder"` //用户id数组
	Note     string   `json:"Note"`     //
}

// 批量修改备注
func (a *Api) Set修改备注(c *gin.Context) {
	var 请求 结构请求_批量修改备注
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if len(请求.PayOrder) == 0 {
		response.FailWithMessage("订单数组为空", c)
		return
	}

	err = Ser_RMBPayOrder.Order更新订单备注_批量(请求.PayOrder, 请求.Note)

	if err != nil {
		response.FailWithMessage("修改失败", c)
		global.GVA_LOG.Error("修改失败:" + err.Error())
		return
	}

	response.OkWithMessage("修改成功", c)
	return
}
