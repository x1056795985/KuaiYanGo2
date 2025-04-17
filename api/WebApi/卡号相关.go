package WebApi

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_Ka"
	"server/Service/Ser_KaClass"
	"server/Service/Ser_LinkUser"
	"server/Service/Ser_Log"
	"server/global"
	"server/structs/Http/response"
	DB "server/structs/db"
)

type 结构请求_单卡号 struct {
	Name string `json:"Name"`
}

// GetKaInfo 获取卡的详细信息

func Get卡号详细信息(c *gin.Context) {
	var 请求 结构请求_单卡号
	//{"Name":"13212315153"}
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	var DB_Ka DB.DB_Ka

	err = global.GVA_DB.Model(DB.DB_Ka{}).Where("Name = ?", 请求.Name).First(&DB_Ka).Error
	// 没查到数据
	if err != nil {
		response.FailWithMessage("查询详细信息失败", c)
		return
	}

	response.OkWithDetailed(DB_Ka, "获取成功", c)
	return
}

type 结构请求_New struct {
	Id        int      `json:"Id"`        //卡类id
	Number    int      `json:"Number"`    //生成数量
	AdminNote string   `json:"AdminNote"` //管理员备注
	KaName    []string `json:"KaName"`    //指定卡号, 如果指定,则生成数量无效
}

// New  制新卡
func New制新卡(c *gin.Context) {
	var 请求 结构请求_New
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}

	if !Ser_KaClass.KaClassId是否存在(请求.Id) {
		response.FailWithMessage("卡类id不存在", c)
		return
	}
	if 请求.Number <= 0 {
		response.FailWithMessage("生成数量必须大于0", c)
		return
	}
	if 请求.Number > 5000 {
		response.FailWithMessage("生成数量每批最大5000", c)
		return
	}

	数组_卡 := make([]DB.DB_Ka, 请求.Number) //make初始化,有3个元素的切片, len和cap都为3

	用户名 := Ser_LinkUser.Token取Name(c.Request.Header.Get("Token"))
	err = Ser_Ka.Ka批量创建(数组_卡[:], 请求.Id, 用户名, 请求.AdminNote, "", 0)

	if err != nil {
		response.FailWithMessage("制卡失败:"+err.Error(), c)
		return
	}

	数组_卡_精简 := make([]DB_Ka_精简, 请求.Number) //make初始化,有3个元素的切片, len和cap都为3
	数组_卡号 := make([]string, 请求.Number)     //make初始化,有3个元素的切片, len和cap都为3
	for 索引 := range 数组_卡_精简 {
		数组_卡号[索引] = 数组_卡[索引].Name
		数组_卡_精简[索引].Name = 数组_卡[索引].Name
		数组_卡_精简[索引].Id = 数组_卡[索引].Id
		数组_卡_精简[索引].RMb = 数组_卡[索引].RMb
		数组_卡_精简[索引].VipTime = 数组_卡[索引].VipTime
		数组_卡_精简[索引].VipNumber = 数组_卡[索引].VipNumber
	}

	response.OkWithDetailed(数组_卡_精简, "制卡成功", c)

	局_文本 := fmt.Sprintf("新制卡号应用:%s,卡类:%s,批次id:{{批次id}}({{卡号索引}}/%d)", Ser_AppInfo.App取AppName(数组_卡[0].AppId), Ser_KaClass.Id取Name(数组_卡[0].KaClassId), 请求.Number)
	go Ser_Log.Log_写卡号操作日志(用户名, c.ClientIP(), 局_文本, 数组_卡号, 1, 4)

	return
}

type DB_Ka_精简 struct {
	Id        int     `json:"Id" gorm:"column:Id;primarykey"`
	Name      string  `json:"Name" gorm:"column:Name;comment:卡号"`
	VipTime   int64   `json:"VipTime" gorm:"column:VipTime;comment:增减时间秒数或点数"`
	RMb       float64 `json:"RMb" gorm:"column:RMb;type:decimal(10,2);default:0;comment:余额增减"`
	VipNumber float64 `json:"VipNumber" gorm:"column:VipNumber;type:decimal(10,2);default:0;comment:积分增减"`
}
