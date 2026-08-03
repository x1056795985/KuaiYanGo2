package appPromotionConfig

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"server/app/models/constant"
	dbm "server/app/models/db"
	"server/app/service"
)

// C活动_创建 在事务中创建活动及其关联配置。
func C活动_创建(c *gin.Context, 数据库 *gorm.DB, appInfo dbm.DB_AppInfo, name string, startTime int64, endTime int64, promotionType int) error {
	return 数据库.Transaction(func(tx *gorm.DB) error {
		var 局_关联配置id int
		switch promotionType {
		default:
			return errors.New("活动类型错误")
		case constant.H活动类型_签到:
			局_签到配置 := dbm.DB_CheckInInfo{
				CreateTime:       time.Now().Unix(),
				UpdateTime:       time.Now().Unix(),
				ShareGivePoints:  10,
				InviteGivePoints: 88,
				CardClassList:    "[]",
			}
			if _, 局_错误 := service.NewCheckInInfo(c, tx).Create(&局_签到配置); 局_错误 != nil {
				return 局_错误
			}
			局_关联配置id = 局_签到配置.Id
		case constant.H活动类型_cps:
			if appInfo.AppType == 3 || appInfo.AppType == 4 {
				return errors.New("卡号模式应用暂不支持该活动")
			}
			局_cps配置 := dbm.DB_CpsInfo{
				CreateTime:         time.Now().Unix(),
				UpdateTime:         time.Now().Unix(),
				BronzeThreshold:    0,
				BronzeKickback:     10,
				SilverThreshold:    10,
				SilverKickback:     20,
				GoldMedalThreshold: 20,
				GoldMedalKickback:  30,
				GrandsonKickback:   2,
				BindingDay:         180,
			}
			if _, 局_错误 := service.NewCpsInfo(c, tx).Create(&局_cps配置); 局_错误 != nil {
				return 局_错误
			}
			局_关联配置id = 局_cps配置.Id
		case constant.H活动类型_大转盘:
			局_大转盘配置 := dbm.DB_LuckyWheelInfo{
				CreateTime:      time.Now().Unix(),
				UpdateTime:      time.Now().Unix(),
				DailyFreeCount:  1,
				InviteGiveCount: 1,
				PrizeList:       "[]",
			}
			if _, 局_错误 := service.NewLuckyWheelInfo(c, tx).Create(&局_大转盘配置); 局_错误 != nil {
				return 局_错误
			}
			局_关联配置id = 局_大转盘配置.Id
		}

		_, 局_错误 := service.NewAppPromotionConfig(c, tx).Create(&dbm.DB_AppPromotionConfig{
			Name:             name,
			AppId:            appInfo.AppId,
			CreateTime:       time.Now().Unix(),
			UpdateTime:       time.Now().Unix(),
			StartTime:        startTime,
			EndTime:          endTime,
			PromotionType:    promotionType,
			TypeAssociatedId: 局_关联配置id,
		})
		return 局_错误
	})
}

// S活动_删除 在事务中删除活动及其关联配置。
func S活动_删除(数据库 *gorm.DB, ids []int) (影响行数 int64, err error) {
	err = 数据库.Transaction(func(tx *gorm.DB) error {
		var 局_配置 []dbm.DB_AppPromotionConfig
		if 局_错误 := tx.Where("id IN (?)", ids).Find(&局_配置).Error; 局_错误 != nil {
			return 局_错误
		}
		var 局_cpsIds []int
		var 局_luckyWheelIds []int
		for _, 局_单项配置 := range 局_配置 {
			if 局_单项配置.PromotionType == constant.H活动类型_cps {
				局_cpsIds = append(局_cpsIds, 局_单项配置.TypeAssociatedId)
			} else if 局_单项配置.PromotionType == constant.H活动类型_大转盘 {
				局_luckyWheelIds = append(局_luckyWheelIds, 局_单项配置.TypeAssociatedId)
			}
		}
		if len(局_cpsIds) > 0 {
			if 局_错误 := tx.Where("id IN (?)", 局_cpsIds).Delete(&dbm.DB_CpsInfo{}).Error; 局_错误 != nil {
				return 局_错误
			}
		}
		if len(局_luckyWheelIds) > 0 {
			if 局_错误 := tx.Where("id IN (?)", 局_luckyWheelIds).Delete(&dbm.DB_LuckyWheelInfo{}).Error; 局_错误 != nil {
				return 局_错误
			}
		}
		局_结果 := tx.Where("id IN (?)", ids).Delete(&dbm.DB_AppPromotionConfig{})
		影响行数 = 局_结果.RowsAffected
		return 局_结果.Error
	})
	return
}
