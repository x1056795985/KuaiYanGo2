package UserClass

import (
	"github.com/gin-gonic/gin"
	UserClass服务 "server/Service/Ser_UserClass"
	"server/global"
	"server/structs/Http/response"
	DB "server/structs/db"
	"strconv"
)

type Api struct{}

// GetUserClassInfo
func (a *Api) GetUserClassInfo(c *gin.Context) {
	var 请求 结构请求_单id
	//{"Id":2}
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	var DB_UserClass DB_UserClass2

	err = global.GVA_DB.Model(DB.DB_UserClass{}).Where("Id = ?", 请求.Id).First(&DB_UserClass).Error
	// 没查到数据

	if err != nil {
		response.FailWithMessage("查询用户类型信息失败.id可能不存在", c)
		return
	}

	response.OkWithDetailed(DB_UserClass, "获取成功", c)
	return
}

type DB_UserClass2 struct {
	DB.DB_UserClass
}

type 结构请求_单id struct {
	Id    int `json:"Id"`
	AppId int `json:"AppId"`
}

type 结构请求_GetUserClassList struct {
	AppId    int    `json:"AppId"`    // Appid 必填
	Page     int    `json:"Page"`     // 页
	Size     int    `json:"Size"`     // 页数量
	Type     int    `json:"Type"`     // 关键字类型  1 id 2 用户名 3绑定信息 4 动态标签
	Keywords string `json:"Keywords"` // 关键字
	Order    int    `json:"Order"`    // 0 倒序 1 正序
}

// GetUserClassList
// 获取用户信息列表
func (a *Api) GetUserClassList(c *gin.Context) {
	var 请求 结构请求_GetUserClassList
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

	var DB_UserClass []DB.DB_UserClass
	var 总数 int64

	局_DB := global.GVA_DB.Model(DB.DB_UserClass{}).Where("AppId=?", 请求.AppId)
	if 请求.Order == 1 {
		局_DB.Order("Id ASC")
	} else {
		局_DB.Order("Id DESC")
	}

	//是否vip状态可用  //1=账号限时,2=账号计点,3卡号限时,4=卡号计点
	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //id
			局_DB.Where("Id = ?", 请求.Keywords)
		}

	}

	//Count(&总数) 必须放在where 后面 不然值会被清0
	err = 局_DB.Count(&总数).Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&DB_UserClass).Error
	//fmt.Println("用户总数%d", 总数, DB_LinksToken)
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		global.GVA_LOG.Error("GetUserClassList:" + err.Error())
		return
	}

	response.OkWithDetailed(结构响应_GetUserClassList{DB_UserClass, 总数}, "获取成功", c)
	return
}

type 结构响应_GetUserClassList struct {
	List  interface{} `json:"List"`  // 列表
	Count int64       `json:"Count"` // 总数
}

// Del批量删除用户类型
func (a *Api) Del批量删除用户类型(c *gin.Context) {
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
	if 请求.AppId < 10000 {
		response.FailWithMessage("AppId请输>=10000的整数", c)
		return
	}
	var 影响行数 int64
	var db = global.GVA_DB
	影响行数 = db.Model(DB.DB_UserClass{}).Where("AppId = ? ", 请求.AppId).Where("Id IN ? ", 请求.Id).Delete("").RowsAffected
	if db.Error != nil {
		response.FailWithMessage("删除失败", c)
		return
	}

	//修改软件中已经删除的用户类型为0
	db = global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(请求.AppId)).Select("UserClassId").Where("UserClassId IN ?", 请求.Id).Update("UserClassId", 0)
	err = db.Error
	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10)+",软件用户修改为未分类数量:"+strconv.FormatInt(db.RowsAffected, 10), c)
	return
}

type 结构请求_ID数组 struct {
	Id    []int `json:"Id"`    //用户id数组
	AppId int   `json:"AppId"` // Appid 必填
}

// save 保存
func (a *Api) SaveUserClass信息(c *gin.Context) {
	var 请求 DB.DB_UserClass
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if 请求.Name == "" {
		response.FailWithMessage("用户名称不能为空", c)
		return
	}
	if 请求.Id <= 0 {
		response.FailWithMessage("Id错误", c)
		return
	}

	if 请求.Weight <= 0 {
		response.FailWithMessage("权重最小为1", c)
		return
	}
	if !UserClass服务.UserClassId是否存在(请求.Id) {
		response.FailWithMessage("用户类型不存在", c)
		return
	}

	if UserClass服务.IsMark存在数量(请求.AppId, 请求.Mark) >= 1 {
		response.FailWithMessage("整数代号已存在", c)
		return
	}

	var db = global.GVA_DB.Model(DB.DB_UserClass{}).Omit("Id", "Appid").Where("Id = ?", 请求.Id)
	err = db.Updates(请求).Error

	if err != nil {
		response.FailWithMessage("保存失败", c)
		return
	}
	response.OkWithMessage("保存成功", c)
	return
}

// NewUserClass信息
func (a *Api) NewUserClass信息(c *gin.Context) {
	var 请求 DB.DB_UserClass
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if 请求.AppId < 10000 {
		response.FailWithMessage("AppId错误", c)
		return
	}
	if 请求.Id > 0 {
		response.FailWithMessage("添加用户不能有id值", c)
		return
	}
	if 请求.Weight <= 0 {
		response.FailWithMessage("权重最小为1", c)
		return
	}
	if UserClass服务.UserClassName是否存在(请求.AppId, 请求.Name) {
		response.FailWithMessage("用户类型名称已存在", c)
		return
	}

	if UserClass服务.UserClassMark是否存在(请求.AppId, 请求.Mark) {
		response.FailWithMessage("整数代号已存在", c)
		return
	}

	//app_id 没有这个字段排除掉
	err = global.GVA_DB.Model(DB.DB_UserClass{}).Create(&请求).Error
	if err != nil {
		response.FailWithMessage("添加失败", c)
		return
	}
	response.OkWithMessage("添加成功", c)
	return
}

// GetAppIdNameList 取id和名字数组
func (a *Api) GetIdNameList(c *gin.Context) {
	var 请求 结构请求_单id
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if 请求.AppId < 10000 {
		response.FailWithMessage("AppId错误", c)
		return
	}
	IdName := UserClass服务.UserClass取map列表String(请求.AppId)

	var 临时Int int
	var Name []键值对

	for Key := range IdName {
		临时Int, _ = strconv.Atoi(Key)
		Name = append(Name, 键值对{Id: 临时Int, Name: IdName[Key]})
	}

	response.OkWithDetailed(响应_AppIdNameList{IdName, Name}, "获取成功", c)
	return
}

type 响应_AppIdNameList struct {
	Map   map[string]string `json:"Map"`
	Array []键值对             `json:"Array"`
}

type 键值对 struct {
	Id   int    `json:"id"`
	Name string `json:"Name"`
}
