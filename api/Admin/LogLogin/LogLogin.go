package LogLogin

import (
	"github.com/gin-gonic/gin"
	App服务 "server/Service/Ser_AppInfo"
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

	var DB_LogLogin DB.DB_LogLogin

	err = global.GVA_DB.Model(DB.DB_LogLogin{}).Where("Id= ?", 请求.Id).First(&DB_LogLogin).Error
	// 没查到数据
	if err != nil {
		response.FailWithMessage("获取失败,可能不存在", c)
		return
	}
	response.OkWithDetailed(DB_LogLogin, "获取成功", c)
	return
}

type 结构请求_GetDB_LogLoginList struct {
	Page         int      `json:"Page"`         // 页
	Size         int      `json:"Size"`         // 页数量
	Type         int      `json:"Type"`         // 关键字类型  1 用户名 2消息关键字
	Keywords     string   `json:"Keywords"`     // 关键字
	Order        int      `json:"Order"`        // 0 倒序 1 正序
	RegisterTime []string `json:"RegisterTime"` // 制卡开始时间 制卡结束时间
	Appid        int      `json:"Appid"`        // 制卡开始时间 制卡结束时间
}

// GetDB_LogLoginList

func (a *Api) GetLogLoginList(c *gin.Context) {
	var 请求 结构请求_GetDB_LogLoginList
	//{"Type":2,"Size":10,"Page":1,"Status":1,"keywords":"1"}
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	局_DB := global.GVA_DB.Model(DB.DB_LogLogin{})
	if 请求.Order == 1 {
		局_DB.Order("Id ASC")
	} else {
		局_DB.Order("Id DESC")
	}
	if 请求.Appid > 0 {
		局_DB.Where("LoginType = ?", 请求.Appid)
	}

	if 请求.RegisterTime != nil && len(请求.RegisterTime) == 2 && 请求.RegisterTime[0] != "" && 请求.RegisterTime[1] != "" {
		制卡开始时间, _ := strconv.Atoi(请求.RegisterTime[0])
		制卡结束时间, _ := strconv.Atoi(请求.RegisterTime[1])
		局_DB.Where("Time > ?", 制卡开始时间).Where("Time < ?", 制卡结束时间+86400)
	}
	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //用户名
			局_DB.Where("User  = ? ", 请求.Keywords)
		case 2: //消息
			局_DB.Where("LOCATE( ?, Note)>0 ", 请求.Keywords)
		case 3: //ip
			局_DB.Where("Ip LIKE ? ", "%"+请求.Keywords+"%")
		}
	}

	var DB_LogLogin []DB.DB_LogLogin
	var 总数 int64
	//Count(&总数) 必须放在where 后面 不然值会被清0
	err = 局_DB.Count(&总数).Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&DB_LogLogin).Error
	//fmt.Println("用户总数%d", 总数, DB_LinksToken)
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		return
	}

	var AppName = App服务.AppInfo取map列表Int()
	var AppNameMap = make(map[string]string, len(AppName))
	//不直接传全部应用数据,只传有的数据
	for 索引 := range DB_LogLogin {
		_, ok := AppNameMap[strconv.Itoa(DB_LogLogin[索引].LoginType)]
		if !ok { //如果不会存在,就植入数据
			AppNameMap[strconv.Itoa(DB_LogLogin[索引].LoginType)] = AppName[DB_LogLogin[索引].LoginType]
		}
	}

	response.OkWithDetailed(结构响应_GetDB_LogLoginList{DB_LogLogin, 总数, AppNameMap}, "获取成功", c)
	return
}

type 结构响应_GetDB_LogLoginList struct {
	List    interface{}       `json:"List"`  // 列表
	Count   int64             `json:"Count"` // 总数
	AppName map[string]string `json:"AppName"`
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
	var db = global.GVA_DB.Model(DB.DB_LogLogin{})

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
		影响行数 = db.Where("Id IN ? ", 请求.Id).Delete(DB.DB_LogLogin{}).RowsAffected
	case 2:
		影响行数 = db.Where("User = ? ", 请求.Keywords).Delete(DB.DB_LogLogin{}).RowsAffected
	case 3: //清空
		影响行数 = db.Where("1=1").Delete(DB.DB_LogLogin{}).RowsAffected
	case 4: //删7天前
		影响行数 = db.Where("Time < ?", time.Now().Unix()-604800).Delete(DB.DB_LogLogin{}).RowsAffected
	case 5: //删除30天前
		影响行数 = db.Where("Time <  ?", time.Now().Unix()-2592000).Delete(DB.DB_LogLogin{}).RowsAffected
	case 6: //删除90天前
		影响行数 = db.Where("Time <  ?", time.Now().Unix()-7776000).Delete(DB.DB_LogLogin{}).RowsAffected
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
