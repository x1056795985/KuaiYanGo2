package LogRMBPayOrder

import (
	"github.com/gin-gonic/gin"
	"server/Service/Ser_Log"
	"server/Service/Ser_RMBPayOrder"
	"server/Service/Ser_User"
	"server/global"
	"server/structs/Http/response"
	DB "server/structs/db"
	"server/utils"
	"strconv"
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
	for 索引 := range DB_LogRMBPayOrder {
		DB_LogRMBPayOrder[索引].Extra = Ser_RMBPayOrder.C处理类型[DB_LogRMBPayOrder[索引].ProcessingType]
		if DB_LogRMBPayOrder[索引].Extra == "" {
			DB_LogRMBPayOrder[索引].Extra = "未知原因" + strconv.Itoa(DB_LogRMBPayOrder[索引].ProcessingType)
		}

	}

	response.OkWithDetailed(结构响应_GetDB_LogRMBPayOrderList{DB_LogRMBPayOrder, 总数}, "获取成功", c)
	return
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
		影响行数 = db.Where("Time <  ?", time.Now().Unix()-3600).Where("Status = 1").Delete(DB.DB_LogRMBPayOrder{}).RowsAffected
		if db.Error == nil {
			response.OkWithMessage("删除过期(1小时前)待支付成功,数量"+strconv.FormatInt(影响行数, 10), c)
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

	if _, err = Ser_User.Id余额增减(新订单.Uid, 新订单.Rmb, true); err != nil {
		response.FailWithMessage("订单创建成功充值用户失败", c)
		return
	}

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

	if global.GVA_CONFIG.Z在线支付.J禁止退款 {
		response.FailWithMessage("已禁止退款,请手动前往服务器目录,修改配置文件 禁止退款:true", c)
		return
	}
	局_订单信息, ok := Ser_RMBPayOrder.Order取订单详细(请求.PayOrder)
	if !ok {
		response.FailWithMessage("订单不存在", c)
		return
	}
	if 局_订单信息.Status != 3 {
		response.FailWithMessage("订单非充值成功状态", c)
		return
	}
	if 请求.IsOutRMB {
		/*		for i := 0; i < 100; i++ {
				go Ser_User.Id余额增减(局_订单信息.Uid, 局_订单信息.Rmb, false)
			}*/
		局_新余额, err2 := Ser_User.Id余额增减(局_订单信息.Uid, 局_订单信息.Rmb, false)
		if err2 != nil {
			response.FailWithMessage(err2.Error(), c)
			return
		} else {
			go Ser_Log.Log_写余额日志(Ser_User.Id取User(局_订单信息.Uid), c.ClientIP(), "管理员操作退款,余额充值订单:"+局_订单信息.PayOrder+",扣除用户已充值余额"+"|新余额≈"+utils.Float64到文本(局_新余额, 2), utils.Float64取负值(局_订单信息.Rmb))
		}
	}
	Ser_RMBPayOrder.Order更新订单状态(局_订单信息.PayOrder, Ser_RMBPayOrder.D订单状态_退款中)
	switch 局_订单信息.Type {
	case "支付宝PC":
		err = Ser_RMBPayOrder.Order_退款_支付宝PC(局_订单信息)
	case "支付宝H5":
		err = Ser_RMBPayOrder.Order_退款_支付宝H5(局_订单信息)
	case "支付宝当面付":
		err = Ser_RMBPayOrder.Order_退款_支付宝当面付(局_订单信息)
	case "微信支付":
		err = Ser_RMBPayOrder.Order_退款_微信支付(局_订单信息)
	default:
		response.FailWithMessage("暂不支持该支付类型退款", c)
		return
	}
	if err != nil {
		response.FailWithMessage("退款失败:"+err.Error(), c)
		Ser_RMBPayOrder.Order更新订单状态(局_订单信息.PayOrder, Ser_RMBPayOrder.D订单状态_退款失败)
		Ser_RMBPayOrder.Order更新订单备注(局_订单信息.PayOrder, 局_订单信息.Note+"退款失败:"+err.Error())
		if 请求.IsOutRMB { //退款失败,再把余额加回去
			新余额, _ := Ser_User.Id余额增减(局_订单信息.Uid, 局_订单信息.Rmb, true)
			Ser_Log.Log_写余额日志(Ser_User.Id取User(局_订单信息.Uid), c.ClientIP(), "管理员操作退款失败恢复余额"+"|新余额≈"+utils.Float64到文本(新余额, 2), 局_订单信息.Rmb)
		}
		return
	}

	Ser_RMBPayOrder.Order更新订单备注(局_订单信息.PayOrder, 请求.Note)
	Ser_RMBPayOrder.Order更新订单状态(局_订单信息.PayOrder, Ser_RMBPayOrder.D订单状态_退款成功)
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
