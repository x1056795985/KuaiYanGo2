package TaskPool

import (
	"fmt"
	E "github.com/duolabmeng6/goefun/eTool"
	"github.com/gin-gonic/gin"
	"server/Service/Ser_PublicJs"
	"server/Service/Ser_TaskPool"
	"server/global"
	"server/structs/Http/response"
	DB "server/structs/db"
	"server/utils"
	"strconv"
)

type Api struct{}

// GetTaskTypeInfo
func (a *Api) GetInfo(c *gin.Context) {
	var 请求 结构请求_单id
	//{"Id":2}
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	TaskPool_类型, err := Ser_TaskPool.Task类型读取(请求.Id)
	if err != nil {
		response.FailWithMessage("读取失败,可能数据不存在Id:"+strconv.Itoa(请求.Id), c)
		return
	}
	response.OkWithDetailed(TaskPool_类型, "获取成功", c)
	return
}

type 结构请求_单id struct {
	Id int `json:"Id"`
}

type 结构请求_GetUserList struct {
	Page     int    `json:"Page"`     // 页
	Size     int    `json:"Size"`     // 页数量
	Type     int    `json:"Type"`     // 关键字类型  1 id 2 任务类型名
	Keywords string `json:"Keywords"` // 关键字
	Order    int    `json:"Order"`    // 0 倒序 1 正序
}

// GetList
// 获取任务类型列表
func (a *Api) GetList(c *gin.Context) {
	var 请求 结构请求_GetUserList
	//{"Type":"2","Size":10,"Page":1,"Status":"1","keywords":"1"}
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	局_DB := global.GVA_DB.Model(DB.TaskPool_类型{})

	if 请求.Order == 1 {
		局_DB.Order("Id ASC")
	} else {
		局_DB.Order("Id DESC")
	}

	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //id
			局_DB.Where("Id = ?", 请求.Keywords)
		case 2: //任务类型名称
			局_DB.Where("LOCATE(?, Name)>0 ", 请求.Keywords)
		}
	}
	var TaskPool []DB.TaskPool_类型
	var 总数 int64
	//Count(&总数) 必须放在where 后面 不然值会被清0
	err = 局_DB.Count(&总数).Select("").Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&TaskPool).Error

	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		return
	}
	response.OkWithDetailed(结构响应_GetList{TaskPool, 总数}, "获取成功", c)
	return
}

type TaskPool_类型带数量 struct {
	QueueCount int `json:"QueueCount"` //队列数量
	TaskCount  int `json:"TaskCount"`  //总计数量
	DB.TaskPool_类型
}

// GetList
// 获取任务类型列表
func (a *Api) GetList2(c *gin.Context) {
	var 请求 结构请求_GetUserList
	//{"Type":"2","Size":10,"Page":1,"Status":"1","keywords":"1"}
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	// 构建SQL查询语句
	sql := `
  SELECT  db_TaskPoolType.*,
           (SELECT  COUNT(*) FROM db_TaskPoolQueue WHERE  db_TaskPoolQueue.Tid =db_TaskPoolType.Id) AS QueueCount,
          (SELECT  COUNT(*) FROM db_TaskPoolData WHERE  db_TaskPoolData.Tid =db_TaskPoolType.Id) AS TaskCount
    FROM db_TaskPoolType
`

	// 添加条件
	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //id
			sql += " WHERE db_TaskPoolType.Id = ?"
		case 2: //任务类型名称
			sql += " WHERE LOCATE(?, db_TaskPoolType.Name) > 0"
		}
	}

	// 添加排序
	if 请求.Order == 1 {
		sql += " ORDER BY db_TaskPoolType.Id ASC"
	} else {
		sql += " ORDER BY db_TaskPoolType.Id DESC"
	}

	// 添加分页
	sql += " LIMIT ? OFFSET ?"

	// 执行查询
	var TaskPool []TaskPool_类型带数量
	var 总数 int64
	if 请求.Keywords != "" && (请求.Type == 1 || 请求.Type == 2) {
		err = global.GVA_DB.Raw(sql, 请求.Keywords, 请求.Size, (请求.Page-1)*请求.Size).Scan(&TaskPool).Error
	} else {
		err = global.GVA_DB.Raw(sql, 请求.Size, (请求.Page-1)*请求.Size).Scan(&TaskPool).Error
	}

	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		return
	}

	// 统计数量
	err = global.GVA_DB.Model(DB.TaskPool_类型{}).Count(&总数).Error
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		return
	}

	response.OkWithDetailed(结构响应_GetList{TaskPool, 总数}, "获取成功", c)
	return
}

type 结构响应_GetList struct {
	List  interface{} `json:"List"`  // 列表
	Count int64       `json:"Count"` // 总数
}

// Del批量删除
func (a *Api) Del批量删除(c *gin.Context) {
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
	影响行数 = db.Model(DB.TaskPool_类型{}).Where("Id IN ? ", 请求.Id).Delete("").RowsAffected
	if db.Error != nil {
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
	return
}

type 结构请求_ID数组 struct {
	Id []int `json:"Id"`
}

// save 保存
func (a *Api) Save(c *gin.Context) {
	var 请求 DB.TaskPool_类型
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
	var count int64

	err = global.GVA_DB.Model(DB.TaskPool_类型{}).Where("Id = ?", 请求.Id).Count(&count).Error
	// 没查到数据
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
	return
}

func (a *Api) New(c *gin.Context) {
	var 请求 DB.TaskPool_类型
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
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
	err = Ser_TaskPool.Task类型创建(请求.Name, 请求.HookSubmitDataStart, 请求.HookSubmitDataEnd, 请求.HookReturnDataStart, 请求.HookReturnDataEnd)

	if err != nil {
		response.FailWithMessage("添加失败:"+err.Error(), c)
		return
	}

	创建不存在的Hook函数(请求)
	response.OkWithMessage("添加成功", c)
	return
}

const 初始模板函数头 = `function `
const 初始模板函数尾 = `(任务JSON格式参数) {
    /*
    return $用户在线信息; // {"Key":"aaaaaa","Status":1,"Tab":"AMD Ryzen 7 6800H with Radeon Graphics         |178BFBFF00A40F41","Uid":21,"User":"aaaaaa"}
    return $应用信息 // {"AppId":10001,"AppName":"演示对接账密限时Rsa交换密匙","Status":3,"VipData":"{\n\"VipData\":\"这里的数据,只有登录成功并且账号会员不过期才会传输出去的数据\",\n\"VipData2\":\"这里的数据,只有登录成功并且账号会员不过期才会传输出去的数据\"\n}
    return $用户在线信息.Uid

  var 局_用户信息 = $api_用户Id取详情($用户在线信息) //{"Id":21,"User":"aaaaaa","PassWord":"af15d5fdacd5fdfea300e88a8e253e82","Phone":"13109812593","Email":"1056795985@qq.com","Qq":"1059795985","SuperPassWord":"af15d5fdacd5fdfea300e88a8e253e82","Status":1,"Rmb":91.39,"Note":"","RealNameAttestation":"","Role":0,"UPAgentId":0,"AgentDiscount":0,"LoginAppid":10000,"LoginIp":"","LoginTime":1519454315,"RegisterIp":"113.235.144.55","RegisterTime":1519454315}
    例子随机 拦截任务提交

    任务JSON格式参数 = 任务JSON格式参数.replace(/'/g, '"') //因为易语言 双引号不方便,所以到js里换成替换单引号成双引号 //注意永远不要相信客户端传参,建议直接在hook函数内固定金额,这里只是测试
   var 局_形参对象 = JSON.parse(任务JSON格式参数); //使用JSON.parse() 将JSON字符串转为JS对象;
    局_结果 = $api_用户Id余额增减($用户在线信息, -局_形参对象.a, "测试任务池Hook内扣余额") //扣款需要时负数需要前面加 - 负号 直接操作就行, 内部会自动判断不用再js先判断余额是否充足,
    if (!局_结果.IsOk) {
        $拦截原因 = "扣费失败" + 局_结果.Err
    }

         if (Math.floor(Math.random() * 10) > 5) {
         $拦截原因 = "如果值不为空,则任务拦截,响应拦截原因"
         }
         */
    return 任务JSON格式参数 //任务JSON格式文本型参数,可以在这里修改内容  然后返回
}`

func 创建不存在的Hook函数(请求 DB.TaskPool_类型) {

	var 局数组_函数名 = []string{请求.HookSubmitDataStart, 请求.HookSubmitDataEnd, 请求.HookReturnDataStart, 请求.HookReturnDataEnd}
	for 索引, _ := range 局数组_函数名 {
		if 局数组_函数名[索引] != "" && !Ser_PublicJs.Name是否存在(2, 局数组_函数名[索引]) {
			var 局_hook函数 = DB.DB_PublicJs{
				AppId: 2,
				Type:  1,
				Note:  "任务类型:" + 请求.Name + ",自动创建",
			}
			局_hook函数.Name = 局数组_函数名[索引]
			局_hook函数.Value = "/云函数/" + 局数组_函数名[索引] + ".js"
			_ = global.GVA_DB.Model(DB.DB_PublicJs{}).Create(&局_hook函数).Error
			_ = E.E文件_保存(global.GVA_CONFIG.Q取运行目录+"/云函数/"+局数组_函数名[索引]+".js", 初始模板函数头+局数组_函数名[索引]+初始模板函数尾)
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

type 结构请求_批量修改状态 struct {
	Id     []int `json:"Id"`     //任务类型id数组
	Status int   `json:"Status"` //1 正常 2 维护
}

func (a *Api) Set修改状态(c *gin.Context) {
	var 请求 结构请求_批量修改状态
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

	if 请求.Status != 1 && 请求.Status != 2 {
		response.FailWithMessage("修改失败:Status状态代码错误", c)
		return
	}

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
	return
}

// Q清空队列
func (a *Api) Q清空队列(c *gin.Context) {
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

	影响行数, err := Ser_TaskPool.Task队列清除指定Tid(请求.Id)
	if err != nil {
		response.FailWithMessage("清空失败:"+err.Error(), c)
		return
	}

	response.OkWithMessage("清空成功,数量:"+strconv.Itoa(影响行数), c)
	return
}
