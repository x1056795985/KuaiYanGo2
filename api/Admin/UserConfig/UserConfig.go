package UserConfig

import (
	"github.com/gin-gonic/gin"
	"server/Service/Ser_Admin"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_AppUser"
	"server/Service/Ser_UserConfig"
	"server/global"
	"server/structs/Http/response"
	DB "server/structs/db"
	"strconv"
	"time"
)

type Api struct{}

// GetInfo
func (a *Api) GetInfo(c *gin.Context) {
	var 请求 DB.DB_UserConfig
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	var DB_UserConfig DB.DB_UserConfig
	Ser_UserConfig.Q取值(请求.AppId, 请求.Uid, 请求.Name)
	err = global.GVA_DB.Model(DB.DB_UserConfig{}).Where(" AppId= ?", 请求.AppId).Where(" Name= ?", 请求.Name).First(&DB_UserConfig).Error
	// 没查到数据
	if err != nil {
		response.FailWithMessage("获取公共变量失败,可能联合主键不存在", c)
		return
	}

	response.OkWithDetailed(DB_UserConfig, "获取成功", c)
	return
}

type 结构请求_GetDB_UserConfigList struct {
	AppId          int    `json:"AppId"`    // Appid 必填
	Page           int    `json:"Page"`     // 页
	Size           int    `json:"Size"`     // 页数量
	Type           int    `json:"Type"`     // 关键字类型  1 Uid 2 用户名 3绑定信息 4 动态标签
	Keywords       string `json:"Keywords"` // 关键字
	Order          int    `json:"Order"`    // 0 倒序 1 正序
	PublicDataType []int  `json:"PublicDataType"`
}

// GetDB_PublicDataList

func (a *Api) GetList(c *gin.Context) {
	var 请求 结构请求_GetDB_UserConfigList
	//{"AppId":2,"Type":2,"Size":10,"Page":1,"Status":1,"keywords":"1"}
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	局_DB := global.GVA_DB.Model(DB.DB_UserConfig{})

	if 请求.AppId > 0 {
		局_DB.Where("AppId=?", 请求.AppId)
	}

	if 请求.Order == 1 {
		局_DB.Order("Time ASC")
	} else if 请求.Order == 2 {
		局_DB.Order("Time DESC")
	}

	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //变量名
			局_DB.Where("LOCATE( ?, Name)>0 ", 请求.Keywords)
		case 2: //用户
			局_DB.Where("LOCATE( ?, User)>0 ", 请求.Keywords)
		case 3: //Uid
			局_DB.Where("Uid = ?", 请求.Keywords)
		}
	}

	var DB_PublicData []结构响应_DB_UserConfig扩展
	var 总数 int64
	//Count(&总数) 必须放在where 后面 不然值会被清0
	err = 局_DB.Count(&总数).Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Omit("AppName").Find(&DB_PublicData).Error
	//fmt.Println("用户总数%d", 总数, DB_LinksToken)
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		global.GVA_LOG.Error("GetDB_PublicDataList:" + err.Error())
		return
	}

	var AppName = Ser_AppInfo.App取map列表String()
	AppName["50"] = "代理云配置"

	var AdminIdNameMap = make(map[int]string) //cache 优化防止多次读库
	for 索引 := range DB_PublicData {
		//fmt.Printf("Id:%v:%v", strconv.Itoa(DB_PublicData[索引].AppId), AppName[strconv.Itoa(DB_PublicData[索引].AppId)])
		DB_PublicData[索引].AppName = AppName[strconv.Itoa(DB_PublicData[索引].AppId)]
		if DB_PublicData[索引].AppId == 1 { //管理平台单独处理
			if AdminIdNameMap[DB_PublicData[索引].Uid] == "" {
				AdminIdNameMap[DB_PublicData[索引].Uid] = Ser_Admin.Id取User(DB_PublicData[索引].Uid)
			}

			DB_PublicData[索引].User = AdminIdNameMap[DB_PublicData[索引].Uid]
			DB_PublicData[索引].Uid = -DB_PublicData[索引].Uid

		}

	}

	response.OkWithDetailed(结构响应_GetDB_PublicDataList{DB_PublicData, 总数}, "获取成功", c)
	return
}

type 结构响应_DB_UserConfig扩展 struct {
	DB.DB_UserConfig
	AppName string `json:"AppName"`
}

type 结构响应_GetDB_PublicDataList struct {
	List  interface{} `json:"List"`  // 列表
	Count int64       `json:"Count"` // 总数
}

type 结构请求_批量Delete struct {
	Data []DB.DB_UserConfig `json:"data"` //用户id数组
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

	if len(请求.Data) == 0 {
		response.FailWithMessage("数组为空", c)
		return
	}

	var 影响行数 int64
	var db = global.GVA_DB

	//AppId,Uid,User 联合主键 根据主键自动删除
	影响行数 = db.Model(DB.DB_UserConfig{}).Delete(请求.Data).RowsAffected
	if db.Error != nil {
		response.FailWithMessage("删除失败", c)
		return
	}

	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
	return
}

// NewDB_PublicData信息
func (a *Api) New(c *gin.Context) {
	var 请求 DB.DB_UserConfig
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if 请求.Name == "" {
		response.FailWithMessage("变量名不能为空", c)
		return
	}
	if 请求.AppId <= 0 {
		response.FailWithMessage("AppId错误", c)
		return
	}
	if !Ser_AppUser.Uid是否存在(请求.AppId, 请求.Uid) {
		response.FailWithMessage("软件用户不存在", c)
		return
	}

	if Ser_UserConfig.Name是否存在(请求.AppId, 请求.Uid, 请求.Name) {
		response.FailWithMessage("变量名已存在", c)
		return
	}

	请求.Time = time.Now().Unix()
	请求.UpdateTime = time.Now().Unix()
	请求.User = Ser_AppUser.Uid取User(请求.AppId, 请求.Uid)

	err = Ser_UserConfig.C创建(请求)
	if err != nil {
		response.FailWithMessage("添加失败", c)
		return
	}
	response.OkWithMessage("添加成功", c)
	return
}

func (a *Api) SetUserConfig(c *gin.Context) {

	var 请求 DB.DB_UserConfig
	err := c.ShouldBindJSON(&请求)
	//解析失败

	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if 请求.Name == "" {
		response.FailWithMessage("变量名不能为空", c)
		return
	}

	if !Ser_UserConfig.Name是否存在(请求.AppId, 请求.Uid, 请求.Name) {
		response.FailWithMessage("配置不存在", c)
		return
	}

	err = Ser_UserConfig.Z置值(请求.AppId, 请求.Uid, 请求.Name, 请求.Value)

	if err != nil {
		response.FailWithMessage("保存失败", c)
		return
	}
	response.OkWithMessage("保存成功", c)
	return
}
