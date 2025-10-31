package PublicData

import (
	"EFunc/utils"
	"github.com/gin-gonic/gin"
	Db服务 "server/Service/Ser_AppInfo"
	"server/global"
	"server/new/app/logic/common/log"
	"server/new/app/logic/common/publicData"
	"server/structs/Http/response"
	DB "server/structs/db"
	"strconv"
	"time"
)

type Api struct{}

// GetInfo
func (a *Api) GetInfo(c *gin.Context) {
	var 请求 DB.DB_PublicData
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	var DB_PublicData DB.DB_PublicData

	err = global.GVA_DB.Model(DB.DB_PublicData{}).Where(" AppId= ?", 请求.AppId).Where(" Name= ?", 请求.Name).First(&DB_PublicData).Error
	// 没查到数据
	if err != nil {
		response.FailWithMessage("获取公共变量失败,可能联合主键不存在", c)
		return
	}

	response.OkWithDetailed(DB_PublicData, "获取成功", c)
	return
}

type 结构请求_GetDB_PublicDataList struct {
	AppId          int    `json:"AppId"`    // Appid 必填
	Page           int    `json:"Page"`     // 页
	Size           int    `json:"Size"`     // 页数量
	Type           int    `json:"Type"`     // 关键字类型  1 id 2 用户名 3绑定信息 4 动态标签
	Keywords       string `json:"Keywords"` // 关键字
	Order          int    `json:"Order"`    // 0 倒序 1 正序
	PublicDataType []int  `json:"PublicDataType"`
}

// GetDB_PublicDataList

func (a *Api) GetPublicDataList(c *gin.Context) {
	var 请求 结构请求_GetDB_PublicDataList
	//{"AppId":2,"Type":2,"Size":10,"Page":1,"Status":1,"keywords":"1"}
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	局_DB := global.GVA_DB.Model(DB.DB_PublicData{})

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
		}

	}
	if len(请求.PublicDataType) > 0 {
		switch 请求.Type {
		case 1: //变量名
			局_DB.Where("Type IN ? ", 请求.PublicDataType)
		}

	}

	var DB_PublicData []结构响应_DB_PublicData扩展
	var 总数 int64
	//Count(&总数) 必须放在where 后面 不然值会被清0
	err = 局_DB.Count(&总数).Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Omit("AppName").Find(&DB_PublicData).Error
	//fmt.Println("用户总数%d", 总数, DB_LinksToken)
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		global.GVA_LOG.Error("GetDB_PublicDataList:" + err.Error())
		return
	}

	var AppName = Db服务.App取map列表String()
	AppName["1"] = "全局"

	for 索引 := range DB_PublicData {
		//fmt.Printf("Id:%v:%v", strconv.Itoa(DB_PublicData[索引].AppId), AppName[strconv.Itoa(DB_PublicData[索引].AppId)])
		DB_PublicData[索引].AppName = AppName[strconv.Itoa(DB_PublicData[索引].AppId)]
		if DB_PublicData[索引].Type == 4 && utils.W文本_取长度(DB_PublicData[索引].Value) > 200 { // 4 是队列
			DB_PublicData[索引].Value = utils.W文本_取左边(DB_PublicData[索引].Value, 200) + "..."
		} else {
			DB_PublicData[索引].Value = DB_PublicData[索引].Value
		}
	}

	response.OkWithDetailed(结构响应_GetDB_PublicDataList{DB_PublicData, 总数}, "获取成功", c)
	return
}

type 结构响应_DB_PublicData扩展 struct {
	DB.DB_PublicData
	AppName string `json:"AppName"`
}

type 结构响应_GetDB_PublicDataList struct {
	List  interface{} `json:"List"`  // 列表
	Count int64       `json:"Count"` // 总数
}

type 结构请求_批量Delete struct {
	Data []DB.DB_PublicData `json:"data"` //用户id数组
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
	影响行数 = db.Model(DB.DB_PublicData{}).Delete(请求.Data).RowsAffected
	if db.Error != nil {
		response.FailWithMessage("删除失败", c)
		return
	}

	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
	return
}

// save 保存
func (a *Api) SaveDB_PublicData信息(c *gin.Context) {

	var 请求 DB.DB_PublicData
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

	if !publicData.L_publicData.Name是否存在(&gin.Context{}, 请求.AppId, 请求.Name) {
		response.FailWithMessage("变量不存在", c)
		return
	}
	请求.Time = time.Now().Unix()
	err = publicData.L_publicData.Z置值_原值(c, 请求)

	if err != nil {
		response.FailWithMessage("保存失败"+err.Error(), c)
		return
	}
	response.OkWithMessage("保存成功", c)
	return
}

// NewDB_PublicData信息
func (a *Api) New(c *gin.Context) {
	var 请求 DB.DB_PublicData
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

	if 请求.Type < 1 {
		response.FailWithMessage("变量类型错误", c)
		return
	}

	if publicData.L_publicData.Name是否存在(&gin.Context{}, 请求.AppId, 请求.Name) {
		response.FailWithMessage("变量名已存在", c)
		return
	}
	请求.Time = time.Now().Unix()
	//app_id 没有这个字段排除掉
	err = publicData.L_publicData.C创建(c, 请求)
	if err != nil {
		response.FailWithMessage("添加失败", c)
		return
	}
	response.OkWithMessage("添加成功", c)
	return
}

// Del批量修改
func (a *Api) Set修改vip限制(c *gin.Context) {
	var 请求 结构请求_批量修改vip限制
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if 请求.IsVip < 0 {
		response.FailWithMessage("IsVip值错误", c)
		return
	}
	if len(请求.Name) == 0 {
		response.FailWithMessage("公共变量数组为空", c)
		return
	}

	err = publicData.L_publicData.P批量修改IsVip(c, 请求.AppID, 请求.Name, 请求.IsVip)

	if err != nil {
		response.FailWithMessage("修改失败", c)
		global.GVA_LOG.Error("修改失败:" + err.Error())
		return
	}

	response.OkWithMessage("修改成功", c)
	return
}

type 结构请求_批量修改vip限制 struct {
	Name  []string `json:"Name"`  //用户id数组
	IsVip int      `json:"IsVip"` //1 解冻 2冻结
	AppID int      `json:"AppID"` //1 解冻 2冻结
}
