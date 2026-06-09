package controller

import (
	"EFunc/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"server/Service/Ser_PublicJs"
	"server/Service/Ser_TaskPool"
	"server/global"
	"server/new/app/controller/Common"
	"server/structs/Http/response"
	DB "server/structs/db"
	"strconv"
)

type TaskPoolFull struct {
	Common.Common
}

func NewTaskPoolFullController() *TaskPoolFull {
	return &TaskPoolFull{}
}

type 请求_TaskPoolGetInfo struct {
	Id int `json:"id"`
}

type 请求_TaskPoolGetList struct {
	Page     int    `json:"page"`
	Size     int    `json:"size"`
	Type     int    `json:"type"`
	Keywords string `json:"keywords"`
	Order    int    `json:"order"`
}

type 请求_TaskPoolDelete struct {
	Id []int `json:"id"`
}

type 请求_TaskPoolSetStatus struct {
	Id     []int `json:"id"`
	Status int   `json:"status"`
}

type 请求_TaskPoolClearQueue struct {
	Id []int `json:"id"`
}

type 请求_TaskPoolUuidAddQueue struct {
	Uuid string `json:"uuid"`
}

type 请求_TaskPoolBatchUuidAddQueue struct {
	Uuid []string `json:"uuids"`
}

type TaskPool_类型带数量 struct {
	QueueCount int `json:"queueCount"`
	TaskCount  int `json:"taskCount"`
	DB.TaskPool_类型
}

type 响应_TaskPoolGetList struct {
	List  interface{} `json:"list"`
	Count int64       `json:"count"`
}

const 初始模板函数头 = `function `
const 初始模板函数尾 = `(任务JSON格式参数) {
    /*
    return $用户在线信息; // {"Key":"aaaaaa","Status":1,"Tab":"AMD Ryzen 7 6800H with Radeon Graphics         |178BFBFF00A40F41","Uid":21,"User":"aaaaaa"}
    return $应用信息 // {"AppId":10001,"AppName":"演示对接账密限时Rsa交换密匙","Status":3,"VipData":"..."}
    return $用户在线信息.Uid
    var 局_用户信息 = $api_用户Id取详情($用户在线信息)
    例子随机 拦截任务提交
    任务JSON格式参数 = 任务JSON格式参数.replace(/'/g, '"')
    var 局_形参对象 = JSON.parse(任务JSON格式参数);
    局_结果 = $api_用户Id增减余额($用户在线信息, -局_形参对象.a, "测试任务池Hook内扣余额")
    if (!局_结果.IsOk) {
        $拦截原因 = "扣费失败" + 局_结果.Err
    }
    if (Math.floor(Math.random() * 10) > 5) {
        $拦截原因 = "如果值不为空,则任务拦截,响应拦截原因"
    }
    */
    return 任务JSON格式参数
}`

// Info 获取任务类型详情
func (C *TaskPoolFull) Info(c *gin.Context) {
	var 请求 请求_TaskPoolGetInfo
	if !C.ToJSON(c, &请求) {
		return
	}

	TaskPool_类型, err := Ser_TaskPool.Task类型读取(请求.Id)
	if err != nil {
		response.FailWithMessage("读取失败,可能数据不存在Id:"+strconv.Itoa(请求.Id), c)
		return
	}
	response.OkWithDetailed(TaskPool_类型, "获取成功", c)
}

// GetList 获取任务类型列表
func (C *TaskPoolFull) GetList(c *gin.Context) {
	var 请求 请求_TaskPoolGetList
	if !C.ToJSON(c, &请求) {
		return
	}

	// 第1步: 只查任务类型表(小表,毫秒级),拿到当前页的Tid列表
	局_DB := global.GVA_DB.Model(DB.TaskPool_类型{})
	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //id
			局_DB.Where("Id = ?", 请求.Keywords)
		case 2: //任务类型名称
			局_DB.Where("LOCATE(?, Name) > 0", 请求.Keywords)
		}
	}

	var 总数 int64
	err := 局_DB.Count(&总数).Error
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		return
	}

	if 请求.Order == 1 {
		局_DB.Order("Sort DESC, Id ASC")
	} else {
		局_DB.Order("Sort DESC, Id DESC")
	}

	var 局_类型列表 []DB.TaskPool_类型
	err = 局_DB.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&局_类型列表).Error
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		return
	}

	if len(局_类型列表) == 0 {
		response.OkWithDetailed(响应_TaskPoolGetList{[]TaskPool_类型带数量{}, 总数}, "获取成功", c)
		return
	}

	// 第2步: 只对当前页出现的Tid统计数量,而不是全表关联子查询
	局_Tid列表 := make([]int, len(局_类型列表))
	for i, item := range 局_类型列表 {
		局_Tid列表[i] = item.Id
	}

	type 结构_数量统计 struct {
		Tid   int `json:"Tid"`
		Count int `json:"Count"`
	}
	var 局_队列数量 []结构_数量统计
	var 局_任务数量 []结构_数量统计

	// 查队列数量 - 只查当前页的Tid
	err = global.GVA_DB.Model(DB.TaskPool_队列{}).
		Select("Tid, COUNT(*) AS Count").
		Where("Tid IN ?", 局_Tid列表).
		Group("Tid").
		Find(&局_队列数量).Error
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		return
	}

	// 查任务数量 - 只查当前页的Tid + 30天时间筛选
	时间戳30天前 := int(utils.S时间_取现行时间戳() - (86400 * 30))
	err = global.GVA_DB.Model(DB.DB_TaskPoolData{}).
		Select("Tid, COUNT(*) AS Count").
		Where("Tid IN ? AND TimeStart > ?", 局_Tid列表, 时间戳30天前).
		Group("Tid").
		Find(&局_任务数量).Error
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		return
	}

	// 构建Tid->Count映射,组装最终结果
	局_队列Map := make(map[int]int, len(局_队列数量))
	for _, v := range 局_队列数量 {
		局_队列Map[v.Tid] = v.Count
	}
	局_任务Map := make(map[int]int, len(局_任务数量))
	for _, v := range 局_任务数量 {
		局_任务Map[v.Tid] = v.Count
	}

	var TaskPool []TaskPool_类型带数量
	for _, item := range 局_类型列表 {
		TaskPool = append(TaskPool, TaskPool_类型带数量{
			QueueCount:   局_队列Map[item.Id],
			TaskCount:    局_任务Map[item.Id],
			TaskPool_类型: item,
		})
	}

	response.OkWithDetailed(响应_TaskPoolGetList{TaskPool, 总数}, "获取成功", c)
}

// New 新建任务类型
func (C *TaskPoolFull) New(c *gin.Context) {
	var 请求 DB.TaskPool_类型
	if !C.ToJSON(c, &请求) {
		return
	}
	if 请求.Id != 0 {
		response.FailWithMessage("添加不能有id值", c)
		return
	}
	if 局_临时文本 := 判断即将创建的Hook函数是否允许(请求); 局_临时文本 != "" {
		response.FailWithMessage(局_临时文本, c)
		return
	}
	if len(请求.Name) < 4 {
		response.FailWithMessage("类型名称长度必须>4", c)
		return
	}

	err := Ser_TaskPool.Task类型创建(请求.Name, 请求.HookSubmitDataStart, 请求.HookSubmitDataEnd, 请求.HookReturnDataStart, 请求.HookReturnDataEnd)
	if err != nil {
		response.FailWithMessage("添加失败:"+err.Error(), c)
		return
	}

	创建不存在的Hook函数(请求)
	response.OkWithMessage("添加成功", c)
}

// Save 保存任务类型
func (C *TaskPoolFull) Save(c *gin.Context) {
	var 请求 DB.TaskPool_类型
	if !C.ToJSON(c, &请求) {
		return
	}
	if 请求.Id <= 0 {
		response.FailWithMessage("Id错误", c)
		return
	}

	var count int64
	_ = global.GVA_DB.Model(DB.TaskPool_类型{}).Where("Id = ?", 请求.Id).Count(&count).Error
	if count == 0 {
		response.FailWithMessage("任务类型不存在", c)
		return
	}
	if len(请求.Name) < 4 {
		response.FailWithMessage("类型名称长度必须>4", c)
		return
	}
	if 局_临时文本 := 判断即将创建的Hook函数是否允许(请求); 局_临时文本 != "" {
		response.FailWithMessage(局_临时文本, c)
		return
	}

	m := map[string]interface{}{
		"Name":                请求.Name,
		"HookReturnDataStart": 请求.HookReturnDataStart,
		"HookReturnDataEnd":   请求.HookReturnDataEnd,
		"HookSubmitDataStart": 请求.HookSubmitDataStart,
		"HookSubmitDataEnd":   请求.HookSubmitDataEnd,
	}

	var db = global.GVA_DB.Model(DB.TaskPool_类型{}).Where("Id=?", 请求.Id).Updates(&m)
	if db.Error != nil {
		fmt.Printf(db.Error.Error())
		response.FailWithMessage("保存失败", c)
		return
	}
	创建不存在的Hook函数(请求)
	response.OkWithMessage("保存成功"+strconv.Itoa(int(db.RowsAffected)), c)
}

// SetStatus 批量修改状态
func (C *TaskPoolFull) SetStatus(c *gin.Context) {
	var 请求 请求_TaskPoolSetStatus
	if !C.ToJSON(c, &请求) {
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

	var err error
	if 请求.Status == 2 {
		err = global.GVA_DB.Model(DB.TaskPool_类型{}).Where("Id IN ? ", 请求.Id).Update("Status", 2).Error
	} else {
		err = global.GVA_DB.Model(DB.TaskPool_类型{}).Where("Id IN ? ", 请求.Id).Update("Status", 1).Error
	}
	if err != nil {
		response.FailWithMessage("修改失败", c)
		global.GVA_LOG.Error("修改失败:" + err.Error())
		return
	}
	response.OkWithMessage("修改成功", c)
}

// Delete 批量删除任务类型
func (C *TaskPoolFull) Delete(c *gin.Context) {
	var 请求 请求_TaskPoolDelete
	if !C.ToJSON(c, &请求) {
		return
	}
	if len(请求.Id) == 0 {
		response.FailWithMessage("Id数组为空", c)
		return
	}

	var db = global.GVA_DB
	影响行数 := db.Model(DB.TaskPool_类型{}).Where("Id IN ? ", 请求.Id).Delete("").RowsAffected
	if db.Error != nil {
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
}

// ClearQueue 清空队列
func (C *TaskPoolFull) ClearQueue(c *gin.Context) {
	var 请求 请求_TaskPoolClearQueue
	if !C.ToJSON(c, &请求) {
		return
	}
	if len(请求.Id) == 0 {
		response.FailWithMessage("Id数组为空", c)
		return
	}

	影响行数, err := Ser_TaskPool.Task队列清除指定Tid(请求.Id)
	if err != nil {
		response.FailWithMessage("清空失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("清空成功,数量:"+strconv.Itoa(影响行数), c)
}

// UuidAddQueue Uuid重新加入队列
func (C *TaskPoolFull) UuidAddQueue(c *gin.Context) {
	var 请求 DB.TaskPool_数据_精简
	if !C.ToJSON(c, &请求) {
		return
	}
	if len(请求.Uuid) != 36 {
		response.FailWithMessage("uuid错误", c)
		return
	}

	err := Ser_TaskPool.Uuid_添加到队列(请求.Uuid)
	if err != nil {
		response.FailWithMessage("重新加入队列失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("操作成功", c)
}

// BatchUuidAddQueue 批量Uuid重新加入队列
func (C *TaskPoolFull) BatchUuidAddQueue(c *gin.Context) {
	var 请求 struct {
		Uuid []string `json:"uuids"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	局_成功数量 := 0
	for i := range len(请求.Uuid) {
		err := Ser_TaskPool.Uuid_添加到队列(请求.Uuid[i])
		if err == nil {
			局_成功数量++
		}
	}
	response.OkWithMessage("重新加入队列,成功:"+utils.D到文本(局_成功数量)+",失败:"+utils.D到文本(len(请求.Uuid)-局_成功数量), c)
}

func 创建不存在的Hook函数(请求 DB.TaskPool_类型) {
	var 局数组_函数名 = []string{请求.HookSubmitDataStart, 请求.HookSubmitDataEnd, 请求.HookReturnDataStart, 请求.HookReturnDataEnd}
	for 索引 := range 局数组_函数名 {
		if 局数组_函数名[索引] != "" && !Ser_PublicJs.Name是否存在(2, 局数组_函数名[索引]) {
			var 局_hook函数 = DB.DB_PublicJs{
				AppId: 2,
				Type:  1,
				Note:  "任务类型:" + 请求.Name + ",自动创建",
			}
			局_hook函数.Name = 局数组_函数名[索引]
			局_hook函数.Value = "/云函数/" + 局数组_函数名[索引] + ".js"
			_ = global.GVA_DB.Model(DB.DB_PublicJs{}).Create(&局_hook函数).Error
			_ = utils.W文件_保存(global.GVA_CONFIG.Q取运行目录+"/云函数/"+局数组_函数名[索引]+".js", 初始模板函数头+局数组_函数名[索引]+初始模板函数尾)
		}
	}
}

func 判断即将创建的Hook函数是否允许(请求 DB.TaskPool_类型) string {
	if utils.W文本_是否包含关键字(请求.HookSubmitDataStart, "/") || utils.W文本_是否包含关键字(请求.HookSubmitDataStart, ".") {
		return "Hook任务创建入库前,函数名不能包含[ / ]或[ . ]符号"
	}
	if utils.W文本_是否包含关键字(请求.HookSubmitDataEnd, "/") || utils.W文本_是否包含关键字(请求.HookSubmitDataEnd, ".") {
		return "Hook任务创建入库后,函数名不能包含[ / ]或[ . ]符号"
	}
	if utils.W文本_是否包含关键字(请求.HookReturnDataStart, "/") || utils.W文本_是否包含关键字(请求.HookReturnDataStart, ".") {
		return "Hook任务执行入库前,函数名不能包含[ / ]或[ . ]符号"
	}
	if utils.W文本_是否包含关键字(请求.HookReturnDataEnd, "/") || utils.W文本_是否包含关键字(请求.HookReturnDataEnd, ".") {
		return "Hook任务执行入库后,函数名不能包含[ / ]或[ . ]符号"
	}
	return ""
}
