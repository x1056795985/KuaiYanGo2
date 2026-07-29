package L_chart

import (
	. "EFunc/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"server/new/app/global"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	"sort"
	"strconv"
	"strings"
	"time"
)

const 系统演示模式 = 1

type 临时应用id总数键值对 struct {
	应用AppId int
	总数      int
	在线数量    int
}

type 结构_请求类型 struct {
	Type   int `json:"Type"`
	AppId  int `json:"AppId"`
	Offset int `json:"Offset"` // 时间偏移量(天数或月数),0=当前周期,负数=往前偏移
}

// Get在线用户Ip地图分布统计 在线用户IP地图分布统计
func Get在线用户Ip地图分布统计(c *gin.Context) []gin.H {
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

// Get在线用户统计 在线用户统计
func Get在线用户统计(c *gin.Context) []gin.H {
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
	db := *global.GVA_DB
	sAppInfo := service.NewAppInfo(c, &db)
	var 局_appId列表 []int
	var 局_appId名称 = sAppInfo.AppInfo取map列表Int(true)
	_ = global.GVA_DB.Model(dbm.DB_LinksToken{}).Distinct("LoginAppid").Find(&局_appId列表).Error
	var Data = make([]gin.H, 0, len(局_appId列表))
	var 局_数量 int64
	for 索引, _ := range 局_appId列表 {
		局_数量 = 0
		global.GVA_DB.Model(dbm.DB_LinksToken{}).Where("LoginAppid=?", 局_appId列表[索引]).Where("Status=1").Where("User!=?", "游客").Count(&局_数量)
		if 局_数量 > 0 {
			Data = append(Data, gin.H{"name": 局_appId名称[局_appId列表[索引]], "value": 局_数量})
		}
	}

	if time.Now().Unix()-局_耗时 > 5 { //超过5秒的缓存
		global.H缓存.Set("图表数据_Get在线用户统计", Data, time.Minute*10)
	}

	return Data
}

// Get统计用户日活月活 统计用户日活月活
func Get统计用户日活月活(c *gin.Context) []gin.H {
	局_type := 结构_请求类型{Type: 1, AppId: 0}
	_ = c.ShouldBindJSON(&局_type)

	Data缓存, ok := global.H缓存.Get("Get在线用户统计登录时间" + strconv.Itoa(局_type.Type) + "|" + strconv.Itoa(局_type.AppId))
	if ok {
		return Data缓存.([]gin.H)
	}

	局_耗时 := time.Now().Unix()
	局_活动数量 := make([]int, 30) // 30天 或12个月
	if 局_type.Type == 2 {
		局_活动数量 = make([]int, 12) // 30天 或12个月
	}

	db := *global.GVA_DB
	局_where := make(map[string]interface{}, 2)

	if 局_type.AppId > 10000 {
		局_where["appId"] = 局_type.AppId
	}
	for i := range len(局_活动数量) {
		if 局_type.Type == 1 {
			局_时间 := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
			局_where["DateStr"] = 局_时间
		} else {
			局_时间 := time.Now().AddDate(0, -i, 0).Format("2006-01")
			局_where["DateStr"] = 局_时间
		}
		var 局_值 *int
		err := db.Model(&dbm.DB_LogUserActive{}).Select("SUM(count) as total").Where(局_where).Scan(&局_值).Error
		if err != nil {
			log.Println(err)
		}
		if 局_值 != nil {
			局_活动数量[len(局_活动数量)-i-1] = *局_值
		} else {
			局_活动数量[len(局_活动数量)-i-1] = 0
		}
	}

	Data := make([]gin.H, 1)
	Data[0] = gin.H{"name": "活动统计", "type": "line", "data": 局_活动数量}

	if time.Now().Unix()-局_耗时 > 5 { //超过5秒的缓存
		global.H缓存.Set("Get在线用户统计登录时间"+strconv.Itoa(局_type.Type)+"|"+strconv.Itoa(局_type.AppId), Data, time.Minute*5)
	}

	return Data
}

// Get统计分时段在线总数 统计分时段在线总数(一次性返回今日/昨日/前日三天数据,方便对比)
func Get统计分时段在线总数(c *gin.Context) []gin.H {
	请求 := 结构_请求类型{Type: 1, AppId: 0}
	_ = c.ShouldBindJSON(&请求)

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 三天的时间范围:前日0点 ~ 明日0点(即今日结束)
	局_前日0点 := today.AddDate(0, 0, -2).Unix()
	局_明日0点 := today.AddDate(0, 0, 1).Unix()

	Data缓存, ok := global.H缓存.Get("Get统计分时段在线总数_三天对比_" + strconv.FormatInt(int64(请求.AppId), 10))
	if ok {
		return Data缓存.([]gin.H)
	}

	局_耗时 := time.Now().Unix()

	// 单次查询三天数据,用一条SQL搞定,按createdAt正序排列
	var 局_临时 = []dbm.DB_TongJiZaiXian{}
	global.GVA_DB.Model(dbm.DB_TongJiZaiXian{}).
		Where("appId = ?", 请求.AppId).
		Where("createdAt >= ?", 局_前日0点).
		Where("createdAt < ?", 局_明日0点).
		Order("createdAt ASC").
		Find(&局_临时)

	// 准备三个24小时的数组
	局_今日数量 := make([]int64, 24)
	局_昨日数量 := make([]int64, 24)
	局_前日数量 := make([]int64, 24)
	局_登录时间 := make([]string, 24)
	for v := range 24 {
		局_登录时间[v] = strconv.Itoa(v) + "时"
	}

	// 今日0点、昨日0点的时间戳,用于判断每条记录属于哪一天
	局_今日0点 := today.Unix()
	局_昨日0点 := today.AddDate(0, 0, -1).Unix()

	for I := range 局_临时 {
		局_记录 := 局_临时[I]
		局_时 := time.Unix(局_记录.CreatedAt, 0).Hour()
		switch {
		case 局_记录.CreatedAt >= 局_今日0点:
			局_今日数量[局_时] += 局_记录.Count
		case 局_记录.CreatedAt >= 局_昨日0点:
			局_昨日数量[局_时] += 局_记录.Count
		default:
			局_前日数量[局_时] += 局_记录.Count
		}
	}

	Data := []gin.H{
		{"name": "今日", "type": "line", "data": 局_今日数量},
		{"name": "昨日", "type": "line", "data": 局_昨日数量},
		{"name": "前日", "type": "line", "data": 局_前日数量},
		{"name": "统计分时段在线时间", "type": "line", "data": 局_登录时间},
	}

	if time.Now().Unix()-局_耗时 > 5 { //超过5秒的缓存
		global.H缓存.Set("Get统计分时段在线总数_三天对比_"+strconv.FormatInt(int64(请求.AppId), 10), Data, time.Minute*5)
	}

	return Data
}

// Get应用用户类型统计 应用用户类型统计
func Get应用用户类型统计(c *gin.Context) []gin.H {
	局_type := 结构_请求类型{Type: 1, AppId: 10000}
	_ = c.ShouldBindJSON(&局_type)
	局_Appid := 局_type.AppId
	if 局_Appid < 10000 || global.GVA_Viper.GetInt("系统模式") == 系统演示模式 {
		var Data = make([]gin.H, 5)
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
	db := *global.GVA_DB
	var 局_ClassId列表 []int
	sUserClass := service.NewUserClass(c, &db)
	var 局_名称 = sUserClass.UserClass取map列表Int(局_Appid)
	局_名称[0] = "未分类"
	_ = global.GVA_DB.Model(dbm.DB_AppUser{}).Table("db_AppUser_" + strconv.Itoa(局_Appid)).Distinct("UserClassId").Find(&局_ClassId列表).Error
	var Data = make([]gin.H, len(局_ClassId列表))
	var 局_数量 int64
	for 索引, _ := range 局_ClassId列表 {
		局_数量 = 0
		global.GVA_DB.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(局_Appid)).Where("UserClassId=?", 局_ClassId列表[索引]).Count(&局_数量)
		Data[索引] = gin.H{"name": 局_名称[局_ClassId列表[索引]], "value": 局_数量}
	}

	if time.Now().Unix()-局_耗时 > 5 { //超过5秒的缓存
		global.H缓存.Set("图表数据_Get用户类型统计"+strconv.Itoa(局_Appid), Data, time.Minute*10)
	}

	return Data
}

// Get应用用户统计 应用用户统计
func Get应用用户统计(c *gin.Context) [][]string {
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
	db := *global.GVA_DB
	sAppInfo := service.NewAppInfo(c, &db)
	var 局_appId列表 = sAppInfo.AppInfo取map列表Int(false)
	sAppUser := service.NewAppUser(c, &db, 0)
	var 局_appId用户数量 = make([]临时应用id总数键值对, len(局_appId列表))
	局_I := 0
	for 键名, _ := range 局_appId列表 {
		局_appId用户数量[局_I] = 临时应用id总数键值对{应用AppId: 键名, 总数: sAppUser.Get用户总数(键名)}
		局_I++
	}

	sort.SliceStable(局_appId用户数量, func(i, j int) bool {
		return 局_appId用户数量[i].总数 < 局_appId用户数量[j].总数
	})
	sort.SliceStable(局_appId用户数量, func(i, j int) bool {
		if 局_appId用户数量[i].总数 != 局_appId用户数量[j].总数 {
			return 局_appId用户数量[i].总数 < 局_appId用户数量[j].总数
		}
		return false
	})

	var Data = make([][]string, len(局_appId用户数量)+1)
	Data[0] = []string{"product", "非会员", "会员", "总数"}
	for 索引 := 0; 索引 < len(局_appId用户数量); 索引++ {
		局_会员, 局_非会员 := sAppUser.Get用户会员和非会员数量(局_appId用户数量[索引].应用AppId)
		Data[索引+1] = []string{局_appId列表[局_appId用户数量[索引].应用AppId], strconv.FormatInt(局_非会员, 10), strconv.FormatInt(局_会员, 10), strconv.Itoa(局_appId用户数量[索引].总数)}
	}

	if time.Now().Unix()-局_耗时 > 5 { //超过5秒的缓存
		global.H缓存.Set("图表数据_Get应用用户统计", Data, time.Minute*10)
	}

	return Data
}

// Get卡号列表统计应用卡可用已用 卡号列表统计应用卡可用已用
func Get卡号列表统计应用卡可用已用(c *gin.Context) [][]string {
	局_type := 结构_请求类型{Type: 1}
	_ = c.ShouldBindJSON(&局_type)
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
	db := *global.GVA_DB
	sAppInfo := service.NewAppInfo(c, &db)
	sKa := service.NewKa(c, &db)
	sLinksToken := service.NewLinksToken(c, &db)
	var 局_appId列表 = sAppInfo.AppInfo取map列表Int(false)

	var 局_appId卡号数量 = make([]临时应用id总数键值对, len(局_appId列表))
	局_I := 0
	for 键名, _ := range 局_appId列表 {
		局_appId卡号数量[局_I] = 临时应用id总数键值对{应用AppId: 键名, 总数: sKa.Get卡号总数(键名), 在线数量: int(sLinksToken.Q指定应用真实在线(键名))}
		局_I++
	}

	sort.SliceStable(局_appId卡号数量, func(i, j int) bool {
		return 局_appId卡号数量[i].总数 < 局_appId卡号数量[j].总数
	})
	sort.SliceStable(局_appId卡号数量, func(i, j int) bool {
		if 局_appId卡号数量[i].总数 != 局_appId卡号数量[j].总数 {
			return 局_appId卡号数量[i].总数 < 局_appId卡号数量[j].总数
		}
		return false
	})

	var Data = make([][]string, len(局_appId卡号数量)+1)
	Data[0] = []string{"product", "已用", "未用", "总数", "在线"}
	for 索引 := 0; 索引 < len(局_appId卡号数量); 索引++ {
		局_已用, 局_未用 := sKa.Get应用已用和未用数量(局_appId卡号数量[索引].应用AppId)
		Data[索引+1] = []string{局_appId列表[局_appId卡号数量[索引].应用AppId], strconv.FormatInt(局_已用, 10), strconv.FormatInt(局_未用, 10), strconv.Itoa(局_appId卡号数量[索引].总数), strconv.Itoa(局_appId卡号数量[索引].在线数量)}
	}

	if time.Now().Unix()-局_耗时 > 5 { //超过5秒的缓存
		global.H缓存.Set("Get卡号列表统计可用已用"+strconv.Itoa(局_type.Type), Data, time.Minute*10)
	}

	return Data
}

// Get卡号列表统计应用卡类可用已用 卡号列表统计应用卡类可用已用
func Get卡号列表统计应用卡类可用已用(c *gin.Context) [][]string {
	局_type := 结构_请求类型{Type: 1, AppId: 10000}
	_ = c.ShouldBindJSON(&局_type)
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
	db := *global.GVA_DB
	sKaClass := service.NewKaClass(c, &db)
	sKa := service.NewKa(c, &db)
	var 局_appId卡类列表 = sKaClass.KaName取map列表Int(局_type.AppId)

	var 局_appId卡号数量 = make([]临时应用id总数键值对, len(局_appId卡类列表))
	局_I := 0
	for 键名, _ := range 局_appId卡类列表 {
		局_appId卡号数量[局_I] = 临时应用id总数键值对{应用AppId: 键名, 总数: sKa.Get卡类卡号总数(键名)}
		局_I++
	}

	sort.SliceStable(局_appId卡号数量, func(i, j int) bool {
		return 局_appId卡号数量[i].总数 < 局_appId卡号数量[j].总数
	})
	sort.SliceStable(局_appId卡号数量, func(i, j int) bool {
		if 局_appId卡号数量[i].总数 != 局_appId卡号数量[j].总数 {
			return 局_appId卡号数量[i].总数 < 局_appId卡号数量[j].总数
		}
		return false
	})

	var Data = make([][]string, len(局_appId卡号数量)+1)
	Data[0] = []string{"product", "已用", "未用", "总数"}
	for 索引 := 0; 索引 < len(局_appId卡号数量); 索引++ {
		局_已用, 局_未用 := sKa.Get卡类已用和未用数量(局_appId卡号数量[索引].应用AppId)
		Data[索引+1] = []string{局_appId卡类列表[局_appId卡号数量[索引].应用AppId], strconv.FormatInt(局_已用, 10), strconv.FormatInt(局_未用, 10), strconv.Itoa(局_appId卡号数量[索引].总数)}
	}

	if time.Now().Unix()-局_耗时 > 5 { //超过5秒的缓存
		global.H缓存.Set("Get卡号列表统计应用卡类可用已用"+strconv.Itoa(局_type.Type), Data, time.Minute*10)
	}

	return Data
}

// Get余额充值消费统计 余额充值消费统计
func Get余额充值消费统计(c *gin.Context) []gin.H {
	局_type := 结构_请求类型{Type: 1}
	_ = c.ShouldBindJSON(&局_type)
	var 时间处理函数 func(int) string
	if 局_type.Type == 2 {
		时间处理函数 = 取相对时间0点时间戳月
	} else {
		时间处理函数 = 取相对时间0点时间戳天
	}

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
	var 局_数量 [7]string

	global.GVA_DB.Model(dbm.DB_LogRMBPayOrder{}).
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

	global.GVA_DB.Model(dbm.DB_LogMoney{}).
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
			局_数量[索引-1] = strings.Replace(a, "-", "", 1)
		}
	}
	Data[1] = gin.H{"name": "消费金额", "type": "line", "data": 局_数量}

	if time.Now().Unix()-局_耗时 > 5 { //超过5秒的缓存
		global.H缓存.Set("图表数据_Get余额充值消费统计"+strconv.Itoa(局_type.Type), Data, time.Minute*10)
	}

	return Data
}

// Get积分点数消费统计 积分点数消费统计
func Get积分点数消费统计(c *gin.Context) []gin.H {
	局_type := 结构_请求类型{Type: 1}
	_ = c.ShouldBindJSON(&局_type)
	var 时间处理函数 func(int) string
	if 局_type.Type == 2 {
		时间处理函数 = 取相对时间0点时间戳月
	} else {
		时间处理函数 = 取相对时间0点时间戳天
	}

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
	var 局_数量 [7]string

	global.GVA_DB.Model(dbm.DB_LogVipNumber{}).
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
			局_数量[索引-1] = strings.Replace(a, "-", "", 1)
		}
	}
	Data[0] = gin.H{"name": "消费积分", "type": "line", "data": 局_数量}

	global.GVA_DB.Model(dbm.DB_LogVipNumber{}).
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
			局_数量[索引-1] = strings.Replace(a, "-", "", 1)
		}
	}
	Data[1] = gin.H{"name": "消费点数", "type": "line", "data": 局_数量}

	if time.Now().Unix()-局_耗时 > 5 { //超过5秒的缓存
		global.H缓存.Set("图表数据_Get积分点数消费统计"+strconv.Itoa(局_type.Type), Data, time.Minute*10)
	}

	return Data
}

// Get卡号列表统计制卡 卡号制卡统计(本月/上月/上上月按天对比)
// 单条SQL按天聚合三个月的制卡数据,返回三条series + x轴日期标签
func Get卡号列表统计制卡(c *gin.Context) []gin.H {
	局_type := 结构_请求类型{Type: 1, AppId: 0}
	_ = c.ShouldBindJSON(&局_type)

	now := time.Now()
	// 本月、上月、上上月的第一天0点
	本月1日 := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	上月1日 := 本月1日.AddDate(0, -1, 0)
	上上月1日 := 本月1日.AddDate(0, -2, 0)
	下月1日 := 本月1日.AddDate(0, 1, 0) // 查询上界(不含)

	// 演示模式
	if global.GVA_Viper.GetInt("系统模式") == 系统演示模式 {
		return 生成演示制卡数据(本月1日, 上月1日, 上上月1日)
	}

	缓存键 := "图表数据_制卡统计_三月对比_" + strconv.FormatInt(int64(局_type.AppId), 10)
	Data缓存, ok := global.H缓存.Get(缓存键)
	if ok {
		return Data缓存.([]gin.H)
	}

	局_耗时 := time.Now().Unix()

	// 单条SQL: 查询上上月1日~下月1日之间所有卡号,按 FLOOR(RegisterTime / 86400) 得到"天序号"
	// 然后用 Go 按 年-月-日 归入对应月份的数组
	type 日统计 struct {
		DayIdx int64 `gorm:"column:DayIdx"` // FLOOR(RegisterTime / 86400)
		Cnt    int64 `gorm:"column:Cnt"`
	}
	var 局_日统计 []日统计
	局_db := global.GVA_DB.Model(dbm.DB_Ka{}).
		Select("FLOOR(RegisterTime / 86400) AS DayIdx, COUNT(*) AS Cnt").
		Where("RegisterTime >= ? AND RegisterTime < ?", 上上月1日.Unix(), 下月1日.Unix()).
		Group("DayIdx").
		Order("DayIdx ASC")
	if 局_type.AppId > 0 {
		局_db = 局_db.Where("AppId = ?", 局_type.AppId)
	}
	局_db.Find(&局_日统计)

	// 调试: 打印查询结果帮助排查
	if len(局_日统计) == 0 {
		global.GVA_LOG.Println("[制卡统计] 未查到数据, AppId=", 局_type.AppId, " 范围=", 上上月1日.Unix(), "~", 下月1日.Unix())
	} else {
		global.GVA_LOG.Println("[制卡统计] 查到", len(局_日统计), "天数据, 第一天DayIdx=", 局_日统计[0].DayIdx, " Cnt=", 局_日统计[0].Cnt)
	}

	// 构建一个 map[天序号]数量, 方便O(1)查找
	局_天map := make(map[int64]int64, len(局_日统计))
	for _, v := range 局_日统计 {
		局_天map[v.DayIdx] = v.Cnt
	}

	// 三个月各自的天数
	本月天数 := 取月份天数(本月1日)
	上月天数 := 取月份天数(上月1日)
	上上月天数 := 取月份天数(上上月1日)
	// x轴取三者最大天数(最多31), 不足的月份对应位置留空(null)
	最大天数 := 本月天数
	if 上月天数 > 最大天数 {
		最大天数 = 上月天数
	}
	if 上上月天数 > 最大天数 {
		最大天数 = 上上月天数
	}

	局_x轴 := make([]string, 最大天数)
	局_本月数据 := make([]interface{}, 最大天数)
	局_上月数据 := make([]interface{}, 最大天数)
	局_上上月数据 := make([]interface{}, 最大天数)

	本月0点序号 := 本月1日.Unix() / 86400
	上月0点序号 := 上月1日.Unix() / 86400
	上上月0点序号 := 上上月1日.Unix() / 86400

	for 日 := int64(1); 日 <= int64(最大天数); 日++ {
		局_x轴[日-1] = strconv.FormatInt(日, 10) + "日"

		// 在该月实际天数内的日期,没数据就填0; 超出该月天数的日期保持null(折线断开)
		if int(日) <= 本月天数 {
			序号 := 本月0点序号 + 日 - 1
			if v, ok := 局_天map[序号]; ok {
				局_本月数据[日-1] = v
			} else {
				局_本月数据[日-1] = int64(0)
			}
		}
		if int(日) <= 上月天数 {
			序号 := 上月0点序号 + 日 - 1
			if v, ok := 局_天map[序号]; ok {
				局_上月数据[日-1] = v
			} else {
				局_上月数据[日-1] = int64(0)
			}
		}
		if int(日) <= 上上月天数 {
			序号 := 上上月0点序号 + 日 - 1
			if v, ok := 局_天map[序号]; ok {
				局_上上月数据[日-1] = v
			} else {
				局_上上月数据[日-1] = int64(0)
			}
		}
	}

	Data := []gin.H{
		{"name": "本月", "type": "line", "data": 局_本月数据},
		{"name": "上月", "type": "line", "data": 局_上月数据},
		{"name": "上上月", "type": "line", "data": 局_上上月数据},
		{"name": "x轴日期", "data": 局_x轴},
	}

	if time.Now().Unix()-局_耗时 > 5 { //超过5秒的缓存
		global.H缓存.Set(缓存键, Data, time.Minute*5)
	}

	return Data
}

// 取月份天数 返回该月有多少天
func 取月份天数(月1日 time.Time) int {
	下月1日 := 月1日.AddDate(0, 1, 0)
	return 下月1日.AddDate(0, 0, -1).Day()
}

// 生成演示制卡数据 演示模式下生成三个月的假数据
func 生成演示制卡数据(本月1日, 上月1日, 上上月1日 time.Time) []gin.H {
	本月天数 := 取月份天数(本月1日)
	上月天数 := 取月份天数(上月1日)
	上上月天数 := 取月份天数(上上月1日)
	最大天数 := 本月天数
	if 上月天数 > 最大天数 {
		最大天数 = 上月天数
	}
	if 上上月天数 > 最大天数 {
		最大天数 = 上上月天数
	}

	局_x轴 := make([]string, 最大天数)
	局_本月 := make([]int64, 最大天数)
	局_上月 := make([]int64, 最大天数)
	局_上上月 := make([]int64, 最大天数)
	for i := range 最大天数 {
		局_x轴[i] = strconv.Itoa(i+1) + "日"
		局_本月[i] = int64(100 + (i*13)%300)
		局_上月[i] = int64(80 + (i*17)%250)
		局_上上月[i] = int64(60 + (i*11)%200)
	}
	return []gin.H{
		{"name": "本月", "type": "line", "data": 局_本月},
		{"name": "上月", "type": "line", "data": 局_上月},
		{"name": "上上月", "type": "line", "data": 局_上上月},
		{"name": "x轴日期", "data": 局_x轴},
	}
}

// Get卡号月度汇总 卡号月度汇总
func Get卡号月度汇总(c *gin.Context) gin.H {
	局_type := 结构_请求类型{Type: 2, AppId: 0}
	_ = c.ShouldBindJSON(&局_type)

	if global.GVA_Viper.GetInt("系统模式") == 系统演示模式 {
		return gin.H{
			"本月制卡": 1500, "上月制卡": 1200,
			"本月使用": 800, "上月使用": 700,
		}
	}

	Data缓存, ok := global.H缓存.Get("图表数据_卡号月度汇总_" + strconv.Itoa(局_type.AppId))
	if ok {
		return Data缓存.(gin.H)
	}

	局_耗时 := time.Now().Unix()
	本月开始 := 取相对时间0点时间戳月(0)
	本月结束 := 取相对时间0点时间戳月(1)
	上月开始 := 取相对时间0点时间戳月(-1)

	局_db := global.GVA_DB.Model(dbm.DB_Ka{})
	if 局_type.AppId > 0 {
		局_db = 局_db.Where("AppId = ?", 局_type.AppId)
	}

	var 局_本月制卡, 局_上月制卡, 局_本月使用, 局_上月使用 int64
	局_db.Select("COUNT(*)").Where("RegisterTime >= ? AND RegisterTime < ?", 本月开始, 本月结束).Count(&局_本月制卡)
	局_db.Select("COUNT(*)").Where("RegisterTime >= ? AND RegisterTime < ?", 上月开始, 本月开始).Count(&局_上月制卡)
	局_db.Select("COUNT(*)").Where("UseTime > 0 AND UseTime >= ? AND UseTime < ?", 本月开始, 本月结束).Count(&局_本月使用)
	局_db.Select("COUNT(*)").Where("UseTime > 0 AND UseTime >= ? AND UseTime < ?", 上月开始, 本月开始).Count(&局_上月使用)

	Data := gin.H{
		"本月制卡": 局_本月制卡,
		"上月制卡": 局_上月制卡,
		"本月使用": 局_本月使用,
		"上月使用": 局_上月使用,
	}

	if time.Now().Unix()-局_耗时 > 5 {
		global.H缓存.Set("图表数据_卡号月度汇总_"+strconv.Itoa(局_type.AppId), Data, time.Minute*10)
	}

	return Data
}

// Get仪表台汇总 仪表台汇总
func Get仪表台汇总(c *gin.Context) gin.H {
	if global.GVA_Viper.GetInt("系统模式") == 系统演示模式 {
		return gin.H{
			"卡号总数":   9999,
			"卡号未使用":  3000,
			"本月充值总额": 8888.50,
			"上月充值总额": 6666.00,
		}
	}

	Data缓存, ok := global.H缓存.Get("图表数据_仪表台汇总")
	if ok {
		return Data缓存.(gin.H)
	}

	局_耗时 := time.Now().Unix()
	var 局_卡号总数, 局_卡号未使用 int64
	global.GVA_DB.Model(dbm.DB_Ka{}).Count(&局_卡号总数)
	global.GVA_DB.Model(dbm.DB_Ka{}).Where("UseTime = 0").Count(&局_卡号未使用)

	本月开始 := 取相对时间0点时间戳月(0)
	本月结束 := 取相对时间0点时间戳月(1)
	上月开始 := 取相对时间0点时间戳月(-1)

	var 局_本月充值, 局_上月充值 float64
	global.GVA_DB.Model(dbm.DB_LogRMBPayOrder{}).
		Where("Status = 3 AND Time >= ? AND Time < ?", 本月开始, 本月结束).
		Select("COALESCE(SUM(Rmb), 0)").Scan(&局_本月充值)
	global.GVA_DB.Model(dbm.DB_LogRMBPayOrder{}).
		Where("Status = 3 AND Time >= ? AND Time < ?", 上月开始, 本月开始).
		Select("COALESCE(SUM(Rmb), 0)").Scan(&局_上月充值)

	Data := gin.H{
		"卡号总数":   局_卡号总数,
		"卡号未使用":  局_卡号未使用,
		"本月充值总额": 局_本月充值,
		"上月充值总额": 局_上月充值,
	}

	if time.Now().Unix()-局_耗时 > 5 {
		global.H缓存.Set("图表数据_仪表台汇总", Data, time.Minute*10)
	}

	return Data
}

// Get卡号列表统计制卡_代理 代理卡号制卡统计(本月/上月/上上月按天对比)
// 与管理端逻辑相同,额外过滤 RegisterUser = 当前代理账号
func Get卡号列表统计制卡_代理(c *gin.Context) []gin.H {
	局_type := 结构_请求类型{Type: 1, AppId: 0}
	_ = c.ShouldBindJSON(&局_type)

	now := time.Now()
	本月1日 := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	上月1日 := 本月1日.AddDate(0, -1, 0)
	上上月1日 := 本月1日.AddDate(0, -2, 0)
	下月1日 := 本月1日.AddDate(0, 1, 0)

	if global.GVA_Viper.GetInt("系统模式") == 系统演示模式 {
		return 生成演示制卡数据(本月1日, 上月1日, 上上月1日)
	}

	缓存键 := "图表数据_代理制卡统计_三月对比_" + c.GetString("User") + "_" + strconv.FormatInt(int64(局_type.AppId), 10)
	Data缓存, ok := global.H缓存.Get(缓存键)
	if ok {
		return Data缓存.([]gin.H)
	}

	局_耗时 := time.Now().Unix()

	type 日统计 struct {
		DayIdx int64 `gorm:"column:DayIdx"`
		Cnt    int64 `gorm:"column:Cnt"`
	}
	var 局_日统计 []日统计
	局_db := global.GVA_DB.Model(dbm.DB_Ka{}).
		Select("FLOOR(RegisterTime / 86400) AS DayIdx, COUNT(*) AS Cnt").
		Where("RegisterTime >= ? AND RegisterTime < ?", 上上月1日.Unix(), 下月1日.Unix()).
		Where("RegisterUser = ?", c.GetString("User")).
		Group("DayIdx").
		Order("DayIdx ASC")
	if 局_type.AppId > 0 {
		局_db = 局_db.Where("AppId = ?", 局_type.AppId)
	}
	局_db.Find(&局_日统计)

	if len(局_日统计) == 0 {
		global.GVA_LOG.Println("[代理制卡统计] 未查到数据, User=", c.GetString("User"), " AppId=", 局_type.AppId, " 范围=", 上上月1日.Unix(), "~", 下月1日.Unix())
	}

	局_天map := make(map[int64]int64, len(局_日统计))
	for _, v := range 局_日统计 {
		局_天map[v.DayIdx] = v.Cnt
	}

	本月天数 := 取月份天数(本月1日)
	上月天数 := 取月份天数(上月1日)
	上上月天数 := 取月份天数(上上月1日)
	最大天数 := 本月天数
	if 上月天数 > 最大天数 {
		最大天数 = 上月天数
	}
	if 上上月天数 > 最大天数 {
		最大天数 = 上上月天数
	}

	局_x轴 := make([]string, 最大天数)
	局_本月数据 := make([]interface{}, 最大天数)
	局_上月数据 := make([]interface{}, 最大天数)
	局_上上月数据 := make([]interface{}, 最大天数)

	本月0点序号 := 本月1日.Unix() / 86400
	上月0点序号 := 上月1日.Unix() / 86400
	上上月0点序号 := 上上月1日.Unix() / 86400

	for 日 := int64(1); 日 <= int64(最大天数); 日++ {
		局_x轴[日-1] = strconv.FormatInt(日, 10) + "日"

		// 在该月实际天数内的日期,没数据就填0; 超出该月天数的日期保持null(折线断开)
		if int(日) <= 本月天数 {
			序号 := 本月0点序号 + 日 - 1
			if v, ok := 局_天map[序号]; ok {
				局_本月数据[日-1] = v
			} else {
				局_本月数据[日-1] = int64(0)
			}
		}
		if int(日) <= 上月天数 {
			序号 := 上月0点序号 + 日 - 1
			if v, ok := 局_天map[序号]; ok {
				局_上月数据[日-1] = v
			} else {
				局_上月数据[日-1] = int64(0)
			}
		}
		if int(日) <= 上上月天数 {
			序号 := 上上月0点序号 + 日 - 1
			if v, ok := 局_天map[序号]; ok {
				局_上上月数据[日-1] = v
			} else {
				局_上上月数据[日-1] = int64(0)
			}
		}
	}

	Data := []gin.H{
		{"name": "本月", "type": "line", "data": 局_本月数据},
		{"name": "上月", "type": "line", "data": 局_上月数据},
		{"name": "上上月", "type": "line", "data": 局_上上月数据},
		{"name": "x轴日期", "data": 局_x轴},
	}

	if time.Now().Unix()-局_耗时 > 5 { //超过5秒的缓存
		global.H缓存.Set(缓存键, Data, time.Minute*5)
	}

	return Data
}

// Get应用用户账号注册统计 应用用户账号注册统计
func Get应用用户账号注册统计(c *gin.Context) []gin.H {
	局_type := 结构_请求类型{Type: 1, AppId: 1}
	_ = c.ShouldBindJSON(&局_type)
	var 时间处理函数 func(int) string
	if 局_type.Type == 2 {
		时间处理函数 = 取相对时间0点时间戳月
	} else {
		时间处理函数 = 取相对时间0点时间戳天
	}

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

	global.GVA_DB.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(局_type.AppId)).
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

	if time.Now().Unix()-局_耗时 > 5 { //超过5秒的缓存
		global.H缓存.Set("图表数据_Get用户账号统计"+strconv.Itoa(局_type.Type)+"_"+strconv.Itoa(局_type.AppId), Data, time.Minute*10)
	}

	return Data
}

// Get用户账号登录注册统计 用户账号登录注册统计
func Get用户账号登录注册统计(c *gin.Context) []gin.H {
	局_type := 结构_请求类型{Type: 1, AppId: 1}
	_ = c.ShouldBindJSON(&局_type)
	var 时间处理函数 func(int) string
	if 局_type.Type == 2 {
		时间处理函数 = 取相对时间0点时间戳月
	} else {
		时间处理函数 = 取相对时间0点时间戳天
	}

	var Data = make([]gin.H, 2)
	if global.GVA_Viper.GetInt("系统模式") == 系统演示模式 {
		Data[0] = gin.H{"name": "注册数量", "type": "line", "data": []int{120, 132, 101, 134, 90, 230, 210}}
		Data[1] = gin.H{"name": "登录数量", "type": "line", "data": []int{220, 182, 191, 234, 290, 330, 310}}
		return Data
	}

	Data缓存, ok := global.H缓存.Get("图表数据_Get用户账号统计" + strconv.Itoa(局_type.Type) + "_" + strconv.Itoa(局_type.Offset))
	if ok {
		return Data缓存.([]gin.H)
	}

	局_耗时 := time.Now().Unix()
	var 局_临时 = make(map[string]interface{})
	var 局_数量 [7]int

	global.GVA_DB.Model(dbm.DB_User{}).
		Select("Count(case when ( RegisterTime between "+时间处理函数(局_type.Offset-6)+" and "+时间处理函数(局_type.Offset-5)+") then 1 else null end) as  '1' ",
			"Count(case when ( RegisterTime between "+时间处理函数(局_type.Offset-5)+" and "+时间处理函数(局_type.Offset-4)+") then 1 else null end) as  '2' ",
			"Count(case when ( RegisterTime between "+时间处理函数(局_type.Offset-4)+" and "+时间处理函数(局_type.Offset-3)+") then 1 else null end) as  '3' ",
			"Count(case when ( RegisterTime between "+时间处理函数(局_type.Offset-3)+" and "+时间处理函数(局_type.Offset-2)+") then 1 else null end) as  '4' ",
			"Count(case when ( RegisterTime between "+时间处理函数(局_type.Offset-2)+" and "+时间处理函数(局_type.Offset-1)+") then 1 else null end) as  '5' ",
			"Count(case when ( RegisterTime between "+时间处理函数(局_type.Offset-1)+" and "+时间处理函数(局_type.Offset)+") then 1 else null end) as  '6' ",
			"Count(case when ( RegisterTime between "+时间处理函数(局_type.Offset)+" and "+时间处理函数(局_type.Offset+1)+") then 1 else null end) as  '7' ").
		First(&局_临时)
	for 键名, 值 := range 局_临时 {
		索引, _ := strconv.Atoi(键名)
		if 值 == nil {
			局_数量[索引-1] = 0
		} else {
			a, _ := strconv.Atoi(D到文本(值))
			局_数量[索引-1] = a
		}
	}
	Data[0] = gin.H{"name": "注册数量", "type": "line", "data": 局_数量}

	//老老实实读取登录日志吧
	global.GVA_DB.Model(dbm.DB_LogLogin{}).
		Select("Count(case when ( Time between "+时间处理函数(局_type.Offset-6)+" and "+时间处理函数(局_type.Offset-5)+") then 1 else null end) as  '1' ",
			"Count(case when ( Time between "+时间处理函数(局_type.Offset-5)+" and "+时间处理函数(局_type.Offset-4)+") then 1 else null end) as  '2' ",
			"Count(case when ( Time between "+时间处理函数(局_type.Offset-4)+" and "+时间处理函数(局_type.Offset-3)+") then 1 else null end) as  '3' ",
			"Count(case when ( Time between "+时间处理函数(局_type.Offset-3)+" and "+时间处理函数(局_type.Offset-2)+") then 1 else null end) as  '4' ",
			"Count(case when ( Time between "+时间处理函数(局_type.Offset-2)+" and "+时间处理函数(局_type.Offset-1)+") then 1 else null end) as  '5' ",
			"Count(case when ( Time between "+时间处理函数(局_type.Offset-1)+" and "+时间处理函数(局_type.Offset)+") then 1 else null end) as  '6' ",
			"Count(case when ( Time between "+时间处理函数(局_type.Offset)+" and "+时间处理函数(局_type.Offset+1)+") then 1 else null end) as  '7' ").
		Where("Note = ?", "用户登录").First(&局_临时)
	for 键名, 值 := range 局_临时 {
		索引, _ := strconv.Atoi(键名)
		if 值 == nil {
			局_数量[索引-1] = 0
		} else {
			a, _ := strconv.Atoi(D到文本(值))
			局_数量[索引-1] = a
		}
	}
	Data[1] = gin.H{"name": "登录数量", "type": "line", "data": 局_数量}

	if time.Now().Unix()-局_耗时 > 5 { //超过5秒的缓存
		global.H缓存.Set("图表数据_Get用户账号统计"+strconv.Itoa(局_type.Type)+"_"+strconv.Itoa(局_type.Offset), Data, time.Minute*10)
	}

	return Data
}

// Get代理组织架构图 代理组织架构图
func Get代理组织架构图(c *gin.Context, 根代理ID int) []*Node {
	var 局_用户数组 []dbm.DB_User

	_ = global.GVA_DB.Model(dbm.DB_User{}).Select("Id", "User", "UPAgentId", "AgentDiscount").Where("UPAgentId !=0").Find(&局_用户数组).Error
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

	Data := getTreeIterative(nodes, 根代理ID)

	return Data
}

type Node struct {
	Id            int     `json:"Id" gorm:"column:Id;primarykey;AUTO_INCREMENT"`
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

// Get任务池任务Id分析 任务池任务Id分析
func Get任务池任务Id分析(c *gin.Context) [][]string {
	局_type := struct {
		TaskId int `json:"TaskId"`
	}{}
	_ = c.ShouldBindJSON(&局_type)

	局_总开始 := time.Now()
	now := time.Now()
	时间戳30天前 := now.AddDate(0, 0, -30).Unix()

	type 局_日统计 struct {
		DayNum  int `json:"DayNum"`
		Success int `json:"Success"`
		Fail    int `json:"Fail"`
	}
	var 局_统计结果 []局_日统计
	err := global.GVA_DB.Model(dbm.DB_TaskPoolData{}).
		Select("TimeStart DIV 86400 AS DayNum, "+
			"SUM(CASE WHEN Status = 3 THEN 1 ELSE 0 END) AS Success, "+
			"SUM(CASE WHEN Status = 4 THEN 1 ELSE 0 END) AS Fail").
		Where("Tid = ? AND TimeStart >= ?", 局_type.TaskId, 时间戳30天前).
		Group("DayNum").
		Find(&局_统计结果).Error
	_ = err

	global.GVA_LOG.Println(fmt.Sprintf("[任务Id分析] Tid=%d 查询耗时=%v 结果数=%d", 局_type.TaskId, time.Since(局_总开始), len(局_统计结果)))

	局_Map := make(map[string]*局_日统计, len(局_统计结果))
	for i := range 局_统计结果 {
		dateKey := time.Unix(int64(局_统计结果[i].DayNum)*86400, 0).Format("01-02")
		局_Map[dateKey] = &局_统计结果[i]
	}

	weekdays := []string{"日", "一", "二", "三", "四", "五", "六"}
	Data := [][]string{{"日期", "失败", "成功", "总数"}}
	for i := 29; i >= 0; i-- {
		t := now.AddDate(0, 0, -i)
		day := t.Day()
		weekday := weekdays[t.Weekday()]
		displayDate := fmt.Sprintf("%d|%s", day, weekday)
		dateKey := t.Format("01-02")
		success := 0
		fail := 0
		if stat, exists := 局_Map[dateKey]; exists {
			success = stat.Success
			fail = stat.Fail
		}
		total := success + fail
		Data = append(Data, []string{
			displayDate,
			strconv.Itoa(fail),
			strconv.Itoa(success),
			strconv.Itoa(total),
		})
	}

	global.GVA_LOG.Println(fmt.Sprintf("[任务Id分析] Tid=%d 总耗时=%v", 局_type.TaskId, time.Since(局_总开始)))
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
	timeStampFirstDay := time.Date(ts.Year(), ts.Month(), 1, 0, 0, 0, 0, ts.Location()).Unix()
	return strconv.Itoa(int(timeStampFirstDay))
}
