package KaClass

import (
	"github.com/gin-gonic/gin"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_UserClass"
	"server/global"
	"server/structs/Http/response"
	DB "server/structs/db"
	"server/utils"
	"strconv"
)

type Api struct{}

// GetKaClassInfo
func (a *Api) GetInfo(c *gin.Context) {
	var 请求 结构请求_单id
	//{"Id":2}
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	var DB_KaClass DB.DB_KaClass

	err = global.GVA_DB.Model(DB.DB_KaClass{}).Where("Id = ?", 请求.Id).First(&DB_KaClass).Error
	// 没查到数据
	if err != nil {
		response.FailWithMessage("查询详细信息失败", c)
		return
	}

	response.OkWithDetailed(DB_KaClass, "获取成功", c)
	return
}

type 结构请求_单id struct {
	Id int `json:"Id"`
}

type 结构请求_GetKaClassList struct {
	AppId    int    `json:"AppId"`    // Appid 必填
	Page     int    `json:"Page"`     // 页
	Size     int    `json:"Size"`     // 页数量
	Type     int    `json:"Type"`     // 关键字类型  1 id 2 用户名 3绑定信息 4 动态标签
	Keywords string `json:"Keywords"` // 关键字
	Order    int    `json:"Order"`    // 0 倒序 1 正序
}

// GetKaClassList
// 获取用户信息列表
func (a *Api) GetKaClassList(c *gin.Context) {
	var 请求 结构请求_GetKaClassList
	//{"AppId":2,"Type":2,"Size":10,"Page":1,"Status":1,"keywords":"1"}
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	if 请求.AppId < 10000 {
		response.FailWithMessage("AppId请输>=10000的整数", c)
		return
	}

	var DB_KaClass []DB.DB_KaClass
	var 总数 int64

	局_DB := global.GVA_DB.Model(DB.DB_KaClass{}).Where("AppId = ?", 请求.AppId)
	if 请求.Order == 1 {
		局_DB.Order("Id ASC")
	} else {
		局_DB.Order("Id DESC")
	}
	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //id
			局_DB.Where("Id = ?", 请求.Keywords)
		case 2: //卡类名称
			局_DB.Where("LOCATE(?, Name)>0 ", 请求.Keywords)

		}
	}

	//Count(&总数) 必须放在where 后面 不然值会被清0
	err = 局_DB.Count(&总数).Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&DB_KaClass).Error
	//fmt.Println("用户总数%d", 总数, DB_LinksToken)
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		global.GVA_LOG.Error("GetKaClassList:" + err.Error())
		return
	}
	var AppType int
	AppType = Ser_AppInfo.App取AppType(请求.AppId)
	UserClass := Ser_UserClass.UserClass取map列表Int(请求.AppId)
	response.OkWithDetailed(结构响应_GetKaClassList{DB_KaClass, 总数, UserClass, AppType}, "获取成功", c)
	return
}

type 结构响应_GetKaClassList struct {
	List      interface{}    `json:"List"`      // 列表
	Count     int64          `json:"Count"`     // 总数
	UserClass map[int]string `json:"UserClass"` //
	AppType   int            `json:"AppType"`   //
}

// Del批量删除
func (a *Api) Delete(c *gin.Context) {
	var 请求 结构请求_ID数组
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}

	if len(请求.Id) == 0 {
		response.FailWithMessage("Id数组为空", c)
		return
	}
	var 影响行数 int64
	var db = global.GVA_DB
	影响行数 = db.Model(DB.DB_KaClass{}).Where("Id IN ? ", 请求.Id).Delete("").RowsAffected
	if db.Error != nil {
		response.FailWithMessage("删除失败", c)
		return
	}

	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
	return
}

type 结构请求_ID数组 struct {
	Id    []int `json:"Id"` //用户id数组
	AppId int   `json:"AppId"`
}

// save 保存
func (a *Api) SaveInfo(c *gin.Context) {
	var 请求 DB.DB_KaClass
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}

	if 请求.Id <= 0 {
		response.FailWithMessage("Id错误", c)
		return
	}

	if len(请求.Note) > 400 {
		response.FailWithMessage(`备注过长,请减少备注长度`, c)
		return
	}
	if 请求.KaLength-len(请求.Prefix) < 10 {
		response.FailWithMessage(`制卡可随机字符长度小于10,请增加卡长度或减少前缀长度`, c)
		return
	}
	if 请求.KaLength+len(请求.Prefix) > 191 {
		response.FailWithMessage(`制卡可随机字符长度最大191,请减少卡长度或减少前缀长度`, c)
		return
	}

	if 请求.VipNumber < 0 || 请求.VipTime < 0 || 请求.InviteCount < 0 || 请求.Num < 0 {
		response.FailWithMessage(`值不能为小于0`, c)
		return
	}

	if 请求.Money < -1 || 请求.AgentMoney < -1 {
		response.FailWithMessage(`售价值不能为小于-1`, c)
		return
	}

	var count int64
	err = global.GVA_DB.Model(DB.DB_KaClass{}).Where("Id = ?", 请求.Id).Count(&count).Error
	// 没查到数据
	if count == 0 {
		response.FailWithMessage("卡类不存在", c)
		return
	}

	//直接排除Aid 禁止修改  Select可能0值 或"" 的字段防止不更新
	var db = global.GVA_DB.Model(DB.DB_KaClass{})

	var data = map[string]interface{}{
		"KaStringType": 请求.KaStringType,
		"Note":         请求.Note,
		"MaxOnline":    请求.MaxOnline,
		"Name":         请求.Name,
		"Prefix":       请求.Prefix,
		"VipTime":      请求.VipTime,
		"KaType":       请求.KaType,
		"Num":          请求.Num,
		"KaLength":     请求.KaLength,
		"NoUserClass":  请求.NoUserClass,
		"UserClassId":  请求.UserClassId,
		"AgentMoney":   请求.AgentMoney,
		"Money":        请求.Money,
		"VipNumber":    请求.VipNumber,
		"RMb":          请求.RMb,
		"InviteCount":  请求.InviteCount,
	}

	if Ser_AppInfo.App是否为卡号(请求.AppId) {
		data["Num"] = 1 //卡号类型卡只能用一次
	}

	err = db.Where("Id = ?", 请求.Id).Updates(&data).Error

	if err != nil {
		response.FailWithMessage("保存失败", c)
		return
	}
	response.OkWithMessage("保存成功", c)
	return
}

// New
func (a *Api) New(c *gin.Context) {
	var 请求 DB.DB_KaClass
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if 请求.Id > 0 {
		response.FailWithMessage("添加用户不能有id值", c)
		return
	}
	if 请求.AppId < 10000 || !Ser_AppInfo.AppId是否存在(请求.AppId) {
		response.FailWithMessage("AppId错误", c)
		return
	}
	if 请求.Name == "" {
		response.FailWithMessage("卡类名称不能为空", c)
		return
	}

	var msg string
	if 请求.Prefix != "" && !utils.Z正则_是否英数(请求.Prefix, &msg) {
		response.FailWithMessage("卡类前缀"+msg, c)
		return
	}

	if !Ser_AppInfo.AppId是否存在(请求.AppId) {
		response.FailWithMessage(`AppId不存在,
请先去[ 应用管理 => 应用列表 ],
添加该应用信息`, c)
		return
	}

	if 请求.KaLength-len(请求.Prefix) < 10 {
		response.FailWithMessage(`制卡可随机字符长度小于10,请增加卡长度或减少前缀长度`, c)
		return
	}

	if 请求.VipNumber < 0 || 请求.VipTime < 0 || 请求.InviteCount < 0 || 请求.Num < 0 {
		response.FailWithMessage(`值不能为小于0`, c)
		return
	}

	if 请求.Money < -1 || 请求.AgentMoney < -1 {
		response.FailWithMessage(`售价值不能为小于-1`, c)
		return
	}
	if Ser_AppInfo.App是否为卡号(请求.AppId) {
		请求.Num = 1 //卡号类型卡只能用一次
	}
	//app_id 没有这个字段排除掉
	err = global.GVA_DB.Model(DB.DB_KaClass{}).Create(&请求).Error
	if err != nil {
		response.FailWithMessage("添加失败", c)
		return
	}
	response.OkWithMessage("添加成功", c)
	return
}
