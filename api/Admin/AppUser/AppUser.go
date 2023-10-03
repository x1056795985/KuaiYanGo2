package AppUser

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/Service/Ser_Admin"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_AppUser"
	"server/Service/Ser_Ka"
	"server/Service/Ser_LinkUser"
	"server/Service/Ser_Log"
	"server/Service/Ser_User"
	"server/Service/Ser_UserClass"
	"server/Service/Ser_UserConfig"
	"server/global"
	"server/structs/Http/response"
	DB "server/structs/db"
	"strconv"
	"time"
)

type Api struct{}

// GetAppUserInfo
func (a *Api) GetAppUserInfo(c *gin.Context) {
	var 请求 结构请求_单id
	//{"Id":2}
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}
	if 请求.AppId < 10000 {
		response.FailWithMessage("AppId请输>=10000的整数", c)
		return
	}

	var DB_AppUser DB_AppUser2

	err = global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(请求.AppId)).Omit("app_type").Where("id = ?", 请求.Id).Find(&DB_AppUser).Error
	// 没查到数据

	if err != nil {
		response.FailWithMessage("查询软件用户详细信息失败", c)
		return
	}

	app信息 := Ser_AppInfo.App取App详情(请求.AppId)
	DB_AppUser.AppType = app信息.AppType

	response.OkWithDetailed(DB_AppUser, "获取成功", c)
	return
}

type DB_AppUser2 struct {
	DB.DB_AppUser
	AppType int `json:"AppType"` //登录平台App名字
}

type 结构请求_单id struct {
	Id    int `json:"Id"`
	AppId int `json:"AppId"`
}

type 结构请求_GetAppUserList struct {
	AppId    int    `json:"AppId"`    // Appid 必填
	Page     int    `json:"Page"`     // 页
	Size     int    `json:"Size"`     // 页数量
	Status   int    `json:"Status"`   // 状态id
	Type     int    `json:"Type"`     // 关键字类型  1 id 2 用户名 3绑定信息 4 动态标签
	Keywords string `json:"Keywords"` // 关键字
	Order    int    `json:"Order"`    // 0 倒序 1 正序
}

// GetAppUserList
// 获取用户信息列表
func (a *Api) GetAppUserList(c *gin.Context) {
	var 请求 结构请求_GetAppUserList
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

	var DB_AppUser []DB_AppUser带User信息
	var 总数 int64
	var 表名_AppUser = "db_AppUser_" + strconv.Itoa(请求.AppId)
	局_DB := global.GVA_DB.Table(表名_AppUser).Debug()
	if Ser_AppInfo.App是否为卡号(请求.AppId) {
		局_DB = 局_DB.Select(表名_AppUser+".*", "db_Ka.Name").Joins("left join db_Ka on " + 表名_AppUser + ".Uid=db_Ka.Id")
	} else {
		//mark 现在只是 链接 user表,后期需要处理 链接卡号表 读取用户名
		局_DB = 局_DB.Select(表名_AppUser+".*", "db_User.User").Joins("left join db_User on " + 表名_AppUser + ".Uid=db_User.Id")
	}
	if 请求.Order == 1 {
		局_DB.Order(表名_AppUser + ".Id ASC")
	} else {
		局_DB.Order(表名_AppUser + ".Id DESC")
	}
	var app信息 DB.DB_AppInfo
	app信息 = Ser_AppInfo.App取App详情(请求.AppId)
	//是否vip状态可用  //1=账号限时,2=账号计点,3卡号限时,4=卡号计点
	if 请求.Status == 1 {
		if app信息.AppType == 2 || app信息.AppType == 4 {
			局_DB.Where(表名_AppUser+".VipTime > ?", 0)
		} else {
			局_DB.Where(表名_AppUser+".VipTime > ?", time.Now().Unix())
		}

	} else if 请求.Status == 2 {
		if app信息.AppType == 2 || app信息.AppType == 4 {
			局_DB.Where(表名_AppUser+".VipTime <= ?", 0)
		} else {
			局_DB.Where(表名_AppUser+".VipTime <= ?", time.Now().Unix())
		}
	}

	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //id
			局_DB.Where(表名_AppUser+".Id = ?", 请求.Keywords)
		case 2: //用户id
			局_DB.Where(表名_AppUser+".Uid = ?", 请求.Keywords)
		case 3: //用户id
			if Ser_AppInfo.App是否为卡号(请求.AppId) {
				局_DB.Where(表名_AppUser+".Uid In ?", gorm.Expr("(Select Id from db_Ka where LOCATE(?, db_Ka.Name)>0 )", 请求.Keywords))
			} else {
				局_DB.Where(表名_AppUser+".Uid In ?", gorm.Expr("(Select Id from db_User where LOCATE(?, db_User.User)>0 )", 请求.Keywords))
			}
		case 4: //绑定信息
			局_DB.Where("LOCATE(?, "+表名_AppUser+".Key)>0 ", 请求.Keywords)
		case 5: //软件用户备注
			局_DB.Where("LOCATE( ?, "+表名_AppUser+".Note)>0 ", 请求.Keywords)
		}

	}

	//Count(&总数) 必须放在where 后面 不然值会被清0
	err = 局_DB.Count(&总数).Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Scan(&DB_AppUser).Error
	//fmt.Println("用户总数%d", 总数, DB_LinksToken)
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		global.GVA_LOG.Error("GetAppUserList:" + err.Error())
		return
	}
	UserClass := Ser_UserClass.UserClass取map列表Int(请求.AppId)
	response.OkWithDetailed(结构响应_GetAppUserList{DB_AppUser, 总数, app信息.AppType, UserClass}, "获取成功", c)
	return
}

type 结构响应_GetAppUserList struct {
	List      interface{}    `json:"List"`      // 列表
	Count     int64          `json:"Count"`     // 总数
	AppType   int            `json:"AppType"`   // //1=账号限时,2=账号计点,3卡号限时,4=卡号计点
	UserClass map[int]string `json:"UserClass"` //
}

type DB_AppUser带User信息 struct {
	DB.DB_AppUser
	User   string `json:"User" gorm:"column:User;index;comment:用户登录名"`                 // 用户登录名
	Name   string `json:"Name" gorm:"column:Name;index;comment:卡号"`                    // 用户登录名
	Status int    `json:"Status" gorm:"column:Status;default:1;comment:用户是状态 1正常 2冻结"` // 1正常 2冻结
}

// Del批量删除软件用户
func (a *Api) Del批量删除软件用户(c *gin.Context) {
	var 请求 结构请求_ID数组
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}

	if 请求.AppId < 10000 {
		response.FailWithMessage("AppId请输>=10000的整数", c)
		return
	}

	if len(请求.Id) == 0 {
		response.FailWithMessage("Id数组为空", c)
		return
	}
	var 影响行数 int64
	var 软件用户Uid = Ser_AppUser.Id取Uid_批量(请求.AppId, 请求.Id)

	var db = global.GVA_DB
	影响行数 = db.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(请求.AppId)).Where("Id IN ? ", 请求.Id).Delete("").RowsAffected
	if db.Error != nil {
		response.FailWithMessage("删除失败", c)
		return
	}
	_ = db.Model(DB.DB_UserConfig{}).Where("AppId = ? ", 请求.AppId).Where("Uid IN ? ", 软件用户Uid).Delete("").RowsAffected

	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
	return
}

type 结构请求_ID数组 struct {
	Id    []int `json:"Id"` //用户id数组
	AppId int   `json:"AppId"`
}

type 结构请求_DB_AppUser_UserConfig struct {
	AppId      int                `json:"AppId"` // Appid 必填
	AppUser    DB.DB_AppUser      `json:"AppUser"`
	UserConfig []DB.DB_UserConfig `json:"UserConfig"`
}

// save 保存
func (a *Api) Save用户信息(c *gin.Context) {
	var 请求 结构请求_DB_AppUser_UserConfig
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if 请求.AppId < 10000 {
		response.FailWithMessage("AppId请输>=10000的整数", c)
		return
	}

	if 请求.AppUser.Id <= 0 {
		response.FailWithMessage("Id错误", c)
		return
	}

	var count int64
	err = global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(请求.AppId)).Where("Id = ?", 请求.AppUser.Id).Count(&count).Error
	// 没查到数据
	if count == 0 {
		response.FailWithMessage("用户不存在", c)
		return
	}

	//直接排除Uid 禁止修改  Select可能0值 或"" 的字段防止不更新
	var db = global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(请求.AppId)).Where("Id = ?", 请求.AppUser.Id)
	err = db.Updates(map[string]interface{}{
		"Key":         请求.AppUser.Key,
		"VipTime":     请求.AppUser.VipTime,
		"VipNumber":   请求.AppUser.VipNumber,
		"Note":        请求.AppUser.Note,
		"MaxOnline":   请求.AppUser.MaxOnline,
		"UserClassId": 请求.AppUser.UserClassId,
	}).Error

	if err != nil {
		response.FailWithMessage("保存失败", c)
		return
	}
	response.OkWithMessage("保存成功", c)
	db = global.GVA_DB.Model(DB.DB_UserConfig{})
	for _, 值 := range 请求.UserConfig {
		_ = Ser_UserConfig.Z置值(请求.AppId, 请求.AppUser.Uid, 值.Name, 值.Value)
	}
	return
}

type 结构请求_DB_AppUser struct {
	AppId int `json:"AppId"` // Appid 必填
	DB.DB_AppUser
}

// New用户信息
func (a *Api) New用户信息(c *gin.Context) {
	var 请求 结构请求_DB_AppUser
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if 请求.AppId <= 0 {
		response.FailWithMessage("AppId错误", c)
		return
	}

	if 请求.Id > 0 {
		response.FailWithMessage("添加用户不能有id值", c)
		return
	}

	if Ser_AppInfo.App是否为卡号(请求.AppId) {
		if !Ser_Ka.KaId是否存在(请求.AppId, 请求.Uid) {
			response.FailWithMessage(`卡号Uid不存在,
请先去[ 卡号列表 => 制新卡 ],
添加信息`, c)
			return
		}
	} else {
		if !Ser_User.UserId是否存在(请求.Uid) {
			response.FailWithMessage(`用户Uid不存在,
请先去[ 用户管理 => 用户账户 ],
添加该用户信息`, c)
			return
		}
	}

	var count int64
	err = global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(请求.AppId)).Where("Uid  = ?", 请求.Uid).Count(&count).Error
	// 没查到数据
	if count != 0 {
		response.FailWithMessage("用户已存在", c)
		return
	}
	请求.RegisterTime = int(time.Now().Unix())
	//app_id 没有这个字段排除掉

	局_信息 := DB.DB_AppUser{
		Uid:          请求.Uid,
		Status:       请求.Status,
		Key:          请求.Key,
		VipTime:      请求.VipTime,
		VipNumber:    请求.VipNumber,
		Note:         请求.Note,
		MaxOnline:    请求.MaxOnline,
		UserClassId:  请求.UserClassId,
		RegisterTime: 请求.RegisterTime,
	}
	err = global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_" + strconv.Itoa(请求.AppId)).Create(&局_信息).Error
	if err != nil {
		response.FailWithMessage("添加失败", c)
		return
	}
	response.OkWithMessage("添加成功", c)
	return
}

// 批量修改状态
func (a *Api) Set修改状态(c *gin.Context) {
	var 请求 结构请求_批量修改状态
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if 请求.AppId <= 0 {
		response.FailWithMessage("AppId错误", c)
		return
	}
	if len(请求.Id) == 0 {
		response.FailWithMessage("Id数组为空", c)
		return
	}

	if 请求.Status != 1 && 请求.Status != 2 {
		response.FailWithMessage("修改失败:Status状态代码错误", c)
		return
	}
	var 表名_AppUser = "db_AppUser_" + strconv.Itoa(请求.AppId)

	err = global.GVA_DB.Table(表名_AppUser).Where("Id IN ? ", 请求.Id).Update("Status", 请求.Status).Error

	if err != nil {
		response.FailWithMessage("修改失败", c)
		global.GVA_LOG.Error("修改失败:" + err.Error())
		return
	}

	if 请求.Status == 2 {
		局_uid数组 := make([]int, 0, len(请求.Id))
		for _, 值 := range 请求.Id {
			局_uid数组 = append(局_uid数组, Ser_AppUser.Id取Uid(请求.AppId, 值))
		}
		_ = Ser_LinkUser.Set批量注销Uid数组(局_uid数组, 请求.AppId)
	}

	response.OkWithMessage("修改成功", c)
	return
}

type 结构请求_批量修改状态 struct {
	Id     []int `json:"Id"`     //用户id数组
	AppId  int   `json:"AppId"`  //用户id数组
	Status int   `json:"Status"` //1 解冻 2冻结
}

// 批量维护 增减时间点数
func (a *Api) Set批量维护_增减时间点数(c *gin.Context) {
	var 请求 结构请求_批量修改状态
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if 请求.AppId <= 0 {
		response.FailWithMessage("AppId错误", c)
		return
	}
	if len(请求.Id) == 0 {
		response.FailWithMessage("Id数组为空", c)
		return
	}
	if 请求.Status > 0 {
		err = Ser_AppUser.Id点数增减_批量(请求.AppId, 请求.Id, int64(请求.Status), true)
	} else {
		err = Ser_AppUser.Id点数增减_批量(请求.AppId, 请求.Id, int64(-请求.Status), false)
	}

	if err != nil {
		response.FailWithMessage("修改失败", c)
		global.GVA_LOG.Error("修改失败:" + err.Error())
		return
	}

	response.OkWithMessage("修改成功", c)

	if Ser_AppInfo.App是否为计点(请求.AppId) {
		for _, 局_id := range 请求.Id {
			Ser_Log.Log_写积分点数时间日志(Ser_AppUser.Id取User(请求.AppId, 局_id), c.ClientIP(), "管理员"+Ser_Admin.Id取User(c.GetInt("Uid"))+"批量增减点数", float64(请求.Status), 请求.AppId, 2)
		}
	}

	return
}
