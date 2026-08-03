package luckyWheel

import (
	"encoding/json"
	"errors"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"server/app/logic/common/ka"
	dbm "server/app/models/db"
	"server/app/service"
)

// PrizeItem 奖品项
type PrizeItem struct {
	KaClassId    int    `json:"kaClassId"`
	Probability  int    `json:"probability"`
	Name         string `json:"name"`
}

// L抽奖_执行 核心抽奖逻辑
func L抽奖_执行(c *gin.Context, 数据库 *gorm.DB, appInfo dbm.DB_AppInfo, 在线信息 dbm.DB_LinksToken, 活动配置 dbm.DB_AppPromotionConfig) (中奖索引 int, 中奖奖品 PrizeItem, err error) {
	var info = struct {
		luckyWheelInfo dbm.DB_LuckyWheelInfo
		prizeList      []PrizeItem
		luckyWheelUser dbm.DB_LuckyWheelUser
		kaClass        dbm.DB_KaClass
	}{}

	// 查询大转盘配置
	info.luckyWheelInfo, err = service.NewLuckyWheelInfo(c, 数据库).Info(活动配置.TypeAssociatedId)
	if err != nil {
		err = errors.New("大转盘配置不存在")
		return
	}

	// 解析奖品列表
	if info.luckyWheelInfo.PrizeList == "" {
		info.luckyWheelInfo.PrizeList = "[]"
	}
	_ = json.Unmarshal([]byte(info.luckyWheelInfo.PrizeList), &info.prizeList)
	if len(info.prizeList) == 0 {
		err = errors.New("未配置奖品")
		return
	}

	// 今日日期标记
	局_今日 := time.Now().Format("20060102")

	err = 数据库.Transaction(func(tx *gorm.DB) error {
		// 查询或创建用户记录
		info.luckyWheelUser, _ = service.NewLuckyWheelUser(c, tx).Info(appInfo.AppId, 在线信息.Uid)
		if info.luckyWheelUser.Id == 0 {
			// 创建用户记录
			info.luckyWheelUser = dbm.DB_LuckyWheelUser{
				AppId:         appInfo.AppId,
				UserId:        在线信息.Uid,
				CreateTime:    time.Now().Unix(),
				UpdateTime:    time.Now().Unix(),
				RemainCount:   0,
				DailyFreeDate: 局_今日,
				DailyFreeUsed: 0,
			}
			if _, e := service.NewLuckyWheelUser(c, tx).Create(&info.luckyWheelUser); e != nil {
				return e
			}
		}

		// 自动领取每日免费次数
		局_来源 := 1 //默认每日免费
		if info.luckyWheelUser.DailyFreeDate != 局_今日 {
			// 跨天重置
			info.luckyWheelUser.DailyFreeDate = 局_今日
			info.luckyWheelUser.DailyFreeUsed = 0
		}
		if info.luckyWheelUser.DailyFreeUsed < info.luckyWheelInfo.DailyFreeCount && info.luckyWheelInfo.DailyFreeCount > 0 {
			// 领取免费次数
			info.luckyWheelUser.RemainCount += 1
			info.luckyWheelUser.DailyFreeUsed += 1
		} else {
			局_来源 = 2 //拉新奖励
		}

		// 行锁重新查
		if e := tx.Model(dbm.DB_LuckyWheelUser{}).Clauses(clause.Locking{Strength: "UPDATE"}).Where("Id = ?", info.luckyWheelUser.Id).First(&info.luckyWheelUser).Error; e != nil {
			return e
		}

		if info.luckyWheelUser.RemainCount <= 0 {
			return errors.New("抽奖次数不足")
		}

		// 扣减次数
		info.luckyWheelUser.RemainCount -= 1
		if _, e := service.NewLuckyWheelUser(c, tx).UpdateMap([]int{info.luckyWheelUser.Id}, map[string]interface{}{
			"remainCount":    info.luckyWheelUser.RemainCount,
			"dailyFreeDate":  info.luckyWheelUser.DailyFreeDate,
			"dailyFreeUsed":  info.luckyWheelUser.DailyFreeUsed,
			"updateTime":     time.Now().Unix(),
		}); e != nil {
			return e
		}

		// 计算中奖结果
		中奖索引, 中奖奖品 = 计算中奖(info.prizeList)

		// 如果中奖(kaClassId > 0),执行卡类直冲
		if 中奖奖品.KaClassId > 0 {
			c.Set("tx", tx)
			defer delete(c.Keys, "tx")
			if e := ka.L_ka.K卡类直冲_事务(c, 中奖奖品.KaClassId, 在线信息.Uid); e != nil {
				return e
			}
		}

		// 写入抽奖记录
		_, e := service.NewLuckyWheelLog(c, tx).Create(&dbm.DB_LuckyWheelLog{
			AppId:       appInfo.AppId,
			UserId:      在线信息.Uid,
			CreateTime:  time.Now().Unix(),
			KaClassId:   中奖奖品.KaClassId,
			KaClassName: 中奖奖品.Name,
			Source:      局_来源,
		})
		return e
	})

	return
}

// L领取每日免费 领取每日免费抽奖次数(用户进入页面时调用)
func L领取每日免费(c *gin.Context, 数据库 *gorm.DB, appInfo dbm.DB_AppInfo, 在线信息 dbm.DB_LinksToken, 活动配置 dbm.DB_AppPromotionConfig) (err error) {
	var info = struct {
		luckyWheelInfo dbm.DB_LuckyWheelInfo
		luckyWheelUser dbm.DB_LuckyWheelUser
	}{}

	info.luckyWheelInfo, err = service.NewLuckyWheelInfo(c, 数据库).Info(活动配置.TypeAssociatedId)
	if err != nil {
		return
	}
	if info.luckyWheelInfo.DailyFreeCount <= 0 {
		return //未开启每日免费
	}

	局_今日 := time.Now().Format("20060102")

	info.luckyWheelUser, _ = service.NewLuckyWheelUser(c, 数据库).Info(appInfo.AppId, 在线信息.Uid)
	if info.luckyWheelUser.Id == 0 {
		// 创建用户记录并领取
		info.luckyWheelUser = dbm.DB_LuckyWheelUser{
			AppId:         appInfo.AppId,
			UserId:        在线信息.Uid,
			CreateTime:    time.Now().Unix(),
			UpdateTime:    time.Now().Unix(),
			RemainCount:   info.luckyWheelInfo.DailyFreeCount,
			DailyFreeDate: 局_今日,
			DailyFreeUsed: info.luckyWheelInfo.DailyFreeCount,
		}
		_, err = service.NewLuckyWheelUser(c, 数据库).Create(&info.luckyWheelUser)
		return
	}

	// 跨天重置并领取
	if info.luckyWheelUser.DailyFreeDate != 局_今日 {
		局_今日剩余 := info.luckyWheelInfo.DailyFreeCount - 0
		info.luckyWheelUser.RemainCount += 局_今日剩余
		info.luckyWheelUser.DailyFreeDate = 局_今日
		info.luckyWheelUser.DailyFreeUsed = info.luckyWheelInfo.DailyFreeCount
		_, err = service.NewLuckyWheelUser(c, 数据库).UpdateMap([]int{info.luckyWheelUser.Id}, map[string]interface{}{
			"remainCount":   info.luckyWheelUser.RemainCount,
			"dailyFreeDate":  info.luckyWheelUser.DailyFreeDate,
			"dailyFreeUsed":  info.luckyWheelUser.DailyFreeUsed,
			"updateTime":     time.Now().Unix(),
		})
	}
	return
}

// 计算中奖 概率计算,概率总和不超过10000,不足部分为"谢谢参与"(未中奖)
func 计算中奖(prizeList []PrizeItem) (索引 int, 中奖奖品 PrizeItem) {
	rand.Seed(time.Now().UnixNano())
	局_随机数 := rand.Intn(10000)
	局_累计 := 0
	for i, item := range prizeList {
		局_累计 += item.Probability
		if 局_随机数 < 局_累计 {
			索引 = i
			中奖奖品 = item
			return
		}
	}
	// 落在概率总和之外,返回"谢谢参与"(未中奖)
	索引 = -1
	中奖奖品 = PrizeItem{KaClassId: 0, Probability: 10000 - 局_累计, Name: "谢谢参与"}
	return
}
