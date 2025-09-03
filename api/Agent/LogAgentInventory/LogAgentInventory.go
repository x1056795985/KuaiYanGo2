package LogAgentInventory

import (
	"github.com/gin-gonic/gin"
	"server/global"
	"server/structs/Http/response"
	DB "server/structs/db"
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

	var Db_Agent_库存日志 DB.Db_Agent_库存日志

	err = global.GVA_DB.Model(DB.Db_Agent_库存日志{}).Where("Id= ?", 请求.Id).First(&Db_Agent_库存日志).Error
	// 没查到数据
	if err != nil {
		response.FailWithMessage("获取失败,可能不存在", c)
		return
	}
	response.OkWithDetailed(Db_Agent_库存日志, "获取成功", c)
	return
}

type 结构请求_GetDb_Agent_库存日志List struct {
	Page         int      `json:"Page"`         // 页
	Size         int      `json:"Size"`         // 页数量
	Type         int      `json:"Type"`         // 关键字类型  1 用户名 2消息关键字
	Keywords     string   `json:"Keywords"`     // 关键字
	Order        int      `json:"Order"`        // 0 倒序 1 正序
	RegisterTime []string `json:"RegisterTime"` // 日志开始时间 日志结束时间
	Count        int64    `json:"Count"`        // 总数
}

// GetDb_Agent_库存日志List

func (a *Api) GetLogAgentInventory(c *gin.Context) {
	var 请求 结构请求_GetDb_Agent_库存日志List
	//{"Type":2,"Size":10,"Page":1,"Status":1,"keywords":"1"}
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	局_DB := global.GVA_DB.Model(DB.Db_Agent_库存日志{}).Where("(User1  = ? OR  User2  = ?)", c.GetString("User"), c.GetString("User"))
	if 请求.Order == 1 {
		局_DB.Order("Id ASC")
	} else {
		局_DB.Order("Id DESC")
	}
	if 请求.RegisterTime != nil && len(请求.RegisterTime) == 2 && 请求.RegisterTime[0] != "" && 请求.RegisterTime[1] != "" {
		制卡开始时间, _ := strconv.ParseInt(请求.RegisterTime[0], 10, 64)
		制卡结束时间, _ := strconv.ParseInt(请求.RegisterTime[1], 10, 64)
		局_DB.Where("Time > ?", 制卡开始时间).Where("Time < ?", 制卡结束时间+86400)
	}
	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //用户名
			局_DB.Where("(User1  = ? OR  User2  = ?)", 请求.Keywords, 请求.Keywords)
		case 2: //消息
			局_DB.Where("LOCATE( ?, Note)>0 ", 请求.Keywords)
		case 3: //ip
			局_DB.Where("Ip = ? ", 请求.Keywords)
		case 4: //库存ID
			局_DB.Where("InventoryId  = ? ", 请求.Keywords)
		}
	}

	var Db_Agent_库存日志 []DB.Db_Agent_库存日志
	var 总数 int64
	//Count(&总数) 必须放在where 后面 不然值会被清0

	if 请求.Count > 500000 { //如果数据大于50万 直接使用,不重新查询了 优化速度
		总数 = 请求.Count
	} else {
		局_DB.Count(&总数)
	}
	err = 局_DB.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&Db_Agent_库存日志).Error
	//fmt.Println("用户总数%d", 总数, DB_LinksToken)
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		return
	}

	response.OkWithDetailed(结构响应_GetDb_Agent_库存日志List{Db_Agent_库存日志, 总数}, "获取成功", c)
	return
}

type 结构响应_GetDb_Agent_库存日志List struct {
	List  interface{} `json:"List"`  // 列表
	Count int64       `json:"Count"` // 总数
}

type 结构请求_批量Delete struct {
	Id       []int  `json:"Id"`       //用户id数组
	Type     int    `json:"Type"`     //  1删除用户数组 2删除指定关键字 3清空 4删除7天前 5删除30天前 6删除90天前
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
	var db = global.GVA_DB.Model(DB.Db_Agent_库存日志{})

	if 请求.Type <= 0 || 请求.Type > 7 {
		response.FailWithMessage("Type错误", c)
		return
	}

	//1删除用户数组 2删除指定用户 3清空 4删除7天前 5删除30天前 6删除90天前  7关键字
	switch 请求.Type {
	case 1:
		if 请求.Type == 1 && len(请求.Id) == 0 {
			response.FailWithMessage("Id数组没有要删除的ID", c)
			return
		}
		影响行数 = db.Where("Id IN ? ", 请求.Id).Delete(DB.Db_Agent_库存日志{}).RowsAffected
	case 2:
		影响行数 = db.Where("User1 = ? OR User2 = ?", 请求.Keywords, 请求.Keywords).Delete(DB.Db_Agent_库存日志{}).RowsAffected
	case 3: //清空
		影响行数 = db.Where("1=1").Delete(DB.Db_Agent_库存日志{}).RowsAffected
	case 4: //删7天前
		影响行数 = db.Where("Time <  ?", time.Now().Unix()-604800).Delete(DB.Db_Agent_库存日志{}).RowsAffected
	case 5: //删除30天前
		影响行数 = db.Where("Time <  ?", time.Now().Unix()-2592000).Delete(DB.Db_Agent_库存日志{}).RowsAffected
	case 6: //删除90天前
		影响行数 = db.Where("Time <  ?", time.Now().Unix()-7776000).Delete(DB.Db_Agent_库存日志{}).RowsAffected
	case 7: //删除关键字
		if len(请求.Keywords) == 0 {
			response.FailWithMessage("关键字不能为空", c)
			return
		}
		影响行数 = db.Where("LOCATE( ?, Note)>0 ", 请求.Keywords).Delete(请求.Id).RowsAffected
	}

	if db.Error != nil {
		response.FailWithMessage("删除失败", c)
		return
	}

	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
	return
}
