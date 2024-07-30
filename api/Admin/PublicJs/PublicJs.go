package PublicJs

import (
	. "EFunc/utils"
	"fmt"
	"github.com/dop251/goja"
	"github.com/gin-gonic/gin"
	Db服务 "server/Service/Ser_AppInfo"
	"server/Service/Ser_Js"
	"server/Service/Ser_PublicJs"
	"server/global"
	"server/structs/Http/response"
	DB "server/structs/db"
	"strconv"
	"time"
)

type Api struct{}

// GetInfo
func (a *Api) GetInfo(c *gin.Context) {
	var 请求 DB.DB_PublicJs
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	var DB_PublicJs DB.DB_PublicJs

	err = global.GVA_DB.Model(DB.DB_PublicJs{}).Where(" Name= ?", 请求.Name).First(&DB_PublicJs).Error
	// 没查到数据
	if err != nil {
		response.FailWithMessage("获取公共变量失败,可能联合主键不存在", c)
		return
	}

	if W文件_是否存在(global.GVA_CONFIG.Q取运行目录 + DB_PublicJs.Value) {
		DB_PublicJs.Value = string(W文件_读入文件(global.GVA_CONFIG.Q取运行目录 + DB_PublicJs.Value))
	} else {
		DB_PublicJs.Value = DB_PublicJs.Value + "[js文件读取失败可能被删除]"
	}

	response.OkWithDetailed(DB_PublicJs, "获取成功", c)
	return
}

type 结构请求_GetDB_PublicJsList struct {
	Page     int    `json:"Page"`     // 页
	Size     int    `json:"Size"`     // 页数量
	Type     int    `json:"Type"`     // 关键字类型  1 Id   2 函数名
	Keywords string `json:"Keywords"` // 关键字
	Order    int    `json:"Order"`    // 0 倒序 1 正序
}

// GetDB_PublicJsList

func (a *Api) GetPublicJsList(c *gin.Context) {
	var 请求 结构请求_GetDB_PublicJsList
	//{"AppId":2,"Type":2,"Size":10,"Page":1,"Status":1,"keywords":"1"}
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}
	//fmt.Println(global.GVA_DB)
	局_DB := global.GVA_DB.Model(DB.DB_PublicJs{})
	// global.GVA_DB.Model(DB.DB_PublicJs{}).Debug()
	//局_DB.Debug() //debug不能放到上面,会导致where不生效,反正是不行,采坑20240220
	if 请求.Order == 1 {
		局_DB.Order("Id ASC")
	} else if 请求.Order == 2 {
		局_DB.Order("Id DESC")
	}

	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //函数名
			局_DB.Where("Name LIKE ?", "%"+请求.Keywords+"%")
		}
	}

	var DB_PublicJs []结构响应_DB_PublicJs扩展
	var 总数 int64
	//Count(&总数) 必须放在where 后面 不然值会被清0
	err = 局_DB.Count(&总数).Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Omit("AppName").Find(&DB_PublicJs).Error
	//fmt.Println("用户总数%d", 总数, DB_LinksToken)
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		global.GVA_LOG.Error("GetDB_PublicJsList:" + err.Error())
		return
	}

	var AppName = Db服务.App取map列表String()
	AppName["1"] = "全局"
	AppName["2"] = "任务池Hook"
	AppName["3"] = "ApiHook"

	for 索引 := range DB_PublicJs {
		//fmt.Printf("Id:%v:%v", strconv.Itoa(DB_PublicJs[索引].AppId), AppName[strconv.Itoa(DB_PublicJs[索引].AppId)])
		DB_PublicJs[索引].AppName = AppName[strconv.Itoa(DB_PublicJs[索引].AppId)]
	}

	response.OkWithDetailed(结构响应_GetDB_PublicJsList{DB_PublicJs, 总数}, "获取成功", c)
	return
}

type 结构响应_DB_PublicJs扩展 struct {
	DB.DB_PublicJs
	AppName string `json:"AppName"`
}

type 结构响应_GetDB_PublicJsList struct {
	List  interface{} `json:"List"`  // 列表
	Count int64       `json:"Count"` // 总数
}

type 结构请求_批量Delete struct {
	Id       []int  `json:"Id"`       //id数组
	Type     int    `json:"Type"`     //  1删除ID数组 2删除指定关键字 3清空 4删除7天前 5删除30天前 6删除90天前
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

	if len(请求.Id) == 0 && 请求.Type == 1 {
		response.FailWithMessage("数组为空", c)
		return
	}

	for _, 值 := range 请求.Id {
		var DB_PublicJs DB.DB_PublicJs
		err = global.GVA_DB.Model(DB.DB_PublicJs{}).Where(" Id = ? ", 值).First(&DB_PublicJs).Error
		//同步删除云函数JS文件
		if W文件_是否存在(global.GVA_CONFIG.Q取运行目录 + DB_PublicJs.Value) {
			err = W文件_删除(global.GVA_CONFIG.Q取运行目录 + DB_PublicJs.Value)
			if err != nil {
				fmt.Printf("E删除文件失败%v", err.Error())
			}
		}
	}

	var 影响行数 int64
	var db = global.GVA_DB
	//1删除用户数组 2删除指定用户 3清空 4删除7天前 5删除30天前 6删除90天前  7 关键字
	switch 请求.Type {
	case 1:
		if 请求.Type == 1 && len(请求.Id) == 0 {
			response.FailWithMessage("Id数组没有要删除的ID", c)
			return
		}
		影响行数 = db.Where("Id IN ? ", 请求.Id).Delete(DB.DB_PublicJs{}).RowsAffected
	}
	if db.Error != nil {
		response.FailWithMessage("删除失败", c)
		return
	}

	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
	return
}

// save 保存
func (a *Api) SaveDB_PublicJs信息(c *gin.Context) {

	var 请求 DB.DB_PublicJs
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
	var 局_临时Id = Ser_PublicJs.Name取Id([]int{Ser_PublicJs.Js类型_公共函数, Ser_PublicJs.Js类型_任务池Hook函数}, 请求.Name) //1 全局,2hook函数
	if 局_临时Id != 0 && 局_临时Id != 请求.Id {
		response.FailWithMessage("变量名已存在", c)
		return
	}

	if !Ser_PublicJs.Id是否存在(请求.Id) {
		response.FailWithMessage("变量不存在", c)
		return
	}

	err = Ser_PublicJs.Z置值2(请求)

	if err != nil {
		response.FailWithMessage("保存失败", c)
		return
	}

	response.OkWithMessage("保存成功", c)
	return
}

// NewDB_PublicJs信息
func (a *Api) New(c *gin.Context) {
	var 请求 DB.DB_PublicJs
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

	var 局_临时Id = Ser_PublicJs.Name取Id([]int{Ser_PublicJs.Js类型_公共函数, Ser_PublicJs.Js类型_任务池Hook函数, Ser_PublicJs.Js类型_ApiHook函数}, 请求.Name) //1 全局,2hook函数
	if 局_临时Id != 0 && 局_临时Id != 请求.Id {
		response.FailWithMessage("变量名已存在", c)
		return
	}

	if W文本_是否包含关键字(请求.Name, "/") || W文本_是否包含关键字(请求.Name, ".") {
		response.FailWithMessage("函数名不能包含[/]或[].]符号", c)
		return
	}

	err = Ser_PublicJs.C创建(请求)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
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
	if len(请求.Id) == 0 {
		response.FailWithMessage("公共变量数组为空", c)
		return
	}

	err = Ser_PublicJs.P批量修改IsVip(请求.Id, 请求.IsVip)

	if err != nil {
		response.FailWithMessage("修改失败", c)
		global.GVA_LOG.Error("修改失败:" + err.Error())
		return
	}

	response.OkWithMessage("修改成功", c)
	return
}

type 结构请求_批量修改vip限制 struct {
	Id    []int `json:"Id"`    //id数组
	IsVip int   `json:"IsVip"` //1 解冻 2冻结
}

// Del批量修改
func (a *Api) C测试执行(c *gin.Context) {
	var 请求 struct {
		Id int `json:"Id"` //id
	}
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}

	defer func() {
		if err2 := recover(); err2 != nil {
			局_GoJa错误, ok := err2.(*goja.Exception)
			if ok {
				response.FailWithMessage("异常:可能JS函数传参或返回值类型错误,具体:"+局_GoJa错误.String(), c)
			} else {
				response.FailWithMessage("异常:可能JS函数传参或返回值类型错误,具体:js引擎未返回报错信息", c)
			}
			return
		}
	}()

	if !Ser_PublicJs.Id是否存在(请求.Id) {
		response.FailWithMessage("JS公共函数不存在", c)
		return
	}
	局_耗时 := time.Now().UnixMilli()

	var 局_PublicJs DB.DB_PublicJs

	局_PublicJs, err = Ser_PublicJs.Q取值2(请求.Id)

	if err != nil {
		response.FailWithMessage("JS公共函数不存在", c)
		return
	}

	if W文件_是否存在(global.GVA_CONFIG.Q取运行目录 + 局_PublicJs.Value) {
		局_PublicJs.Value = string(W文件_读入文件(global.GVA_CONFIG.Q取运行目录 + 局_PublicJs.Value))
	} else {
		response.FailWithMessage("js文件读取失败可能被删除", c)
		return
	}

	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	局_云函数型参数 := "{}"

	vm := Ser_Js.JS引擎初始化_用户(&AppInfo, &局_在线信息, &局_PublicJs)

	_, err = vm.RunString(局_PublicJs.Value)
	if 局_详细错误, ok := err.(*goja.Exception); ok {
		response.FailWithMessage("JS代码运行失败:"+局_详细错误.String(), c)
		return
	}
	var 局_待执行js函数名 func(string) interface{}
	ret := vm.Get(局_PublicJs.Name)
	if ret == nil {
		response.FailWithMessage("Js中没有["+局_PublicJs.Name+"()]函数", c)
		return
	}
	err = vm.ExportTo(ret, &局_待执行js函数名)
	if err != nil {
		response.FailWithMessage("Js绑定函数到变量失败", c)
		return
	}
	局_return := 局_待执行js函数名(局_云函数型参数)
	局_耗时 = time.Now().UnixMilli() - 局_耗时
	response.OkWithDetailed(gin.H{"Return": 局_return, "Time": 局_耗时}, "执行成功,耗时:"+strconv.FormatInt(局_耗时, 10), c)
	return
}
