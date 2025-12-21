package Ser_Chare

import (
	"EFunc/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_AppUser"
	"server/Service/Ser_Ka"
	"server/Service/Ser_KaClass"
	"server/Service/Ser_LinkUser"
	"server/Service/Ser_UserClass"
	"server/global"
	dbm "server/new/app/models/db"
	DB "server/structs/db"
	"sort"
	"strconv"
	"strings"
	"time"
)

const 系统演示模式 = 1

func Get在线用户Ip地图分布统计(c *gin.Context) []gin.H {

	/*[
	  {name: "河北省", value: 100},
	  {name: "山西省", value: 90},
	  {name: "辽宁省", value: 40},
	  {name: "吉林省", value: 50},
	  {name: "黑龙江省", value: 60},
	  {name: "江苏省", value: 20},
	  {name: "浙江省", value: 8},
	  {name: "安徽省", value: 20},
	  {name: "福建省", value: 46},
	  {name: "江西省", value: 32},
	  {name: "山东省", value: 2},
	  {name: "河南省", value: 2},
	  {name: "湖北省", value: 26},
	  {name: "湖南省", value: 30},
	  {name: "广东省", value: 29},
	  {name: "海南省", value: 20},
	  {name: "四川省", value: 212},
	  {name: "贵州省", value: 235},
	  {name: "云南省", value: 20},
	  {name: "陕西省", value: 289},
	  {name: "甘肃省", value: 274},
	  {name: "青海省", value: 260},
	  {name: "台湾省", value: 244},
	  {name: "内蒙古自治区", value: 235},
	  {name: "广西壮族自治区", value: 27},
	  {name: "西藏自治区", value: 20},
	  {name: "宁夏回族自治区", value: 20},
	  {name: "新疆维吾尔自治区", value: 20},
	  {name: "北京市", value: 20},
	  {name: "天津市", value: 20},
	  {name: "上海市", value: 20},
	  {name: "重庆市", value: 20},
	  {name: "香港特别行政区", value: 20},
	  {name: "澳门特别行政区", value: 20},
	  {name: "南海诸岛", value: 8},
	]*/
	if global.GVA_Viper.GetInt("系统模式") == 系统演示模式 {
		var Data = make([]gin.H, 35)

		Data[0] = gin.H{"name": "河北省", "value": 100}
		Data[1] = gin.H{"name": "山西省", "value": 90}
		Data[2] = gin.H{"name": "辽宁省", "value": 40}
		Data[3] = gin.H{"name": "吉林省", "value": 50}
		Data[4] = gin.H{"name": "黑龙江省", "value": 60}
		Data[5] = gin.H{"name": "江苏省", "value": 20}
		Data[6] = gin.H{"name": "浙江省", "value": 8}
		Data[7] = gin.H{"name": "安徽省", "value": 20}
		Data[8] = gin.H{"name": "福建省", "value": 46}
		Data[9] = gin.H{"name": "江西省", "value": 32}
		Data[10] = gin.H{"name": "山东省", "value": 2}
		Data[11] = gin.H{"name": "河南省", "value": 2}
		Data[12] = gin.H{"name": "湖北省", "value": 26}
		Data[13] = gin.H{"name": "湖南省", "value": 30}
		Data[14] = gin.H{"name": "广东省", "value": 29}
		Data[15] = gin.H{"name": "海南省", "value": 20}
		Data[16] = gin.H{"name": "四川省", "value": 212}
		Data[17] = gin.H{"name": "贵州省", "value": 235}
		Data[18] = gin.H{"name": "云南省", "value": 20}
		Data[19] = gin.H{"name": "陕西省", "value": 289}
		Data[20] = gin.H{"name": "甘肃省", "value": 274}
		Data[21] = gin.H{"name": "青海省", "value": 260}
		Data[22] = gin.H{"name": "台湾省", "value": 244}
		Data[23] = gin.H{"name": "内蒙古自治区", "value": 235}
		Data[24] = gin.H{"name": "广西壮族自治区", "value": 27}
		Data[25] = gin.H{"name": "西藏自治区", "value": 20}
		Data[26] = gin.H{"name": "宁夏回族自治区", "value": 20}
		Data[27] = gin.H{"name": "新疆维吾尔自治区", "value": 20}
		Data[28] = gin.H{"name": "北京市", "value": 20}
		Data[29] = gin.H{"name": "天津市", "value": 20}
		Data[30] = gin.H{"name": "上海市", "value": 20}
		Data[31] = gin.H{"name": "重庆市", "value": 20}
		Data[32] = gin.H{"name": "香港特别行政区", "value": 20}
		Data[33] = gin.H{"name": "澳门特别行政区", "value": 20}
		Data[34] = gin.H{"name": "南海诸岛", "value": 8}

		return Data
	}

	Data缓存, ok := global.H缓存.Get("图表数据_Get在线用户Ip地图分布统计")
	if ok {
		return Data缓存.([]gin.H)
	}

	局_耗时 := time.Now().Unix()
	// 执行SQL查询
	rows, err := global.GVA_DB.Raw(`SELECT COUNT(*) AS count, province
FROM (
    SELECT IPCity, 
           CASE 
               WHEN IPCity LIKE '%北京%' THEN '北京市'
               WHEN IPCity LIKE '%上海%' THEN '上海市'
               WHEN IPCity LIKE '%天津%' THEN '天津市'
               WHEN IPCity LIKE '%重庆%' THEN '重庆市'
               WHEN IPCity LIKE '%河北%' THEN '河北省'
               WHEN IPCity LIKE '%山西%' THEN '山西省'
               WHEN IPCity LIKE '%内蒙古%' THEN '内蒙古自治区'
               WHEN IPCity LIKE '%辽宁%' THEN '辽宁省'
               WHEN IPCity LIKE '%吉林%' THEN '吉林省'
               WHEN IPCity LIKE '%黑龙江%' THEN '黑龙江省'
               WHEN IPCity LIKE '%江苏%' THEN '江苏省'
               WHEN IPCity LIKE '%浙江%' THEN '浙江省'
               WHEN IPCity LIKE '%安徽%' THEN '安徽省'
               WHEN IPCity LIKE '%福建%' THEN '福建省'
               WHEN IPCity LIKE '%江西%' THEN '江西省'
               WHEN IPCity LIKE '%山东%' THEN '山东省'
               WHEN IPCity LIKE '%河南%' THEN '河南省'
               WHEN IPCity LIKE '%湖北%' THEN '湖北省'
               WHEN IPCity LIKE '%湖南%' THEN '湖南省'
               WHEN IPCity LIKE '%广东%' THEN '广东省'
               WHEN IPCity LIKE '%广西%' THEN '广西壮族自治区'
               WHEN IPCity LIKE '%海南%' THEN '海南省'
               WHEN IPCity LIKE '%四川%' THEN '四川省'
               WHEN IPCity LIKE '%贵州%' THEN '贵州省'
               WHEN IPCity LIKE '%云南%' THEN '云南省'
               WHEN IPCity LIKE '%西藏%' THEN '西藏自治区'
               WHEN IPCity LIKE '%陕西%' THEN '陕西省'
               WHEN IPCity LIKE '%甘肃%' THEN '甘肃省'
               WHEN IPCity LIKE '%青海%' THEN '青海省'
               WHEN IPCity LIKE '%宁夏%' THEN '宁夏回族自治区'
               WHEN IPCity LIKE '%新疆%' THEN '新疆维吾尔自治区'
               WHEN IPCity LIKE '%台湾%' THEN '台湾省'
               WHEN IPCity LIKE '%香港%' THEN '香港特别行政区'
               WHEN IPCity LIKE '%澳门%' THEN '澳门特别行政区'
               ELSE '其他'
           END AS province
    FROM db_links_Token WHERE Uid !=0
) AS subquery
GROUP BY province;
`).Rows()
	var Data = make([]gin.H, 0)
	if err != nil {
		return Data
	}
	defer rows.Close()

	// 将查询结果放入Data数组
	for rows.Next() {
		var count int
		var province string
		rows.Scan(&count, &province)
		Data = append(Data, gin.H{"name": province, "value": count})
	}

	if time.Now().Unix()-局_耗时 > 5 { //超过5秒的缓存
		global.H缓存.Set("图表数据_Get在线用户Ip地图分布统计", Data, time.Minute*10)
	}

	return Data
}

func Get在线用户统计(c *gin.Context) []gin.H {

	/*[
	{value: 1048, name: '测试应用1'},
	{value: 735, name: '测试应用2'},
	{value: 580, name: '测试应用3'},
	{value: 484, name: '测试应用4'},
	{value: 300, name: '测试应用5'},
	]*/
	if global.GVA_Viper.GetInt("系统模式") == 系统演示模式 {
		var Data = make([]gin.H, 5)
		Data[0] = gin.H{"name": "演示模式应用1", "value": 1048}
		Data[1] = gin.H{"name": "演示模式应用2", "value": 735}
		Data[2] = gin.H{"name": "演示模式应用3", "value": 580}
		Data[3] = gin.H{"name": "演示模式应用4", "value": 484}
		Data[4] = gin.H{"name": "演示模式应用5", "value": 300}
		return Data
	}

	Data缓存, ok := global.H缓存.Get("图表数据_Get在线用户统计")
	if ok {
		return Data缓存.([]gin.H)
	}

	局_耗时 := time.Now().Unix()
	var 局_appId列表 []int
	var 局_appId名称 = Ser_AppInfo.AppInfo取map列表Int(true)
	_ = global.GVA_DB.Model(DB.DB_LinksToken{}).Distinct("LoginAppid").Find(&局_appId列表).Error
	var Data = make([]gin.H, 0, len(局_appId列表))
	var 局_数量 int64
	for 索引, _ := range 局_appId列表 {
		局_数量 = 0
		global.GVA_DB.Model(DB.DB_LinksToken{}).Where("LoginAppid=?", 局_appId列表[索引]).Where("Status=1").Where("User!=?", "游客").Count(&局_数量)
		if 局_数量 > 0 {
			Data = append(Data, gin.H{"name": 局_appId名称[局_appId列表[索引]], "value": 局_数量})
		}

	}

	if time.Now().Unix()-局_耗时 > 5 { //超过5秒的缓存
		global.H缓存.Set("图表数据_Get在线用户统计", Data, time.Minute*10)
	}

	return Data
}

func Get在线用户统计登录活动时间(c *gin.Context) []gin.H {
	局_type := 结构_请求类型{Type: 1, AppId: 1}
	_ = c.ShouldBindJSON(&局_type)
	var 时间处理函数 func(int) string
	if 局_type.Type == 2 {
		时间处理函数 = 取相对时间0点时间戳月
	} else if 局_type.Type == 3 {
		时间处理函数 = 取相对时间0点时间戳时
	} else {
		时间处理函数 = 取相对时间0点时间戳天
	}

	var Data = make([]gin.H, 2)
	if global.GVA_Viper.GetInt("系统模式") == 系统演示模式 {
		//Data[0] = gin.H{"name": "登录统计", "type": "line", "data": []int{1, 2, 3, 4, 5, 6, 7}}
		//Data[1] = gin.H{"name": "活动统计", "type": "line", "data": []int{7, 6, 5, 4, 3, 2, 1}}
		return Data
	}

	Data缓存, ok := global.H缓存.Get("Get在线用户统计登录时间" + strconv.Itoa(局_type.Type))
	if ok {
		return Data缓存.([]gin.H)
	}

	局_耗时 := time.Now().Unix()
	var 局_临时 = make(map[string]interface{})

	局_sql := make([]string, func() int {
		if 局_type.Type == 3 {
			return 24
		}
		return 7
	}())

	for 索引 := range 局_sql {
		//fmt.Printf("负值开始%v,结束%v索引%v\n", -(len(局_sql) - 索引 - 1), -(len(局_sql) - 索引 - 2), 索引+1)
		//fmt.Printf("负值开始%v,结束%v索引%v\n", 时间处理函数(-(len(局_sql) - 索引 - 1)), 时间处理函数(-(len(局_sql) - 索引 - 2)), 索引+1)
		局_sql[索引] = fmt.Sprintf("Count(case when ( LoginTime between %s and %s ) then 1 else null end) as  '%d' ", 时间处理函数(-(len(局_sql) - 索引 - 1)), 时间处理函数(-(len(局_sql) - 索引 - 2)), 索引+1)
	}

	global.GVA_DB.Model(DB.DB_LinksToken{}).Select(局_sql).
		First(&局_临时)

	var 局_登录数量 = make([]int, len(局_临时))
	for 键名, 值 := range 局_临时 {
		索引, _ := strconv.Atoi(键名)
		if 值 == nil {
			局_登录数量[索引-1] = 0
		} else {
			a, _ := strconv.Atoi(string(值.([]uint8)))
			局_登录数量[索引-1] = a
		}
	}
	//fmt.Println(局_登录数量)
	Data[0] = gin.H{"name": "登录统计", "type": "line", "data": 局_登录数量}

	局_sql = make([]string, func() int {
		if 局_type.Type == 3 {
			return 24
		}
		return 7
	}())
	for 索引 := range 局_sql {
		//fmt.Printf("负值开始%v,结束%v索引%v\n", -(len(局_sql) - 索引 - 1), -(len(局_sql) - 索引 - 2), 索引+1)
		//fmt.Printf("负值开始%v,结束%v索引%v\n", 时间处理函数(-(len(局_sql) - 索引 - 1)), 时间处理函数(-(len(局_sql) - 索引 - 2)), 索引+1)
		局_sql[索引] = fmt.Sprintf("Count(case when ( LastTime between %s and %s ) then 1 else null end) as  '%d' ", 时间处理函数(-(len(局_sql) - 索引 - 1)), 时间处理函数(-(len(局_sql) - 索引 - 2)), 索引+1)
	}

	global.GVA_DB.Model(DB.DB_LinksToken{}).Select(局_sql).
		First(&局_临时)
	var 局_活动数量 = make([]int, len(局_临时)) //注意一定要新建一个变量局_活动数量,  make创建的变量这个会传指针
	for 键名, 值 := range 局_临时 {
		索引, _ := strconv.Atoi(键名)

		if 值 == nil {
			局_活动数量[索引-1] = 0
		} else {
			a, _ := strconv.Atoi(string(值.([]uint8)))
			局_活动数量[索引-1] = a
		}

	}
	//fmt.Println(局_活动数量)
	Data[1] = gin.H{"name": "活动统计", "type": "line", "data": 局_活动数量} //注意一定要新建一个变量局_活动数量,  make创建的变量这个会传指针

	if time.Now().Unix()-局_耗时 > 5 { //超过5秒的缓存
		global.H缓存.Set("Get在线用户统计登录时间"+strconv.Itoa(局_type.Type), Data, time.Minute*5)
	}

	return Data
}
func Get统计分时段在线总数(c *gin.Context) []gin.H {
	请求 := 结构_请求类型{Type: 1, AppId: 0}
	_ = c.ShouldBindJSON(&请求)
	var 局_开始时间戳 int64
	switch 请求.Type {
	case 1: //24小时
		局_开始时间戳 = time.Now().Unix() - 86400
	case 2: // 30天
		局_开始时间戳 = time.Now().Unix() - 86400*30
	case 3: //365天
		局_开始时间戳 = time.Now().Unix() - 86400*365
	}

	Data缓存, ok := global.H缓存.Get("Get统计分时段在线总数" + strconv.Itoa(请求.Type))
	if ok {
		return Data缓存.([]gin.H)
	}

	局_耗时 := time.Now().Unix()
	var 局_临时 = []dbm.DB_TongJiZaiXian{}

	tx := global.GVA_DB
	tx.Model(dbm.DB_TongJiZaiXian{}).Where("appId=?", 请求.AppId).Where("createdAt>?", 局_开始时间戳).Order("createdAt ASC").Find(&局_临时)

	var 局_登录数量 = make([]int64, 0, 32)
	var 局_登录时间 = make([]string, 0, 32)
	var 局_时间 int
	var S计数 int64
	for I, _ := range 局_临时 {
		//将时间戳转为时间 类型
		S时间 := time.Unix(局_临时[I].CreatedAt, 0)

		switch 请求.Type {
		case 1: //24小时
			if 局_时间 == 0 {
				局_时间 = S时间.Hour()
			}
			if S时间.Hour() == 局_时间 {
				S计数 += 局_临时[I].Count
			} else {
				局_登录数量 = append(局_登录数量, S计数)
				局_登录时间 = append(局_登录时间, strconv.Itoa(局_时间)+"时")
				局_时间 = S时间.Hour()
				S计数 = 局_临时[I].Count
			}
		case 2: // 30天
			if 局_时间 == 0 {
				局_时间 = S时间.Day()
			}
			if S时间.Day() == 局_时间 {
				S计数 += 局_临时[I].Count
			} else {
				局_登录数量 = append(局_登录数量, S计数)
				局_登录时间 = append(局_登录时间, strconv.Itoa(局_时间)+"天")
				局_时间 = S时间.Day()
				S计数 = 局_临时[I].Count
			}
		case 3: //365天
			if 局_时间 == 0 {
				局_时间 = int(S时间.Month())
			}
			if int(S时间.Month()) == 局_时间 {
				S计数 += 局_临时[I].Count
			} else {
				局_登录数量 = append(局_登录数量, S计数)
				局_登录时间 = append(局_登录时间, strconv.Itoa(局_时间)+"月")
				局_时间 = int(S时间.Month())
				S计数 = 局_临时[I].Count
			}

		}
		if len(局_临时) == I+1 {
			局_登录数量 = append(局_登录数量, S计数)
			switch 请求.Type {
			case 1:
				局_登录时间 = append(局_登录时间, strconv.Itoa(局_时间)+"时")
			case 2:
				局_登录时间 = append(局_登录时间, strconv.Itoa(局_时间)+"日")
			case 3:
				局_登录时间 = append(局_登录时间, strconv.Itoa(局_时间)+"月")
			}
		}

	}
	//fmt.Println(局_登录数量)
	Data := []gin.H{
		{"name": "统计分时段在线总数", "type": "line", "data": 局_登录数量},
		{"name": "统计分时段在线时间", "type": "line", "data": 局_登录时间},
	}

	if time.Now().Unix()-局_耗时 > 5 { //超过5秒的缓存
		global.H缓存.Set("Get统计分时段在线总数"+strconv.Itoa(请求.Type), Data, time.Minute*5)
	}

	return Data
}
func Get应用用户类型统计(c *gin.Context) []gin.H {
	局_type := 结构_请求类型{Type: 1, AppId: 10000}
	_ = c.ShouldBindJSON(&局_type)
	局_Appid := 局_type.AppId
	/*[
	{value: 1048, name: '测试应用1'},
	{value: 735, name: '测试应用2'},
	{value: 580, name: '测试应用3'},
	{value: 484, name: '测试应用4'},
	{value: 300, name: '测试应用5'},
	]*/
	if 局_Appid < 10000 || global.GVA_Viper.GetInt("系统模式") == 系统演示模式 {
		var Data = make([]gin.H, 5)
		Data[0] = gin.H{"name": "未分类1", "value": 330}
		Data[0] = gin.H{"name": "Vip1", "value": 1048}
		Data[1] = gin.H{"name": "Vip2", "value": 735}
		Data[2] = gin.H{"name": "Vip3", "value": 580}
		Data[3] = gin.H{"name": "Vip4", "value": 484}
		Data[4] = gin.H{"name": "Vip5", "value": 300}
		return Data
	}

	Data缓存, ok := global.H缓存.Get("图表数据_Get用户类型统计" + strconv.Itoa(局_Appid))
	if ok {
		return Data缓存.([]gin.H)
	}

	局_耗时 := time.Now().Unix()
	var 局_ClassId列表 []int
	var 局_名称 = Ser_UserClass.UserClass取map列表Int(局_Appid)
	局_名称[0] = "未分类"
	_ = global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_" + strconv.Itoa(局_Appid)).Distinct("UserClassId").Find(&局_ClassId列表).Error
	var Data = make([]gin.H, len(局_ClassId列表))
	var 局_数量 int64
	for 索引, _ := range 局_ClassId列表 {
		局_数量 = 0
		global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(局_Appid)).Where("UserClassId=?", 局_ClassId列表[索引]).Count(&局_数量)
		Data[索引] = gin.H{"name": 局_名称[局_ClassId列表[索引]], "value": 局_数量}
	}

	if time.Now().Unix()-局_耗时 > 5 { //超过5秒的缓存
		global.H缓存.Set("图表数据_Get用户类型统计"+strconv.Itoa(局_Appid), Data, time.Minute*10)
	}

	return Data
}
func Get应用用户统计(c *gin.Context) [][]string {

	/*[
	  ['product', '非会员', '会员', '总数'],
	  ['测试应用1', 43, 25, 999],
	  ['测试应用2', 23, 33, 999],
	  ['测试应用3', 36, 45, 999],
	  ['测试应用4', 4, 65, 999],
	  ['测试应用5', 86, 65, 999]
	]*/
	if global.GVA_Viper.GetInt("系统模式") == 系统演示模式 {
		var Data = [][]string{[]string{"product", "非会员", "会员", "总数"},
			[]string{"测试应用1", "43", "25", "999"},
			[]string{"测试应用2", "23", "33", "999"},
			[]string{"测试应用3", "36", "45", "999"},
			[]string{"测试应用4", "4", "65", "999"},
			[]string{"测试应用5", "86", "65", "999"},
		}

		return Data
	}

	Data缓存, ok := global.H缓存.Get("图表数据_Get应用用户统计")
	if ok {
		return Data缓存.([][]string)
	}
	局_耗时 := time.Now().Unix()
	var 局_appId列表 = Ser_AppInfo.AppInfo取map列表Int(false)
	var 局_appId用户数量 = make([]临时应用id总数键值对, len(局_appId列表))
	局_I := 0
	for 键名, _ := range 局_appId列表 {
		局_appId用户数量[局_I] = 临时应用id总数键值对{应用AppId: 键名, 总数: Ser_AppUser.Get用户总数(键名)}
		局_I++
	}
	// 用 age 排序，年龄相等的元素保持原始顺序
	sort.SliceStable(局_appId用户数量, func(i, j int) bool {
		return 局_appId用户数量[i].总数 < 局_appId用户数量[j].总数
	})

	//fmt.Println(局_appId用户数量) // [{David 2} {Eve 2} {Alice 23} {Bob 25}]

	//下面实现排序order by age asc, name desc，如果 age 和 name 都相等则保持原始排序
	sort.SliceStable(局_appId用户数量, func(i, j int) bool {
		if 局_appId用户数量[i].总数 != 局_appId用户数量[j].总数 {
			return 局_appId用户数量[i].总数 < 局_appId用户数量[j].总数
		}
		return false
	})

	fmt.Println(局_appId用户数量) // [{Eve 2} {David 2} {Alice 23} {Bob 25}]

	var Data = make([][]string, len(局_appId用户数量)+1)
	Data[0] = []string{"product", "非会员", "会员", "总数"}
	for 索引 := 0; 索引 < len(局_appId用户数量); 索引++ {
		局_会员, 局_非会员 := Ser_AppUser.Get用户会员和非会员数量(局_appId用户数量[索引].应用AppId)
		Data[索引+1] = []string{局_appId列表[局_appId用户数量[索引].应用AppId], strconv.FormatInt(局_非会员, 10), strconv.FormatInt(局_会员, 10), strconv.Itoa(局_appId用户数量[索引].总数)}
	}

	if time.Now().Unix()-局_耗时 > 5 { //超过5秒的缓存
		global.H缓存.Set("图表数据_Get应用用户统计", Data, time.Minute*10)
	}

	return Data
}
func Get卡号列表统计应用卡可用已用(c *gin.Context) [][]string {
	局_type := 结构_请求类型{Type: 1}
	_ = c.ShouldBindJSON(&局_type)
	/*[
	  ['product', '已用', '未用', '总数'],
	  ['测试应用1', 43, 25, 999],
	  ['测试应用2', 23, 33, 999],
	  ['测试应用3', 36, 45, 999],
	  ['测试应用4', 4, 65, 999],
	  ['测试应用5', 86, 65, 999]
	]*/
	if global.GVA_Viper.GetInt("系统模式") == 系统演示模式 {
		var Data = [][]string{[]string{"product", "已用", "未用", "总数"},
			[]string{"卡类1", "43", "25", "999"},
			[]string{"卡类2", "23", "33", "999"},
			[]string{"卡类3", "36", "45", "999"},
			[]string{"卡类4", "4", "65", "999"},
			[]string{"卡类5", "86", "65", "999"},
		}

		return Data
	}

	Data缓存, ok := global.H缓存.Get("Get卡号列表统计可用已用" + strconv.Itoa(局_type.Type))
	if ok {
		return Data缓存.([][]string)
	}
	局_耗时 := time.Now().Unix()
	var 局_appId列表 = Ser_AppInfo.AppInfo取map列表Int(false)

	var 局_appId卡号数量 = make([]临时应用id总数键值对, len(局_appId列表))
	局_I := 0
	for 键名, _ := range 局_appId列表 {
		局_appId卡号数量[局_I] = 临时应用id总数键值对{应用AppId: 键名, 总数: Ser_Ka.Get卡号总数(键名), 在线数量: int(Ser_LinkUser.Q指定应用真实在线(键名))}
		局_I++
	}
	// 用 age 排序，年龄相等的元素保持原始顺序
	sort.SliceStable(局_appId卡号数量, func(i, j int) bool {
		return 局_appId卡号数量[i].总数 < 局_appId卡号数量[j].总数
	})

	fmt.Println(局_appId卡号数量) // [{David 2} {Eve 2} {Alice 23} {Bob 25}]

	//下面实现排序order by age asc, name desc，如果 age 和 name 都相等则保持原始排序
	sort.SliceStable(局_appId卡号数量, func(i, j int) bool {
		if 局_appId卡号数量[i].总数 != 局_appId卡号数量[j].总数 {
			return 局_appId卡号数量[i].总数 < 局_appId卡号数量[j].总数
		}
		return false
	})

	fmt.Println(局_appId卡号数量) // [{Eve 2} {David 2} {Alice 23} {Bob 25}]

	var Data = make([][]string, len(局_appId卡号数量)+1)

	Data[0] = []string{"product", "已用", "未用", "总数", "在线"}
	for 索引 := 0; 索引 < len(局_appId卡号数量); 索引++ {
		局_已用, 局_未用 := Ser_Ka.Get应用已用和未用数量(局_appId卡号数量[索引].应用AppId)
		Data[索引+1] = []string{局_appId列表[局_appId卡号数量[索引].应用AppId], strconv.FormatInt(局_已用, 10), strconv.FormatInt(局_未用, 10), strconv.Itoa(局_appId卡号数量[索引].总数), strconv.Itoa(局_appId卡号数量[索引].在线数量)}
	}

	if time.Now().Unix()-局_耗时 > 5 { //超过5秒的缓存
		global.H缓存.Set("Get卡号列表统计可用已用"+strconv.Itoa(局_type.Type), Data, time.Minute*10)
	}

	return Data
}
func Get卡号列表统计应用卡类可用已用(c *gin.Context) [][]string {
	局_type := 结构_请求类型{Type: 1, AppId: 10000}
	_ = c.ShouldBindJSON(&局_type)
	/*[
	  ['product', '已用', '未用', '总数'],
	  ['测试应用1', 43, 25, 999],
	  ['测试应用2', 23, 33, 999],
	  ['测试应用3', 36, 45, 999],
	  ['测试应用4', 4, 65, 999],
	  ['测试应用5', 86, 65, 999]
	]*/
	if global.GVA_Viper.GetInt("系统模式") == 系统演示模式 {
		var Data = [][]string{[]string{"product", "已用", "未用", "总数"},
			[]string{"卡类1", "43", "25", "999"},
			[]string{"卡类2", "23", "33", "999"},
			[]string{"卡类3", "36", "45", "999"},
			[]string{"卡类4", "4", "65", "999"},
			[]string{"卡类5", "86", "65", "999"},
		}

		return Data
	}

	Data缓存, ok := global.H缓存.Get("Get卡号列表统计应用卡类可用已用" + strconv.Itoa(局_type.Type))
	if ok {
		return Data缓存.([][]string)
	}
	局_耗时 := time.Now().Unix()
	var 局_appId卡类列表 = Ser_KaClass.KaName取map列表Int(局_type.AppId)

	var 局_appId卡号数量 = make([]临时应用id总数键值对, len(局_appId卡类列表))
	局_I := 0
	for 键名, _ := range 局_appId卡类列表 {
		局_appId卡号数量[局_I] = 临时应用id总数键值对{应用AppId: 键名, 总数: Ser_Ka.Get卡类卡号总数(键名)}
		局_I++
	}
	// 用 age 排序，年龄相等的元素保持原始顺序
	sort.SliceStable(局_appId卡号数量, func(i, j int) bool {
		return 局_appId卡号数量[i].总数 < 局_appId卡号数量[j].总数
	})

	fmt.Println(局_appId卡号数量) // [{David 2} {Eve 2} {Alice 23} {Bob 25}]

	//下面实现排序order by age asc, name desc，如果 age 和 name 都相等则保持原始排序
	sort.SliceStable(局_appId卡号数量, func(i, j int) bool {
		if 局_appId卡号数量[i].总数 != 局_appId卡号数量[j].总数 {
			return 局_appId卡号数量[i].总数 < 局_appId卡号数量[j].总数
		}
		return false
	})

	fmt.Println(局_appId卡号数量) // [{Eve 2} {David 2} {Alice 23} {Bob 25}]

	var Data = make([][]string, len(局_appId卡号数量)+1)
	Data[0] = []string{"product", "已用", "未用", "总数"}
	for 索引 := 0; 索引 < len(局_appId卡号数量); 索引++ {
		局_已用, 局_未用 := Ser_Ka.Get卡类已用和未用数量(局_appId卡号数量[索引].应用AppId)
		Data[索引+1] = []string{局_appId卡类列表[局_appId卡号数量[索引].应用AppId], strconv.FormatInt(局_已用, 10), strconv.FormatInt(局_未用, 10), strconv.Itoa(局_appId卡号数量[索引].总数)}
	}

	if time.Now().Unix()-局_耗时 > 5 { //超过5秒的缓存
		global.H缓存.Set("Get卡号列表统计应用卡类可用已用"+strconv.Itoa(局_type.Type), Data, time.Minute*10)
	}

	return Data
}

type 临时应用id总数键值对 struct {
	应用AppId int
	总数      int
	在线数量    int
}
type 结构_请求类型 struct {
	Type  int `json:"Type"`
	AppId int `json:"AppId"`
}

func Get余额充值消费统计(c *gin.Context) []gin.H {
	局_type := 结构_请求类型{Type: 1}
	_ = c.ShouldBindJSON(&局_type)
	var 时间处理函数 func(int) string
	if 局_type.Type == 2 {
		时间处理函数 = 取相对时间0点时间戳月
	} else {
		时间处理函数 = 取相对时间0点时间戳天
	}

	/*[
	  {
	    name: '充值金额',
	    type: 'line',
	    data: [320, 332, 341, 354, 390, 220, 450]
	  }, {
	    name: '消费金额',
	    type: 'line',
	    data: [120, 132, 101, 134, 90, 130, 210]
	  }
	]*/
	var Data = make([]gin.H, 2)
	if global.GVA_Viper.GetInt("系统模式") == 系统演示模式 {
		Data[0] = gin.H{"name": "充值金额", "type": "line", "data": []int{320, 332, 341, 354, 390, 220, 450}}
		Data[1] = gin.H{"name": "消费金额", "type": "line", "data": []int{120, 132, 101, 134, 90, 130, 210}}
		return Data
	}

	Data缓存, ok := global.H缓存.Get("图表数据_Get余额充值消费统计" + strconv.Itoa(局_type.Type))
	if ok {
		return Data缓存.([]gin.H)
	}

	局_耗时 := time.Now().Unix()
	var 局_临时 = make(map[string]interface{})

	/*     select
	  Count(case when (Time between 1684900520 and 1684901520) then 1 else null end) as  "昨天",
	Count(case when (Time between 1684901520 and 1684902520) then 1 else null end) as  "前天"
	 from db_log_login

	*/
	var 局_数量 [7]string

	global.GVA_DB.Model(DB.DB_LogRMBPayOrder{}).
		Select("SUM(case when ( Time between "+时间处理函数(-6)+" and "+时间处理函数(-5)+") then Rmb else null end) as  '1' ",
			"SUM(case when ( Time between "+时间处理函数(-5)+" and "+时间处理函数(-4)+") then Rmb else null end) as  '2' ",
			"SUM(case when ( Time between "+时间处理函数(-4)+" and "+时间处理函数(-3)+") then Rmb else null end) as  '3' ",
			"SUM(case when ( Time between "+时间处理函数(-3)+" and "+时间处理函数(-2)+") then Rmb else null end) as  '4' ",
			"SUM(case when ( Time between "+时间处理函数(-2)+" and "+时间处理函数(-1)+") then Rmb else null end) as  '5' ",
			"SUM(case when ( Time between "+时间处理函数(-1)+" and "+时间处理函数(0)+") then Rmb else null end) as  '6' ",
			"SUM(case when ( Time between "+时间处理函数(0)+" and "+时间处理函数(1)+") then Rmb else null end) as  '7' ").
		Order("").Where("Status=3").
		First(&局_临时)

	for 键名, 值 := range 局_临时 {
		索引, _ := strconv.Atoi(键名)
		if 值 == nil {
			局_数量[索引-1] = "0"
		} else {
			a := string(值.([]uint8))
			局_数量[索引-1] = a
		}
	}
	Data[0] = gin.H{"name": "充值金额", "type": "line", "data": 局_数量}

	global.GVA_DB.Model(DB.DB_LogMoney{}).
		Select("Count(case when ( Time between "+时间处理函数(-6)+" and "+时间处理函数(-5)+") then Count else null end) as  '1' ",
			"SUM(case when ( Time between "+时间处理函数(-5)+" and "+时间处理函数(-4)+") then Count else null end) as  '2' ",
			"SUM(case when ( Time between "+时间处理函数(-4)+" and "+时间处理函数(-3)+") then Count else null end) as  '3' ",
			"SUM(case when ( Time between "+时间处理函数(-3)+" and "+时间处理函数(-2)+") then Count else null end) as  '4' ",
			"SUM(case when ( Time between "+时间处理函数(-2)+" and "+时间处理函数(-1)+") then Count else null end) as  '5' ",
			"SUM(case when ( Time between "+时间处理函数(-1)+" and "+时间处理函数(0)+") then Count else null end) as  '6' ",
			"SUM(case when ( Time between "+时间处理函数(0)+" and "+时间处理函数(1)+") then Count else null end) as  '7' ").
		Order("").Where("Count<0").
		First(&局_临时)

	for 键名, 值 := range 局_临时 {
		索引, _ := strconv.Atoi(键名)
		if 值 == nil {
			局_数量[索引-1] = "0"
		} else {
			a := string(值.([]uint8))
			局_数量[索引-1] = strings.Replace(a, "-", "", 1) //把负号替换掉

		}
	}
	Data[1] = gin.H{"name": "消费金额", "type": "line", "data": 局_数量}

	if time.Now().Unix()-局_耗时 > 5 { //超过5秒的缓存
		global.H缓存.Set("图表数据_Get余额充值消费统计"+strconv.Itoa(局_type.Type), Data, time.Minute*10)
	}

	return Data
}
func Get积分点数消费统计(c *gin.Context) []gin.H {
	局_type := 结构_请求类型{Type: 1}
	_ = c.ShouldBindJSON(&局_type)
	var 时间处理函数 func(int) string
	if 局_type.Type == 2 {
		时间处理函数 = 取相对时间0点时间戳月
	} else {
		时间处理函数 = 取相对时间0点时间戳天
	}

	/*[
	  {
	    name: '消费点数',
	    type: 'line',
	    data: [320, 332, 341, 354, 390, 220, 450]
	  }, {
	    name: '消费积分',
	    type: 'line',
	    data: [120, 132, 101, 134, 90, 130, 210]
	  }
	]*/
	var Data = make([]gin.H, 2)

	if global.GVA_Viper.GetInt("系统模式") == 系统演示模式 {
		Data[0] = gin.H{"name": "消费点数", "type": "line", "data": []int{320, 332, 341, 354, 390, 220, 450}}
		Data[1] = gin.H{"name": "消费积分", "type": "line", "data": []int{120, 132, 101, 134, 90, 130, 210}}
		return Data
	}

	Data缓存, ok := global.H缓存.Get("图表数据_Get积分点数消费统计" + strconv.Itoa(局_type.Type))
	if ok {
		return Data缓存.([]gin.H)
	}

	局_耗时 := time.Now().Unix()
	var 局_临时 = make(map[string]interface{})

	/*     select
	  Count(case when (Time between 1684900520 and 1684901520) then 1 else null end) as  "昨天",
	Count(case when (Time between 1684901520 and 1684902520) then 1 else null end) as  "前天"
	 from db_log_login

	*/
	var 局_数量 [7]string

	global.GVA_DB.Model(DB.DB_LogVipNumber{}).
		Select("SUM(case when ( Time between "+时间处理函数(-6)+" and "+时间处理函数(-5)+") then Count else null end) as  '1' ",
			"SUM(case when ( Time between "+时间处理函数(-5)+" and "+时间处理函数(-4)+") then Count else null end) as  '2' ",
			"SUM(case when ( Time between "+时间处理函数(-4)+" and "+时间处理函数(-3)+") then Count else null end) as  '3' ",
			"SUM(case when ( Time between "+时间处理函数(-3)+" and "+时间处理函数(-2)+") then Count else null end) as  '4' ",
			"SUM(case when ( Time between "+时间处理函数(-2)+" and "+时间处理函数(-0)+") then Count else null end) as  '5' ",
			"SUM(case when ( Time between "+时间处理函数(-1)+" and "+时间处理函数(0)+") then Count else null end) as  '6' ",
			"SUM(case when ( Time between "+时间处理函数(0)+" and "+时间处理函数(1)+") then Count else null end) as  '7' ").
		Order("").Where("Type=1").
		First(&局_临时)

	for 键名, 值 := range 局_临时 {
		索引, _ := strconv.Atoi(键名)
		if 值 == nil {
			局_数量[索引-1] = "0"
		} else {
			a := string(值.([]uint8))
			局_数量[索引-1] = a
			局_数量[索引-1] = strings.Replace(a, "-", "", 1) //把负号替换掉
		}
	}
	Data[0] = gin.H{"name": "消费积分", "type": "line", "data": 局_数量}

	global.GVA_DB.Model(DB.DB_LogVipNumber{}).
		Select("Count(case when ( Time between "+时间处理函数(-6)+" and "+时间处理函数(-5)+") then Count else null end) as  '1' ",
			"SUM(case when ( Time between "+时间处理函数(-5)+" and "+时间处理函数(-4)+") then Count else null end) as  '2' ",
			"SUM(case when ( Time between "+时间处理函数(-4)+" and "+时间处理函数(-3)+") then Count else null end) as  '3' ",
			"SUM(case when ( Time between "+时间处理函数(-3)+" and "+时间处理函数(-2)+") then Count else null end) as  '4' ",
			"SUM(case when ( Time between "+时间处理函数(-2)+" and "+时间处理函数(-1)+") then Count else null end) as  '5' ",
			"SUM(case when ( Time between "+时间处理函数(-1)+" and "+时间处理函数(0)+") then Count else null end) as  '6' ",
			"SUM(case when ( Time between "+时间处理函数(0)+" and "+时间处理函数(1)+") then Count else null end) as  '7' ").
		Order("").Where("Type=2").
		First(&局_临时)

	for 键名, 值 := range 局_临时 {
		索引, _ := strconv.Atoi(键名)
		if 值 == nil {
			局_数量[索引-1] = "0"
		} else {
			a := string(值.([]uint8))
			局_数量[索引-1] = strings.Replace(a, "-", "", 1) //把负号替换掉

		}
	}
	Data[1] = gin.H{"name": "消费点数", "type": "line", "data": 局_数量}

	if time.Now().Unix()-局_耗时 > 5 { //超过5秒的缓存
		global.H缓存.Set("图表数据_Get积分点数消费统计"+strconv.Itoa(局_type.Type), Data, time.Minute*10)
	}

	return Data
}
func Get卡号列表统计制卡(c *gin.Context) []gin.H {
	局_type := 结构_请求类型{Type: 1}
	_ = c.ShouldBindJSON(&局_type)
	var 时间处理函数 func(int) string
	if 局_type.Type == 2 {
		时间处理函数 = 取相对时间0点时间戳月
	} else {
		时间处理函数 = 取相对时间0点时间戳天
	}

	/*[
	  {
	    name: '制卡数量',
	    type: 'line',
	    data: [320, 332, 341, 354, 390, 430, 450]
	  }, {
	    name: '使用数量',
	    type: 'line',
	    data: [120, 132, 101, 134, 90, 230, 210]
	  }
	]*/
	var Data = make([]gin.H, 2)
	if global.GVA_Viper.GetInt("系统模式") == 系统演示模式 {
		Data[0] = gin.H{"name": "制卡数量", "type": "line", "data": []int{320, 332, 341, 354, 390, 220, 450}}
		return Data
	}

	Data缓存, ok := global.H缓存.Get("图表数据_Get卡号制卡使用统计" + strconv.Itoa(局_type.Type))
	if ok {
		return Data缓存.([]gin.H)
	}

	局_耗时 := time.Now().Unix()
	var 局_临时 = make(map[string]interface{})

	var 局_数量 [7]int
	局_db := global.GVA_DB.Model(DB.DB_Ka{})
	局_db.Select("SUM(case when (  RegisterTime between "+时间处理函数(-6)+" and "+时间处理函数(-5)+") then 1 else null end) as  '1' ",
		"SUM(case when (  RegisterTime between "+时间处理函数(-5)+" and "+时间处理函数(-4)+") then 1 else null end) as  '2' ",
		"SUM(case when (  RegisterTime between "+时间处理函数(-4)+" and "+时间处理函数(-3)+") then 1 else null end) as  '3' ",
		"SUM(case when (  RegisterTime between "+时间处理函数(-3)+" and "+时间处理函数(-2)+") then 1 else null end) as  '4' ",
		"SUM(case when (  RegisterTime between "+时间处理函数(-2)+" and "+时间处理函数(-0)+") then 1 else null end) as  '5' ",
		"SUM(case when (  RegisterTime between "+时间处理函数(-1)+" and "+时间处理函数(0)+") then 1 else null end) as  '6' ",
		"SUM(case when (  RegisterTime between "+时间处理函数(0)+" and "+时间处理函数(1)+") then 1 else null end) as  '7' ").
		Order("").
		First(&局_临时)

	for 键名, 值 := range 局_临时 {
		索引, _ := strconv.Atoi(键名)
		if 值 == nil {
			局_数量[索引-1] = 0
		} else {
			a, _ := strconv.Atoi(string(值.([]uint8)))
			局_数量[索引-1] = a
		}
	}
	Data[0] = gin.H{"name": "制卡数量", "type": "line", "data": 局_数量}

	if time.Now().Unix()-局_耗时 > 5 { //超过5秒的缓存
		global.H缓存.Set("图表数据_Get卡号制卡统计"+strconv.Itoa(局_type.Type), Data, time.Minute*10)
	}

	return Data
}
func Get卡号列表统计制卡_代理(c *gin.Context) []gin.H {
	局_type := 结构_请求类型{Type: 1}
	_ = c.ShouldBindJSON(&局_type)
	var 时间处理函数 func(int) string
	if 局_type.Type == 2 {
		时间处理函数 = 取相对时间0点时间戳月
	} else {
		时间处理函数 = 取相对时间0点时间戳天
	}

	/*[
	  {
	    name: '制卡数量',
	    type: 'line',
	    data: [320, 332, 341, 354, 390, 430, 450]
	  }, {
	    name: '使用数量',
	    type: 'line',
	    data: [120, 132, 101, 134, 90, 230, 210]
	  }
	]*/
	var Data = make([]gin.H, 2)

	Data缓存, ok := global.H缓存.Get("图表数据_Get代理" + c.GetString("User") + "卡号制卡使用统计" + strconv.Itoa(局_type.Type))
	if ok {
		return Data缓存.([]gin.H)
	}

	局_耗时 := time.Now().Unix()
	var 局_临时 = make(map[string]interface{})

	var 局_数量 [7]int
	局_db := global.GVA_DB.Model(DB.DB_Ka{}).Where("RegisterUser = ?", c.GetString("User"))
	局_db.Select("SUM(case when (  RegisterTime between "+时间处理函数(-6)+" and "+时间处理函数(-5)+") then 1 else null end) as  '1' ",
		"SUM(case when (  RegisterTime between "+时间处理函数(-5)+" and "+时间处理函数(-4)+") then 1 else null end) as  '2' ",
		"SUM(case when (  RegisterTime between "+时间处理函数(-4)+" and "+时间处理函数(-3)+") then 1 else null end) as  '3' ",
		"SUM(case when (  RegisterTime between "+时间处理函数(-3)+" and "+时间处理函数(-2)+") then 1 else null end) as  '4' ",
		"SUM(case when (  RegisterTime between "+时间处理函数(-2)+" and "+时间处理函数(-0)+") then 1 else null end) as  '5' ",
		"SUM(case when (  RegisterTime between "+时间处理函数(-1)+" and "+时间处理函数(0)+") then 1 else null end) as  '6' ",
		"SUM(case when (  RegisterTime between "+时间处理函数(0)+" and "+时间处理函数(1)+") then 1 else null end) as  '7' ").
		Order("").
		First(&局_临时)

	for 键名, 值 := range 局_临时 {
		索引, _ := strconv.Atoi(键名)
		if 值 == nil {
			局_数量[索引-1] = 0
		} else {
			a, _ := strconv.Atoi(string(值.([]uint8)))
			局_数量[索引-1] = a
		}
	}
	Data[0] = gin.H{"name": "制卡数量", "type": "line", "data": 局_数量}

	if time.Now().Unix()-局_耗时 > 5 { //超过5秒的缓存
		global.H缓存.Set("图表数据_Get代理"+c.GetString("User")+"卡号制卡使用统计"+strconv.Itoa(局_type.Type), Data, time.Minute*10)
	}

	return Data
}
func 取相对时间0点时间戳时(时增减 int) string {
	ts := time.Now()
	timeStampYesterday := time.Date(ts.Year(), ts.Month(), ts.Day(), ts.Hour(), 0, 0, 0, ts.Location()).Unix()
	局_最终时间 := int(timeStampYesterday) + (时增减 * 3600)
	return strconv.Itoa(局_最终时间)
}
func 取相对时间0点时间戳天(天数增减 int) string {
	ts := time.Now().AddDate(0, 0, 天数增减)
	timeStampYesterday := time.Date(ts.Year(), ts.Month(), ts.Day(), 0, 0, 0, 0, ts.Location()).Unix()
	return strconv.Itoa(int(timeStampYesterday))
}
func 取相对时间0点时间戳月(月增减 int) string {
	ts := time.Now().AddDate(0, 月增减, 0)
	// 使用1号而不是0号来获取当月的第一天
	timeStampFirstDay := time.Date(ts.Year(), ts.Month(), 1, 0, 0, 0, 0, ts.Location()).Unix()
	return strconv.Itoa(int(timeStampFirstDay))
}

func Get应用用户账号注册统计(c *gin.Context) []gin.H {
	局_type := 结构_请求类型{Type: 1, AppId: 1}
	_ = c.ShouldBindJSON(&局_type)
	var 时间处理函数 func(int) string
	if 局_type.Type == 2 {
		时间处理函数 = 取相对时间0点时间戳月
	} else {
		时间处理函数 = 取相对时间0点时间戳天
	}

	/*[{
	    name: '注册数量',
	    type: 'line',
	    data: [120, 132, 101, 134, 90, 230, 210]
	  },
	  {
	    name: '登录数量',
	    type: 'line',
	    data: [220, 182, 191, 234, 290, 330, 310]
	  }
	]*/
	var Data = make([]gin.H, 1)
	if global.GVA_Viper.GetInt("系统模式") == 系统演示模式 {
		Data[0] = gin.H{"name": "注册数量", "type": "line", "data": []int{120, 132, 101, 134, 90, 230, 210}}
		return Data
	}

	Data缓存, ok := global.H缓存.Get("图表数据_Get用户账号统计" + strconv.Itoa(局_type.Type) + "_" + strconv.Itoa(局_type.AppId))
	if ok {
		return Data缓存.([]gin.H)
	}

	局_耗时 := time.Now().Unix()
	var 局_临时 = make(map[string]interface{})

	var 局_数量 [7]int
	//
	global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(局_type.AppId)).
		Select("Count(case when ( RegisterTime between "+时间处理函数(-6)+" and "+时间处理函数(-5)+") then 1 else null end) as  '1' ",
			"Count(case when ( RegisterTime between "+时间处理函数(-5)+" and "+时间处理函数(-4)+") then 1 else null end) as  '2' ",
			"Count(case when ( RegisterTime between "+时间处理函数(-4)+" and "+时间处理函数(-3)+") then 1 else null end) as  '3' ",
			"Count(case when ( RegisterTime between "+时间处理函数(-3)+" and "+时间处理函数(-2)+") then 1 else null end) as  '4' ",
			"Count(case when ( RegisterTime between "+时间处理函数(-2)+" and "+时间处理函数(-1)+") then 1 else null end) as  '5' ",
			"Count(case when ( RegisterTime between "+时间处理函数(-1)+" and "+时间处理函数(0)+") then 1 else null end) as  '6' ",
			"Count(case when ( RegisterTime between "+时间处理函数(0)+" and "+时间处理函数(1)+") then 1 else null end) as  '7' ").
		First(&局_临时)

	for 键名, 值 := range 局_临时 {
		索引, _ := strconv.Atoi(键名)
		if 值 == nil {
			局_数量[索引-1] = 0
		} else {
			a, _ := strconv.Atoi(string(值.([]uint8)))
			局_数量[索引-1] = a
		}
	}
	Data[0] = gin.H{"name": "注册数量", "type": "line", "data": 局_数量}

	//统计日活========================

	if time.Now().Unix()-局_耗时 > 5 { //超过5秒的缓存
		global.H缓存.Set("图表数据_Get用户账号统计"+strconv.Itoa(局_type.Type)+"_"+strconv.Itoa(局_type.AppId), Data, time.Minute*10)
	}

	return Data
}
func Get用户账号登录注册统计(c *gin.Context) []gin.H {
	局_type := 结构_请求类型{Type: 1, AppId: 1}
	_ = c.ShouldBindJSON(&局_type)
	var 时间处理函数 func(int) string
	if 局_type.Type == 2 {
		时间处理函数 = 取相对时间0点时间戳月
	} else {
		时间处理函数 = 取相对时间0点时间戳天
	}

	/*[{
	    name: '注册数量',
	    type: 'line',
	    data: [120, 132, 101, 134, 90, 230, 210]
	  },
	  {
	    name: '登录数量',
	    type: 'line',
	    data: [220, 182, 191, 234, 290, 330, 310]
	  }
	]*/
	var Data = make([]gin.H, 2)
	if global.GVA_Viper.GetInt("系统模式") == 系统演示模式 {
		Data[0] = gin.H{"name": "注册数量", "type": "line", "data": []int{120, 132, 101, 134, 90, 230, 210}}
		Data[1] = gin.H{"name": "登录数量", "type": "line", "data": []int{220, 182, 191, 234, 290, 330, 310}}
		return Data
	}

	Data缓存, ok := global.H缓存.Get("图表数据_Get用户账号统计" + strconv.Itoa(局_type.Type))
	if ok {
		return Data缓存.([]gin.H)
	}

	局_耗时 := time.Now().Unix()
	var 局_临时 = make(map[string]interface{})

	/*     select
	  Count(case when (Time between 1684900520 and 1684901520) then 1 else null end) as  "昨天",
	Count(case when (Time between 1684901520 and 1684902520) then 1 else null end) as  "前天"
	 from db_log_login

	*/
	var 局_数量 [7]int
	/*global.GVA_DB.Model(DB.DB_User{}).
		Select("Count(case when ( LoginTime< "+取相对时间0点时间戳天(-7)+") then 1 else null end) as  '1' ",
			"Count(case when ( LoginTime<  "+取相对时间0点时间戳天(-6)+") then 1 else null end) as  '2' ",
			"Count(case when ( LoginTime< "+取相对时间0点时间戳天(-5)+") then 1 else null end) as  '3' ",
			"Count(case when ( LoginTime<  "+取相对时间0点时间戳天(-4)+") then 1 else null end) as  '4' ",
			"Count(case when ( LoginTime<  "+取相对时间0点时间戳天(-3)+") then 1 else null end) as  '5' ",
			"Count(case when ( LoginTime<  "+取相对时间0点时间戳天(-2)+") then 1 else null end) as  '6' ",
			"Count(case when ( LoginTime< "+取相对时间0点时间戳天(-1)+") then 1 else null end) as  '7' ").
		Order("").
		First(&局_临时)

	for 键名, 值 := range 局_临时 {
		索引, _ := strconv.Atoi(键名)
		a, _ := strconv.Atoi(string(值.([]uint8)))
		局_数量[索引-1] = a
	}
	Data[0] = gin.H{"name": "用户总数", "type": "line", "data": 局_数量}*/

	global.GVA_DB.Model(DB.DB_User{}).
		Select("Count(case when ( RegisterTime between "+时间处理函数(-6)+" and "+时间处理函数(-5)+") then 1 else null end) as  '1' ",
			"Count(case when ( RegisterTime between "+时间处理函数(-5)+" and "+时间处理函数(-4)+") then 1 else null end) as  '2' ",
			"Count(case when ( RegisterTime between "+时间处理函数(-4)+" and "+时间处理函数(-3)+") then 1 else null end) as  '3' ",
			"Count(case when ( RegisterTime between "+时间处理函数(-3)+" and "+时间处理函数(-2)+") then 1 else null end) as  '4' ",
			"Count(case when ( RegisterTime between "+时间处理函数(-2)+" and "+时间处理函数(-1)+") then 1 else null end) as  '5' ",
			"Count(case when ( RegisterTime between "+时间处理函数(-1)+" and "+时间处理函数(0)+") then 1 else null end) as  '6' ",
			"Count(case when ( RegisterTime between "+时间处理函数(0)+" and "+时间处理函数(1)+") then 1 else null end) as  '7' ").
		First(&局_临时)
	for 键名, 值 := range 局_临时 {
		索引, _ := strconv.Atoi(键名)
		if 值 == nil {
			局_数量[索引-1] = 0
		} else {
			a, _ := strconv.Atoi(utils.D到文本(值))
			局_数量[索引-1] = a
		}
	}
	Data[0] = gin.H{"name": "注册数量", "type": "line", "data": 局_数量}

	//mark 这个获取算法有bug,这个是通过用户列表最后一次登录时间统计的数据,但是实际如果用户连续两天登录,前一天登录的数据就没了,所以不可使用
	/*	global.GVA_DB.Model(DB.DB_User{}).
		Select("Count(case when ( LoginTime between "+时间处理函数(-6)+" and "+时间处理函数(-5)+") then 1 else null end) as  '1' ",
			"Count(case when ( LoginTime between "+时间处理函数(-5)+" and "+时间处理函数(-4)+") then 1 else null end) as  '2' ",
			"Count(case when ( LoginTime between "+时间处理函数(-4)+" and "+时间处理函数(-3)+") then 1 else null end) as  '3' ",
			"Count(case when ( LoginTime between "+时间处理函数(-3)+" and "+时间处理函数(-2)+") then 1 else null end) as  '4' ",
			"Count(case when ( LoginTime between "+时间处理函数(-2)+" and "+时间处理函数(-1)+") then 1 else null end) as  '5' ",
			"Count(case when ( LoginTime between "+时间处理函数(-1)+" and "+时间处理函数(0)+") then 1 else null end) as  '6' ",
			"Count(case when ( LoginTime between "+时间处理函数(0)+" and "+时间处理函数(1)+") then 1 else null end) as  '7' ").
		First(&局_临时)*/
	//老老实实读取登录日志吧
	global.GVA_DB.Model(DB.DB_LogLogin{}).
		Select("Count(case when ( Time between "+时间处理函数(-6)+" and "+时间处理函数(-5)+") then 1 else null end) as  '1' ",
			"Count(case when ( Time between "+时间处理函数(-5)+" and "+时间处理函数(-4)+") then 1 else null end) as  '2' ",
			"Count(case when ( Time between "+时间处理函数(-4)+" and "+时间处理函数(-3)+") then 1 else null end) as  '3' ",
			"Count(case when ( Time between "+时间处理函数(-3)+" and "+时间处理函数(-2)+") then 1 else null end) as  '4' ",
			"Count(case when ( Time between "+时间处理函数(-2)+" and "+时间处理函数(-1)+") then 1 else null end) as  '5' ",
			"Count(case when ( Time between "+时间处理函数(-1)+" and "+时间处理函数(0)+") then 1 else null end) as  '6' ",
			"Count(case when ( Time between "+时间处理函数(0)+" and "+时间处理函数(1)+") then 1 else null end) as  '7' ").
		Where("Note = ?", "用户登录").First(&局_临时)
	for 键名, 值 := range 局_临时 {
		索引, _ := strconv.Atoi(键名)

		if 值 == nil {
			局_数量[索引-1] = 0
		} else {
			a, _ := strconv.Atoi(utils.D到文本(值))
			局_数量[索引-1] = a
		}

	}
	Data[1] = gin.H{"name": "登录数量", "type": "line", "data": 局_数量}

	if time.Now().Unix()-局_耗时 > 5 { //超过5秒的缓存
		global.H缓存.Set("图表数据_Get用户账号统计"+strconv.Itoa(局_type.Type), Data, time.Minute*10)
	}

	return Data
}

func Get代理组织架构图(c *gin.Context, 根代理ID int) []*Node {

	/*{
	  "id": -1, "label": "管理员",
	  "style": {"color": "#ffffff", "background": "#108ffe"},
	  "children": [
	    {
	      "id": 2, "pid": -1, "label": "刘备",
	      "style": {"color": "#000000", "background": "#79bbff"},
	      "children": [
	        {
	          "id": 3, "pid": 2, "label": "关羽",   "style": {"color": "#000000", "background": " #c6e2ff"}, "children": [
	            {"id": 6, "pid": 3, "label": "关平"},
	          ]
	        },
	        {
	          "id": 4, "pid": 2, "label": "张飞",   "style": {"color": "#000000", "background": " #c6e2ff"}, "children": [
	            {"id": 7, "pid": 4, "label": "张苞"}]
	        },
	        {
	          "id": 5, "pid": 2, "label": "诸葛亮",   "style": {"color": "#000000", "background": " #c6e2ff"},
	        }

	      ]
	    }
	  ]
	}*/

	/*	Data缓存, ok := global.H缓存.Get("图表数据_Get代理组织架构图")
		if ok {
			return Data缓存.(*Node)
		}
	*/
	//局_耗时 := time.Now().Unix()
	var 局_用户数组 []DB.DB_User

	_ = global.GVA_DB.Model(DB.DB_User{}).Select("Id", "User", "UPAgentId", "AgentDiscount").Where("UPAgentId !=0").Find(&局_用户数组).Error
	if len(局_用户数组) == 0 { //防止无代理会报错
		return []*Node{}
	}
	nodes := make([]*Node, 0, len(局_用户数组))
	for 索引, _ := range 局_用户数组 {
		nodes = append(nodes, &Node{
			Id:            局_用户数组[索引].Id,
			UPAgentId:     局_用户数组[索引].UPAgentId,
			User:          局_用户数组[索引].User,
			AgentDiscount: 局_用户数组[索引].AgentDiscount,
		})
	}
	// 构建节点数据
	/*	nodes := []*Node{
			{Id: -1, User: "管理员"},
			{Id: 2, UPAgentId: -1, User: "刘备"},
			{Id: 3, UPAgentId: 2, User: "关羽"},
			{Id: 4, UPAgentId: 2, User: "张飞"},
			{Id: 5, UPAgentId: 2, User: "诸葛亮"},
			{Id: 6, UPAgentId: 3, User: "关平"},
			{Id: 7, UPAgentId: 4, User: "张苞"},
		}
	*/
	// 构建树形结构

	Data := getTreeIterative(nodes, 根代理ID)

	/*
		if time.Now().Unix()-局_耗时 > 5 { //超过5秒的缓存
			global.H缓存.Set("图表数据_Get在线用户统计", Data, time.Minute*10)
		}*/

	return Data
}

type Node struct {
	Id            int     `json:"Id" gorm:"column:Id;primarykey;AUTO_INCREMENT"` // id
	User          string  `json:"User" gorm:"column:User;size:191;UNIQUE;index;comment:用户登录名"`
	UPAgentId     int     `json:"UPAgentId" gorm:"column:UPAgentId;comment:上级代理id"`
	AgentDiscount int     `json:"AgentDiscount" gorm:"column:AgentDiscount;comment:分成百分比"`
	Children      []*Node `json:"Children,omitempty" gorm:"column:Children;comment:下级代理id"`
}

func getTreeIterative(list []*Node, parentId int) []*Node {
	memo := make(map[int]*Node)
	for _, v := range list {
		if _, ok := memo[v.Id]; ok {
			v.Children = memo[v.Id].Children
			memo[v.Id] = v
		} else {
			v.Children = make([]*Node, 0)
			memo[v.Id] = v
		}
		if _, ok := memo[v.UPAgentId]; ok {
			memo[v.UPAgentId].Children = append(memo[v.UPAgentId].Children, memo[v.Id])
		} else {
			memo[v.UPAgentId] = &Node{Children: []*Node{memo[v.Id]}}
		}
	}
	return memo[parentId].Children

}

// 定义日期统计结构
type dailyStat struct {
	success int
	fail    int
}

// 获取统计结果
func Get任务池任务Id分析(c *gin.Context) [][]string {
	局_type := struct {
		TaskId int `json:"TaskId"`
	}{}
	_ = c.ShouldBindJSON(&局_type)

	var TaskDataList []DB.DB_TaskPoolData
	global.GVA_DB.Model(DB.DB_TaskPoolData{}).Select("TimeStart,Status").
		Where("Tid = ?", 局_type.TaskId).
		Find(&TaskDataList)

	// 初始化30天数据容器（从今天往前推29天）
	now := time.Now()
	dayStats := make(map[string]*dailyStat)
	for i := 0; i < 30; i++ {
		date := now.AddDate(0, 0, -i).Format("01-02")
		dayStats[date] = &dailyStat{}
	}

	// 统计每日数据
	for _, task := range TaskDataList {
		taskDate := time.Unix(int64(task.TimeStart), 0).Format("01-02")
		if stat, exists := dayStats[taskDate]; exists {
			if task.Status == 3 { // 3表示成功
				stat.success++
			} else if task.Status == 4 { // 4表示失败
				stat.fail++
			}
		}
	}

	// 构建结果数组（按时间倒序：最近→最早）
	Data := [][]string{{"日期", "失败", "成功", "总数"}}
	weekdays := []string{"日", "一", "二", "三", "四", "五", "六"} // 添加中文星期缩写数组
	for i := 29; i >= 0; i-- {
		t := now.AddDate(0, 0, -i)
		day := t.Day()
		weekday := weekdays[t.Weekday()] // 获取中文星期缩写
		// 修改日期格式为 "日(周几)"
		displayDate := fmt.Sprintf("%d|%s", day, weekday)
		dateKey := t.Format("01-02")
		stat := dayStats[dateKey]
		total := stat.success + stat.fail
		Data = append(Data, []string{
			displayDate,
			strconv.Itoa(stat.fail),
			strconv.Itoa(stat.success),
			strconv.Itoa(total),
		})
	}

	return Data
}
