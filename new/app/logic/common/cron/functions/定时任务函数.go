package functions

import (
	"EFunc/utils"
	"errors"
	"fmt"
	"github.com/dop251/goja"
	"github.com/gin-gonic/gin"
	"github.com/imroc/req/v3"
	App服务 "server/Service/Ser_AppInfo"
	"server/Service/Ser_Js"
	"server/Service/Ser_PublicJs"
	"server/Service/Ser_TaskPool"
	"server/global"
	"server/new/app/logic/admin/L_pay"
	"server/new/app/logic/common/log"
	"server/new/app/models/db"
	"server/new/app/service"
	DB "server/structs/db"
	utils2 "server/utils"
	"strconv"
	"strings"
	"time"
)

func S刷新数据库定时任务2() {
	_ = S刷新数据库定时任务(false)

}

// 定时刷新数据库,或redis订阅刷新
func S刷新数据库定时任务(主动 bool) error {
	if global.GVA_DB == nil {
		return nil
	}
	c := &global.Cron定时任务
	c.L临界.Lock()
	defer c.L临界.Unlock()

	局_is刷新 := true                            //默认刷新
	局_临时, ok := global.H缓存.Get("map集群任务Hash") //读取缓存,或redis 该值
	if ok {                                   //如果存在
		局_缓存hash, ok2 := 局_临时.(string)        //且置可以断言成功
		if ok2 && 局_缓存hash == c.Map集群任务Hash { //并且和本地上次刷新值相同,那就不刷新了,
			局_is刷新 = false
		}
	}
	_, 局_是否存在 := global.H缓存.Get("map集群任务Hash") //如果不存在也刷新
	if 局_is刷新 || 主动 || !局_是否存在 {

		var S = service.NewCronService(global.GVA_DB)
		tx := *global.GVA_DB
		infoArr, err := S.GetAllInfo(&tx, 1)
		if err != nil {
			global.GVA_LOG.Error("刷新数据库定时任务失败:" + err.Error())
			return err
		}

		for 键名 := range c.Map集群任务列表 {
			// 手动释放Job1所指向的内存
			//c.map集群任务列表[键名].Job = nil
			c.Cron.Remove(c.Map集群任务列表[键名].EntryID)
		}
		//系统自带的
		infoArr = append(infoArr, db.DB_Cron{Id: -1, Status: 1, IsLog: 2, Type: -1, Name: "在线列表定时注销已过期", Cron: "0 */1 * * * *"})   //每分钟执行一次
		infoArr = append(infoArr, db.DB_Cron{Id: -2, Status: 1, IsLog: 2, Type: -2, Name: "在线列表定时删除已过期", Cron: "0 */1 * * * *"})   //每分钟执行一次
		infoArr = append(infoArr, db.DB_Cron{Id: -3, Status: 1, IsLog: 2, Type: -3, Name: "任务池Task数据删除过期", Cron: "0 */1 * * * *"}) //每分钟执行一次
		infoArr = append(infoArr, db.DB_Cron{Id: -4, Status: 1, IsLog: 2, Type: -4, Name: "定时关闭待支付订单", Cron: "0 */1 * * * *"})     //每分钟执行一次
		infoArr = append(infoArr, db.DB_Cron{Id: -5, Status: 1, IsLog: 2, Type: -5, Name: "删除已过期唯一积分记录", Cron: "0 0 0 * * ?"})     //每天0点执行一次
		infoArr = append(infoArr, db.DB_Cron{Id: -6, Status: 1, IsLog: 2, Type: -6, Name: "统计在线应用用户总量", Cron: "0 0 * * * ?"})      //每小时执行一次

		hashStr := ""
		for 索引, _ := range infoArr {
			if infoArr[索引].Status == 1 {
				err = c.T添加集群任务(infoArr[索引], T通用任务包装函数)
				if err != nil {
					global.GVA_LOG.Error("T添加任务定时任务id:" + strconv.Itoa(infoArr[索引].Id) + ",失败:" + err.Error())
				}
				//所有影响任务执行的参数,都需要加入哈希值
				hashStr += strconv.Itoa(infoArr[索引].Id) + "|" + infoArr[索引].Cron + "|" + strconv.Itoa(infoArr[索引].IsLog) + "|" + strconv.Itoa(infoArr[索引].Type) + "|" + infoArr[索引].RunText
			}
		}
		hashStr = utils2.Md5String(hashStr)
		c.Map集群任务Hash = hashStr
		if 主动 {
			global.GVA_LOG.Info("已刷新定时任务,新任务hash:" + hashStr)
			global.H缓存.Set("map集群任务Hash", hashStr, -1) //谁更新数据库任务信息,谁主动更新缓存的hash值
		}
	}
	return nil
}

// D定时集群任务预处理分发
func T通用任务包装函数(时间戳 int64, R任务数据 db.DB_Cron) {
	//global.GVA_LOG.Info(fmt.Sprintf("T通用任务包装函数被触发:%v \r\n", R任务数据))
	局_hast := utils2.Md5String("hash" + strconv.Itoa(int(时间戳)) + "|" + strconv.Itoa(R任务数据.Id) + "|" + R任务数据.Name)
	err := global.H缓存.Add(局_hast, 1, time.Second*time.Duration(300))
	if err != nil { //抢锁失败,跳过.被别的机器抢到执行了
		return
	}
	_, _ = T通用任务执行函数2(时间戳, R任务数据)

}

// 具体执行,无抢锁
func T通用任务执行函数2(时间戳 int64, R任务数据 db.DB_Cron) (string, error) {
	返回 := ""
	c := gin.Context{}

	var err error
	//抢到了任务,开始执行
	switch R任务数据.Type {
	case -1: //在线列表定时注销已过期
		D定时任务_注销已过期的Token(&c)
		return "", nil
	case -2: //在线列表定时删除已过期
		D定时任务_删除已过期的Token(&c)
		return "", nil
	case -3: //任务池Task数据删除过期
		Ser_TaskPool.Task数据删除过期()
		return "", nil
	case -4: //关闭超时订单
		err = L_pay.G关闭超时订单()
		return "", err
	case -5: //删除已过期唯一积分记录
		D定时任务_删除已过期唯一积分记录(&c)
		return "", err
	case -6: //删除已过期唯一积分记录
		D定时任务_统计应用在线用户总数(&c)
		return "", err
	case 1: //1,http请求,2公共js函数,3 SQL 4 shell"`
		返回, err = D定时任务_http请求(时间戳, R任务数据)
	case 2: //1,http请求,2公共js函数,3 SQL 4 shell"`
		返回, err = D定时任务_执行公共函数(时间戳, R任务数据)
	case 3: //1,http请求,2公共js函数,3 SQL 4 shell"`
		返回, err = D定时任务_SQL(时间戳, R任务数据)
	case 4: //1,http请求,2公共js函数,3 SQL 4 shell"`
		//shell    因为可能造成,登陆后台就可以获取服务器权限,所以暂时搁置
	}
	var 结果 int8 = 1
	if err != nil {
		返回 = err.Error()
		结果 = 2
	}

	if R任务数据.IsLog == 1 {
		tx := *global.GVA_DB
		var S = service.S_CronLog{}
		err = S.Create(&tx, db.DB_Cron_log{
			CronID:     R任务数据.Id,
			RunTime:    时间戳,
			Type:       R任务数据.Type,
			RunText:    R任务数据.RunText,
			Result:     结果,
			ReturnText: 返回,
		})
		if err != nil {
			global.GVA_LOG.Error("D定时任务_日志插入失败:" + err.Error())
		}
	}
	return 返回, err
}

func D定时任务_注销已过期的Token(c *gin.Context) {
	if global.GVA_DB == nil {
		return
	}
	tx := *global.GVA_DB
	s := service.NewLinksToken(c, &tx)
	err := s.Z注销已过期的Token()
	if err != nil {
		global.GVA_LOG.Error("在线列表定时注销已过期失败:" + err.Error())
	}
}

func D定时任务_删除已过期的Token(c *gin.Context) {
	if global.GVA_DB == nil {
		return
	}

	tx := *global.GVA_DB
	s := service.NewLinksToken(c, &tx)
	err := s.S删除已过期的Token()
	if err != nil {
		global.GVA_LOG.Error("在线列表定时删除已过期失败:" + err.Error())
	}
}
func D定时任务_统计应用在线用户总数(c *gin.Context) {
	if global.GVA_DB == nil {
		return
	}
	tx := *global.GVA_DB
	var results []db.DB_TongJiZaiXian
	//删除createTime 时间戳超过一年的

	err := tx.Raw(`
    SELECT LoginAppid AS appId, COUNT(*) AS count 
    FROM db_links_Token 
    WHERE Status = 1 AND LoginAppid>10000
    GROUP BY LoginAppid
`).Scan(&results).Error
	if err != nil {
		global.GVA_LOG.Error("D定时任务_统计应用在线用户总数失败:" + err.Error())
		return
	}
	if len(results) == 0 {
		return
	}

	var s = service.NewTongJIzaiXian(c, &tx)
	时间戳 := time.Now().Unix()
	var 局_总计数量 int64
	for i, _ := range results {
		局_总计数量 += results[i].Count
		results[i].CreatedAt = 时间戳
	}
	results = append(results, db.DB_TongJiZaiXian{
		AppId:     0,
		Count:     局_总计数量,
		CreatedAt: 时间戳,
	})
	err = s.BatchCreate(&results)
	if err != nil {
		global.GVA_LOG.Error("D定时任务_统计应用在线用户总数失败:" + err.Error())
	}
	// 删除一年前的数据
	err = tx.Where("createdAt < ?", time.Now().AddDate(0, 0, -365).Unix()).Delete(&db.DB_TongJiZaiXian{}).Error

}
func D定时任务_http请求(时间戳 int64, R任务数据 db.DB_Cron) (string, error) {
	client := req.C().EnableInsecureSkipVerify() // Use C() to create a client.
	resp, err := client.R().Get(R任务数据.RunText)
	返回 := ""
	if err != nil {
		return 返回, err
	} else {
		返回 = resp.String()
	}
	return 返回, nil
}

func D定时任务_SQL(时间戳 int64, R任务数据 db.DB_Cron) (string, error) {
	返回 := "执行成功"
	if global.GVA_DB == nil {
		return 返回, errors.New("执行失败:未连接数据库")
	} else {
		局_db := *global.GVA_DB
		// 减少余额
		sql := strings.Replace(R任务数据.RunText, "{{十位时间戳}}", strconv.FormatInt(time.Now().Unix(), 10), -1)
		tx := 局_db.Exec(sql)
		return "影响行数:" + strconv.Itoa(int(tx.RowsAffected)), tx.Error
	}
}

func D定时任务_执行公共函数(时间戳 int64, R任务数据 db.DB_Cron) (string, error) {
	返回 := ""
	局_函数名 := utils.W文本_取文本左边(R任务数据.RunText, "(")
	局_云函数型参数 := utils.W文本_取出中间文本(R任务数据.RunText, "(", ")")
	局_云函数型参数 = strings.TrimSpace(局_云函数型参数)
	局_云函数型参数 = strings.Replace(局_云函数型参数, `“`, `"`, -1) //中文引号改英文引号
	runes := []rune(局_云函数型参数)
	firstChar := ""
	if len(runes) > 1 {
		firstChar = string(runes[0])
	}

	局_云函数型参数 = utils.W文本_取出中间文本(R任务数据.RunText, firstChar, firstChar) //获取字符串参数
	局_js数据, err := Ser_PublicJs.P取值2(Ser_PublicJs.Js类型_公共函数, 局_函数名)
	if err != nil {
		return 返回, errors.New("获取全局公共js函数失败:" + err.Error())
	}
	vm := Ser_Js.JS引擎初始化_用户(&DB.DB_AppInfo{}, &DB.DB_LinksToken{}, &局_js数据)
	_, err = vm.RunString(局_js数据.Value)
	if 局_详细错误, ok := err.(*goja.Exception); ok {
		return 返回, errors.New("JS代码运行失败:" + 局_详细错误.String())
	}
	var 局_待执行js函数名 func(string) interface{}
	a := vm.Get(局_函数名)
	if a == nil {
		return 返回, errors.New("Js中没有[" + 局_函数名 + "()]函数")
	}
	err = vm.ExportTo(a, &局_待执行js函数名)
	if err != nil {
		return 返回, errors.New("Js绑定函数到变量失败")
	}

	返回 = fmt.Sprintf("%v", 局_待执行js函数名(局_云函数型参数)) //不管是什么类型,直接转文本

	return 返回, nil
}
func D定时任务_删除已过期唯一积分记录(c *gin.Context) {
	if global.GVA_DB == nil {
		return
	}

	tx := *global.GVA_DB
	局_全部应用 := App服务.AppInfo取map列表Int()

	for key, _ := range 局_全部应用 {
		if key <= 10000 { //过滤掉其他的
			continue
		}
		s := service.NewUniqueNumLog(c, &tx, key)
		局_数量, err := s.Delete已过期()
		if err != nil {
			global.GVA_LOG.Error("删除已过期唯一积分记录失败:" + err.Error())
		}
		if 局_数量 > 0 {
			global.GVA_LOG.Info("删除已过期唯一积分记录:" + strconv.Itoa(int(局_数量)))
		}
	}

}
