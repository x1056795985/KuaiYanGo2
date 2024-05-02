package L_chart

import (
	. "EFunc/utils"
	"server/global"
	DB "server/structs/db"
	"time"
)

func Q取余额消费排行榜(Type int64) (data2 map[string]interface{}, err error) {

	/*	SELECT User, SUM(count) AS total_spent FROM db_Log_Money
		WHERE COUNT<0 and USER IN ("刘备") GROUP BY User ORDER BY total_spent DESC*/
	data2 = make(map[string]interface{}, 3)

	db := *global.GVA_DB
	var data []struct {
		User        string  `json:"User"`
		Total_spent float64 `gorm:"column:total_spent"`
	}
	var 局_时间戳 = time.Now().Unix()

	switch Type {
	default:
		局_时间戳 -= 1 * 86400
	case 2:
		局_时间戳 -= 7 * 86400
	case 3:
		局_时间戳 -= 30 * 86400
	}

	err = db.Model(DB.DB_LogMoney{}).Select("User, SUM(count) AS total_spent").
		Where("Count<0").
		Where("Time>?", 局_时间戳).
		Group("User").Limit(10).
		Order("total_spent Desc").Scan(&data).Error

	var 用户名 = make([]string, 0, len(data))
	var 消费 = make([]float64, 0, len(data))
	for _, v := range data {
		用户名 = append(用户名, v.User)
		消费 = append(消费, Float64取绝对值(v.Total_spent))
	}

	var data同比增长 []struct {
		User        string  `json:"User"`
		Total_spent float64 `gorm:"column:total_spent"`
	}
	err = db.Model(DB.DB_LogMoney{}).Select("User, SUM(count) AS total_spent").
		Where("Count>0").
		Where("Time>?", 局_时间戳).
		Where("User IN ?", 用户名).
		Group("User").
		Order("total_spent ASC").Scan(&data同比增长).Error

	var 增长 = make([]float64, len(data))
	for i, v := range data {
		增长[i] = 0
		for _, v2 := range data同比增长 {
			if v.User == v2.User {
				增长[i] = v2.Total_spent
				break
			}
		}
	}
	data2["user"] = 用户名
	data2["x"] = 消费
	data2["y"] = 增长
	return
}
func Q取余额增长排行榜(Type int64) (data2 map[string]interface{}, err error) {

	/*	SELECT User, SUM(count) AS total_spent FROM db_Log_Money
		WHERE COUNT<0 and USER IN ("刘备") GROUP BY User ORDER BY total_spent DESC*/
	data2 = make(map[string]interface{}, 3)

	db := *global.GVA_DB
	var data []struct {
		User        string  `json:"User"`
		Total_spent float64 `gorm:"column:total_spent"`
	}
	var 局_时间戳 = time.Now().Unix()

	switch Type {
	default:
		局_时间戳 -= 1 * 86400
	case 2:
		局_时间戳 -= 7 * 86400
	case 3:
		局_时间戳 -= 30 * 86400
	}

	err = db.Model(DB.DB_LogMoney{}).Select("User, SUM(count) AS total_spent").
		Where("Count>0").
		Where("Time>?", 局_时间戳).
		Group("User").Limit(10).
		Order("total_spent Desc").Scan(&data).Error

	var 用户名 = make([]string, 0, len(data))
	var 增长 = make([]float64, 0, len(data))
	for _, v := range data {
		用户名 = append(用户名, v.User)
		增长 = append(增长, Float64取绝对值(v.Total_spent))
	}

	var data同比消费 []struct {
		User        string  `json:"User"`
		Total_spent float64 `gorm:"column:total_spent"`
	}
	err = db.Model(DB.DB_LogMoney{}).Select("User, SUM(count) AS total_spent").
		Where("Count<0").
		Where("Time>?", 局_时间戳).
		Where("User IN ?", 用户名).
		Group("User").
		Order("total_spent ASC").Scan(&data同比消费).Error

	var 消费 = make([]float64, len(data))
	for i, v := range data {
		消费[i] = 0
		for _, v2 := range data同比消费 {
			if v.User == v2.User {
				消费[i] = Float64取绝对值(v2.Total_spent)
				break
			}
		}
	}
	data2["user"] = 用户名
	data2["x"] = 增长
	data2["y"] = 消费
	return
}
func Q取积分消费排行榜(Type int64) (data2 map[string]interface{}, err error) {

	/*	SELECT USER,AppId, SUM(count) AS total_spent FROM db_Log_VipNumber
		WHERE COUNT<0  GROUP BY User,AppId ORDER BY total_spent DESC*/

	data2 = make(map[string]interface{}, 2)

	db := *global.GVA_DB
	var data []struct {
		User        string  `json:"User"`
		Total_spent float64 `gorm:"column:total_spent"`
	}
	var 局_时间戳 = time.Now().Unix()

	switch Type {
	default:
		局_时间戳 -= 1 * 86400
	case 2:
		局_时间戳 -= 7 * 86400
	case 3:
		局_时间戳 -= 30 * 86400
	}

	err = db.Model(DB.DB_LogVipNumber{}).Select("CONCAT(AppId, '-',User) AS User, SUM(count) AS total_spent").
		Where("Count<0").
		Where("Time>?", 局_时间戳).
		Group("User").Group("AppId").Limit(20).
		Order("total_spent ASC").Scan(&data).Error

	var 用户名 = make([]string, 0, len(data))
	var 消费 = make([]float64, 0, len(data))
	for _, v := range data {
		用户名 = append(用户名, v.User)
		消费 = append(消费, Float64取绝对值(v.Total_spent))
	}

	data2["user"] = 用户名
	data2["series"] = 消费
	return
}
