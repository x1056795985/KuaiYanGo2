package publicData

import (
	"EFunc/utils"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"server/global"
	DB "server/structs/db"
	"strings"
)

var L_publicData publicData

func init() {
	L_publicData = publicData{}
}

type publicData struct {
}

func (j *publicData) Name是否存在(c *gin.Context, AppId int, Name string) bool {
	var Count int64
	db := *global.GVA_DB
	db.Model(DB.DB_PublicData{}).Select("1").Where("Name=?", Name).Where("AppId=?", AppId).Take(&Count)
	return Count > 0
}

// 会正确处理队列
func (j *publicData) Z置值(c *gin.Context, Appid int, 变量名, 变量值 string) (err error) {
	db := *global.GVA_DB
	var 局_云变量数据 DB.DB_PublicData
	err = db.Model(DB.DB_PublicData{}).Where("AppId=?", Appid).Where("Name=?", 变量名).First(&局_云变量数据).Error
	if err != nil {
		err = errors.Join(err, errors.New("变量不存在"))
		return
	}
	//队列类型的单独处理,加锁读取,避免队列数据被修改
	if 局_云变量数据.Type <= 3 {
		err = db.Model(DB.DB_PublicData{}).
			Where("AppId=?", Appid).Where("Name=?", 变量名).
			Update("Value", 变量值).Error
		if err != nil {
			return
		}
	} else if 局_云变量数据.Type == 4 {
		err = db.Transaction(func(tx *gorm.DB) error {
			err = tx.Model(DB.DB_PublicData{}).
				Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("AppId=?", Appid).Where("Name=?", 变量名).
				First(&局_云变量数据).Error //加锁再查一次
			if err != nil {
				return err
			}
			if 局_云变量数据.Value != "" {
				局_云变量数据.Value += "\n"
			}
			局_云变量数据.Value += 变量值
			err = tx.Model(DB.DB_PublicData{}).
				Where("AppId=?", Appid).Where("Name=?", 变量名).
				Update("Value", 局_云变量数据.Value).Error //更新变量
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return
		}
	}
	return err
}

func (j *publicData) Z置值_原值(c *gin.Context, PublicData DB.DB_PublicData) error {
	db := *global.GVA_DB
	return db.Model(DB.DB_PublicData{}).Select("Value", "IsVip", "Note", "Time", "Sort").Omit("Type", "AppId", "Name").Where("AppId=?", PublicData.AppId).Where("Name=?", PublicData.Name).Updates(PublicData).Error
}

func (j *publicData) Q取值(c *gin.Context, Appid int, Name string) string {
	var value DB.DB_PublicData
	value, _ = j.Q取值2(c, Appid, Name)
	return value.Value
}
func (j *publicData) Q取值2(c *gin.Context, Appid int, 变量名 string) (返回 DB.DB_PublicData, err error) {
	db := *global.GVA_DB
	var 局_云变量数据 DB.DB_PublicData
	err = db.Model(DB.DB_PublicData{}).Where("AppId=?", Appid).Where("Name=?", 变量名).First(&局_云变量数据).Error
	if err != nil {
		err = errors.Join(err, errors.New("变量不存在"))
		return
	}
	//队列类型的单独处理,加锁读取,避免队列数据被修改
	if 局_云变量数据.Type == 4 {
		err = db.Transaction(func(tx *gorm.DB) error {
			err = tx.Model(DB.DB_PublicData{}).
				Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("AppId=?", 1).Where("Name=?", 变量名).
				First(&局_云变量数据).Error //加锁再查一次
			if err != nil {
				return err
			}
			// 使用 SplitN 获取第一行
			局_临时数组 := strings.SplitN(局_云变量数据.Value, "\n", 2)
			if len(局_临时数组) == 2 {
				局_云变量数据.Value = 局_临时数组[1]
			} else {
				局_云变量数据.Value = ""
			}
			err = tx.Model(DB.DB_PublicData{}).
				Where("AppId=?", 1).Where("Name=?", 变量名).
				Update("Value", 局_云变量数据.Value).Error
			if err != nil {
				return err
			}
			switch len(局_临时数组) {
			default:
				//只要不是0 都使用0作为返回变量,
				局_云变量数据.Value = 局_临时数组[0]
			case 0:
				局_云变量数据.Value = ""
			}

			return nil
		})
		if err != nil {
			return
		}

	}

	返回 = 局_云变量数据
	return
}

func (j *publicData) P批量修改IsVip(c *gin.Context, AppId int, Name []string, IsVip int) error {
	db := *global.GVA_DB
	return db.Model(DB.DB_PublicData{}).Where("AppId=?", AppId).Where("Name in ?", Name).Update("IsVip", IsVip).Error
}

func (j *publicData) C创建(c *gin.Context, PublicData DB.DB_PublicData) error {
	db := *global.GVA_DB
	err := db.Model(DB.DB_PublicData{}).Create(&PublicData).Error
	return err
}
func (j *publicData) Q取队列长度(c *gin.Context, Appid int, 变量名 string) (返回 int, err error) {
	var 局_云变量数据 DB.DB_PublicData
	db := *global.GVA_DB
	err = db.Model(DB.DB_PublicData{}).Where("AppId=?", Appid).Where("Name=?", 变量名).First(&局_云变量数据).Error
	if err != nil {
		err = errors.Join(err, errors.New("变量不存在"))
		return
	}
	返回 = utils.W文本_取行数(局_云变量数据.Value)
	return
}
