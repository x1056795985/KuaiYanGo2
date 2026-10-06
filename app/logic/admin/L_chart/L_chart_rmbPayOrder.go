package L_chart

import (
	"math"
	"server/app/global"
	"server/app/models/constant"
	"server/app/models/dbm"
	"strconv"
	"time"
)

// 金额_保留2位小数 金额四舍五入保留2位小数
func 金额_保留2位小数(值 float64) float64 {
	return math.Round(值*100) / 100
}

// ============ 充值订单(db_Log_RMBPayOrder)图表统计 ============
// 统计口径: 收入类统计仅统计成功订单(Status=3), 金额取实付金额ActualRmb(无实付金额时取订单金额Rmb)

// 时间_充值订单周期开始 取统计周期开始时间戳(当天/本周/本月 0点)
// Type: 1今日 2本周(周一为一周开始) 3本月
func 时间_充值订单周期开始(Type int, now time.Time) int64 {
	局_今日0点 := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	switch Type {
	case 2:
		局_周偏移天数 := (int(局_今日0点.Weekday()) + 6) % 7 //周日按一周最后一天算
		return 局_今日0点.Unix() - int64(局_周偏移天数)*86400
	case 3:
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Unix()
	default:
		return 局_今日0点.Unix()
	}
}

// Q取AppId名称映射 取所有应用 AppId->AppName 映射(AppId=2为代理平台)
func Q取AppId名称映射() (map[int]string, error) {
	db := global.Get局db()
	var 局_应用列表 []dbm.DB_AppInfo
	if err := db.Model(dbm.DB_AppInfo{}).Find(&局_应用列表).Error; err != nil {
		return nil, err
	}
	局_映射 := map[int]string{2: "代理平台"}
	for _, v := range 局_应用列表 {
		局_映射[v.AppId] = v.AppName
	}
	return 局_映射, nil
}

// 统计_成功金额表达式 收入金额取实付金额,旧数据无实付金额时取订单金额
func 统计_成功金额表达式() string {
	return "IFNULL(NULLIF(ActualRmb,0),Rmb)"
}

// Q取充值订单仪表台汇总 充值订单顶部汇总卡片(今日/本周/本月收入及同比,客单价,待处理订单)
func Q取充值订单仪表台汇总() (data2 map[string]interface{}, err error) {
	data2 = make(map[string]interface{}, 12)
	db := global.Get局db()
	局_now := time.Now()
	局_今日0点 := 时间_充值订单周期开始(1, 局_now)
	局_昨日0点 := 局_今日0点 - 86400
	局_本周0点 := 时间_充值订单周期开始(2, 局_now)
	局_本月0点 := 时间_充值订单周期开始(3, 局_now)
	局_上月0点 := time.Date(局_now.Year(), 局_now.Month()-1, 1, 0, 0, 0, 0, 局_now.Location()).Unix()
	局_金额表达式 := 统计_成功金额表达式()

	// 局_查区间统计 查询[开始,结束)时间内成功订单总金额与单数,结束时间<=0表示不限制结束
	局_查区间统计 := func(开始时间 int64, 结束时间 int64) (金额 float64, 单数 int64, err error) {
		var 局_data struct {
			TotalRmb   float64 `gorm:"column:total_rmb"`
			OrderCount int64   `gorm:"column:order_count"`
		}
		局_db := db.Model(dbm.DB_LogRMBPayOrder{}).
			Select("IFNULL(SUM("+局_金额表达式+"),0) AS total_rmb, COUNT(*) AS order_count").
			Where("Status = ?", constant.D订单状态_成功)
		if 结束时间 > 0 {
			局_db = 局_db.Where("Time >= ? AND Time < ?", 开始时间, 结束时间)
		} else {
			局_db = 局_db.Where("Time >= ?", 开始时间)
		}
		err = 局_db.Scan(&局_data).Error
		return 局_data.TotalRmb, 局_data.OrderCount, err
	}

	var 局_今日金额, 局_昨日金额, 局_本周金额, 局_本月金额, 局_上月金额 float64
	var 局_今日单数, 局_本周单数, 局_本月单数, 局_上月单数 int64
	if 局_今日金额, 局_今日单数, err = 局_查区间统计(局_今日0点, 0); err != nil {
		return
	}
	if 局_昨日金额, _, err = 局_查区间统计(局_昨日0点, 局_今日0点); err != nil {
		return
	}
	if 局_本周金额, 局_本周单数, err = 局_查区间统计(局_本周0点, 0); err != nil {
		return
	}
	if 局_本月金额, 局_本月单数, err = 局_查区间统计(局_本月0点, 0); err != nil {
		return
	}
	if 局_上月金额, 局_上月单数, err = 局_查区间统计(局_上月0点, 局_本月0点); err != nil {
		return
	}

	//待处理订单(等待支付+已付待处理)
	var 局_待处理单数 int64
	if err = db.Model(dbm.DB_LogRMBPayOrder{}).
		Where("Status IN ?", []int{constant.D订单状态_等待支付, constant.D订单状态_已付待处理}).
		Count(&局_待处理单数).Error; err != nil {
		return
	}

	//本月平均客单价
	var 局_本月客单价 float64
	if 局_本月单数 > 0 {
		局_本月客单价 = 金额_保留2位小数(局_本月金额 / float64(局_本月单数))
	}

	data2["今日金额"] = 金额_保留2位小数(局_今日金额)
	data2["今日单数"] = 局_今日单数
	data2["昨日金额"] = 金额_保留2位小数(局_昨日金额)
	data2["本周金额"] = 金额_保留2位小数(局_本周金额)
	data2["本周单数"] = 局_本周单数
	data2["本月金额"] = 金额_保留2位小数(局_本月金额)
	data2["本月单数"] = 局_本月单数
	data2["上月金额"] = 金额_保留2位小数(局_上月金额)
	data2["上月单数"] = 局_上月单数
	data2["本月客单价"] = 局_本月客单价
	data2["待处理单数"] = 局_待处理单数
	return
}

// Q取充值订单分应用月收入统计 分应用本月收入与上月收入对比(仅成功订单)
func Q取充值订单分应用月收入统计() (data2 map[string]interface{}, err error) {
	data2 = make(map[string]interface{}, 6)
	db := global.Get局db()
	局_now := time.Now()
	局_本月0点 := 时间_充值订单周期开始(3, 局_now)
	局_上月0点 := time.Date(局_now.Year(), 局_now.Month()-1, 1, 0, 0, 0, 0, 局_now.Location()).Unix()
	局_金额表达式 := 统计_成功金额表达式()

	type 结构_应用统计 struct {
		AppId      int     `gorm:"column:AppId"`
		TotalRmb   float64 `gorm:"column:total_rmb"`
		OrderCount int64   `gorm:"column:order_count"`
	}
	var 局_本月数据 []结构_应用统计
	var 局_上月数据 []结构_应用统计
	if err = db.Model(dbm.DB_LogRMBPayOrder{}).
		Select("AppId, SUM("+局_金额表达式+") AS total_rmb, COUNT(*) AS order_count").
		Where("Status = ? AND Time >= ?", constant.D订单状态_成功, 局_本月0点).
		Group("AppId").Order("total_rmb DESC").Scan(&局_本月数据).Error; err != nil {
		return
	}
	if err = db.Model(dbm.DB_LogRMBPayOrder{}).
		Select("AppId, SUM("+局_金额表达式+") AS total_rmb, COUNT(*) AS order_count").
		Where("Status = ? AND Time >= ? AND Time < ?", constant.D订单状态_成功, 局_上月0点, 局_本月0点).
		Group("AppId").Scan(&局_上月数据).Error; err != nil {
		return
	}

	局_AppId名称, err2 := Q取AppId名称映射()
	if err2 != nil {
		err = err2
		return
	}

	//以本月收入降序为主,补齐上月有但本月没有的应用
	局_上月金额Map := make(map[int]float64, len(局_上月数据))
	局_上月单数Map := make(map[int]int64, len(局_上月数据))
	for _, v := range 局_上月数据 {
		局_上月金额Map[v.AppId] = v.TotalRmb
		局_上月单数Map[v.AppId] = v.OrderCount
	}
	局_本月AppId集合 := make(map[int]bool, len(局_本月数据))
	局_AppId数组 := make([]int, 0, len(局_本月数据)+len(局_上月数据))
	局_名称数组 := make([]string, 0, len(局_AppId数组))
	局_本月金额数组 := make([]float64, 0, len(局_AppId数组))
	局_本月单数数组 := make([]int64, 0, len(局_AppId数组))
	追加应用 := func(AppId int, 金额 float64, 单数 int64) {
		局_本月AppId集合[AppId] = true
		名称 := 局_AppId名称[AppId]
		if 名称 == "" {
			名称 = "未知应用(" + strconv.Itoa(AppId) + ")"
		}
		局_AppId数组 = append(局_AppId数组, AppId)
		局_名称数组 = append(局_名称数组, 名称)
		局_本月金额数组 = append(局_本月金额数组, 金额_保留2位小数(金额))
		局_本月单数数组 = append(局_本月单数数组, 单数)
	}
	for _, v := range 局_本月数据 {
		追加应用(v.AppId, v.TotalRmb, v.OrderCount)
	}
	for _, v := range 局_上月数据 {
		if 局_本月AppId集合[v.AppId] {
			continue
		}
		追加应用(v.AppId, 0, 0)
	}

	局_上月金额数组 := make([]float64, len(局_AppId数组))
	局_上月单数数组 := make([]int64, len(局_AppId数组))
	for i, AppId := range 局_AppId数组 {
		局_上月金额数组[i] = 金额_保留2位小数(局_上月金额Map[AppId])
		局_上月单数数组[i] = 局_上月单数Map[AppId]
	}

	data2["AppId"] = 局_AppId数组
	data2["应用"] = 局_名称数组
	data2["本月金额"] = 局_本月金额数组
	data2["本月单数"] = 局_本月单数数组
	data2["上月金额"] = 局_上月金额数组
	data2["上月单数"] = 局_上月单数数组
	return
}

// Q取充值订单分应用近7天统计 分应用近7天每天成功订单金额折线(含今天共7天)
func Q取充值订单分应用近7天统计() (data2 map[string]interface{}, err error) {
	data2 = make(map[string]interface{}, 2)
	db := global.Get局db()
	局_now := time.Now()
	局_今日0点 := 时间_充值订单周期开始(1, 局_now)
	局_7天前0点 := 局_今日0点 - 6*86400 //含今天共7天
	局_金额表达式 := 统计_成功金额表达式()

	var 局_data []struct {
		AppId    int     `gorm:"column:AppId"`
		Day      string  `gorm:"column:day"`
		TotalRmb float64 `gorm:"column:total_rmb"`
	}
	if err = db.Model(dbm.DB_LogRMBPayOrder{}).
		Select("AppId, FROM_UNIXTIME(Time, '%Y-%m-%d') AS day, SUM("+局_金额表达式+") AS total_rmb").
		Where("Status = ? AND Time >= ?", constant.D订单状态_成功, 局_7天前0点).
		Group("AppId, day").Scan(&局_data).Error; err != nil {
		return
	}

	//生成近7天日期数组及日期->索引映射(6天前->今天)
	局_日期数组 := make([]string, 7)
	局_日期索引 := make(map[string]int, 7)
	for i := 6; i >= 0; i-- {
		局_整天 := time.Unix(局_今日0点-int64(i)*86400, 0)
		索引 := 6 - i
		局_日期数组[索引] = 局_整天.Format("01-02")
		局_日期索引[局_整天.Format("2006-01-02")] = 索引
	}

	局_AppId名称, err2 := Q取AppId名称映射()
	if err2 != nil {
		err = err2
		return
	}

	//按应用聚合每天金额
	局_应用金额Map := make(map[int][]float64)
	局_AppId顺序 := make([]int, 0, 8)
	for _, v := range 局_data {
		if _, ok := 局_应用金额Map[v.AppId]; !ok {
			局_应用金额Map[v.AppId] = make([]float64, 7)
			局_AppId顺序 = append(局_AppId顺序, v.AppId)
		}
		if 索引, ok := 局_日期索引[v.Day]; ok {
			局_应用金额Map[v.AppId][索引] += v.TotalRmb
		}
	}

	type 结构_折线系列 struct {
		Name string    `json:"name"`
		Data []float64 `json:"data"`
	}
	局_系列数组 := make([]结构_折线系列, 0, len(局_AppId顺序))
	for _, AppId := range 局_AppId顺序 {
		名称 := 局_AppId名称[AppId]
		if 名称 == "" {
			名称 = "未知应用(" + strconv.Itoa(AppId) + ")"
		}
		局_每天金额 := 局_应用金额Map[AppId]
		for i := range 局_每天金额 {
			局_每天金额[i] = 金额_保留2位小数(局_每天金额[i])
		}
		局_系列数组 = append(局_系列数组, 结构_折线系列{Name: 名称, Data: 局_每天金额})
	}

	data2["日期"] = 局_日期数组
	data2["系列"] = 局_系列数组
	return
}

// Q取充值订单用户充值排行榜 用户充值排行榜TOP10(仅成功订单)
// 按 AppId+Uid 分组(同用户在不同应用算多人), 显示名=应用名-用户名(卡号)
// Type: 1今日 2本周 3本月
func Q取充值订单用户充值排行榜(Type int) (data2 map[string]interface{}, err error) {
	data2 = make(map[string]interface{}, 5)
	db := global.Get局db()
	局_开始时间 := 时间_充值订单周期开始(Type, time.Now())
	局_金额表达式 := 统计_成功金额表达式()

	var 局_data []struct {
		AppId      int     `gorm:"column:AppId"`
		Uid        int     `gorm:"column:Uid"`
		UserName   string  `gorm:"column:user_name"`
		TotalRmb   float64 `gorm:"column:total_rmb"`
		OrderCount int64   `gorm:"column:order_count"`
	}
	if err = db.Model(dbm.DB_LogRMBPayOrder{}).
		Select("AppId, Uid, IFNULL(MAX(User),'') AS user_name, SUM("+局_金额表达式+") AS total_rmb, COUNT(*) AS order_count").
		Where("Status = ? AND Time >= ? AND Uid > 0", constant.D订单状态_成功, 局_开始时间).
		Group("AppId, Uid").Order("total_rmb DESC").Limit(10).
		Scan(&局_data).Error; err != nil {
		return
	}

	局_AppId名称, err2 := Q取AppId名称映射()
	if err2 != nil {
		err = err2
		return
	}

	局_用户数组 := make([]string, 0, len(局_data))
	局_金额数组 := make([]float64, 0, len(局_data))
	局_单数数组 := make([]int64, 0, len(局_data))
	局_AppId数组 := make([]int, 0, len(局_data))
	局_Uid数组 := make([]int, 0, len(局_data))
	for _, v := range 局_data {
		名称 := 局_AppId名称[v.AppId]
		if 名称 == "" {
			名称 = "未知应用(" + strconv.Itoa(v.AppId) + ")"
		}
		if v.UserName == "" {
			v.UserName = "未知用户"
		}
		局_用户数组 = append(局_用户数组, 名称+"-"+v.UserName)
		局_金额数组 = append(局_金额数组, 金额_保留2位小数(v.TotalRmb))
		局_单数数组 = append(局_单数数组, v.OrderCount)
		局_AppId数组 = append(局_AppId数组, v.AppId)
		局_Uid数组 = append(局_Uid数组, v.Uid)
	}

	data2["用户"] = 局_用户数组
	data2["金额"] = 局_金额数组
	data2["单数"] = 局_单数数组
	data2["AppId"] = 局_AppId数组
	data2["Uid"] = 局_Uid数组
	return
}

// Q取充值订单支付方式统计 近30天成功订单支付方式金额占比
func Q取充值订单支付方式统计() (data2 map[string]interface{}, err error) {
	data2 = make(map[string]interface{}, 3)
	db := global.Get局db()
	局_30天前 := time.Now().Unix() - 30*86400
	局_金额表达式 := 统计_成功金额表达式()

	var 局_data []struct {
		Type       string  `gorm:"column:Type"`
		TotalRmb   float64 `gorm:"column:total_rmb"`
		OrderCount int64   `gorm:"column:order_count"`
	}
	if err = db.Model(dbm.DB_LogRMBPayOrder{}).
		Select("Type, SUM("+局_金额表达式+") AS total_rmb, COUNT(*) AS order_count").
		Where("Status = ? AND Time >= ?", constant.D订单状态_成功, 局_30天前).
		Group("Type").Order("total_rmb DESC").Limit(10).
		Scan(&局_data).Error; err != nil {
		return
	}

	局_名称数组 := make([]string, 0, len(局_data))
	局_金额数组 := make([]float64, 0, len(局_data))
	局_单数数组 := make([]int64, 0, len(局_data))
	for _, v := range 局_data {
		名称 := v.Type
		if 名称 == "" {
			名称 = "未知方式"
		}
		局_名称数组 = append(局_名称数组, 名称)
		局_金额数组 = append(局_金额数组, 金额_保留2位小数(v.TotalRmb))
		局_单数数组 = append(局_单数数组, v.OrderCount)
	}

	data2["名称"] = 局_名称数组
	data2["金额"] = 局_金额数组
	data2["单数"] = 局_单数数组
	return
}

// Q取充值订单金额区间分布 近30天成功订单充值金额区间分布(单数+金额)
func Q取充值订单金额区间分布() (data2 map[string]interface{}, err error) {
	data2 = make(map[string]interface{}, 3)
	db := global.Get局db()
	局_30天前 := time.Now().Unix() - 30*86400
	局_金额表达式 := 统计_成功金额表达式()

	var 局_data struct {
		Count1 int64 `gorm:"column:count_1"` //0~10
		Count2 int64 `gorm:"column:count_2"` //10~30
		Count3 int64 `gorm:"column:count_3"` //30~50
		Count4 int64 `gorm:"column:count_4"` //50~100
		Count5 int64 `gorm:"column:count_5"` //100~300
		Count6 int64 `gorm:"column:count_6"` //300以上
		Rmb1   float64 `gorm:"column:rmb_1"`
		Rmb2   float64 `gorm:"column:rmb_2"`
		Rmb3   float64 `gorm:"column:rmb_3"`
		Rmb4   float64 `gorm:"column:rmb_4"`
		Rmb5   float64 `gorm:"column:rmb_5"`
		Rmb6   float64 `gorm:"column:rmb_6"`
	}
	if err = db.Model(dbm.DB_LogRMBPayOrder{}).Select(
		"SUM(CASE WHEN Rmb>=0 AND Rmb<10 THEN 1 ELSE 0 END) AS count_1, "+
			"SUM(CASE WHEN Rmb>=10 AND Rmb<30 THEN 1 ELSE 0 END) AS count_2, "+
			"SUM(CASE WHEN Rmb>=30 AND Rmb<50 THEN 1 ELSE 0 END) AS count_3, "+
			"SUM(CASE WHEN Rmb>=50 AND Rmb<100 THEN 1 ELSE 0 END) AS count_4, "+
			"SUM(CASE WHEN Rmb>=100 AND Rmb<300 THEN 1 ELSE 0 END) AS count_5, "+
			"SUM(CASE WHEN Rmb>=300 THEN 1 ELSE 0 END) AS count_6, "+
			"SUM(CASE WHEN Rmb>=0 AND Rmb<10 THEN "+局_金额表达式+" ELSE 0 END) AS rmb_1, "+
			"SUM(CASE WHEN Rmb>=10 AND Rmb<30 THEN "+局_金额表达式+" ELSE 0 END) AS rmb_2, "+
			"SUM(CASE WHEN Rmb>=30 AND Rmb<50 THEN "+局_金额表达式+" ELSE 0 END) AS rmb_3, "+
			"SUM(CASE WHEN Rmb>=50 AND Rmb<100 THEN "+局_金额表达式+" ELSE 0 END) AS rmb_4, "+
			"SUM(CASE WHEN Rmb>=100 AND Rmb<300 THEN "+局_金额表达式+" ELSE 0 END) AS rmb_5, "+
			"SUM(CASE WHEN Rmb>=300 THEN "+局_金额表达式+" ELSE 0 END) AS rmb_6").
		Where("Status = ? AND Time >= ?", constant.D订单状态_成功, 局_30天前).
		Scan(&局_data).Error; err != nil {
		return
	}

	局_区间数组 := []string{"0~10", "10~30", "30~50", "50~100", "100~300", "300以上"}
	局_单数数组 := []int64{局_data.Count1, 局_data.Count2, 局_data.Count3, 局_data.Count4, 局_data.Count5, 局_data.Count6}
	局_金额数组 := []float64{
		金额_保留2位小数(局_data.Rmb1), 金额_保留2位小数(局_data.Rmb2), 金额_保留2位小数(局_data.Rmb3),
		金额_保留2位小数(局_data.Rmb4), 金额_保留2位小数(局_data.Rmb5), 金额_保留2位小数(局_data.Rmb6),
	}

	data2["区间"] = 局_区间数组
	data2["单数"] = 局_单数数组
	data2["金额"] = 局_金额数组
	return
}
