package service

import (
	"errors"
	"gorm.io/gorm"
	"server/new/app/models/db"
)

type S_Setting struct {
}

func (s *S_Setting) Info(tx *gorm.DB, key string) (string, error) {
	var value = ""
	err := tx.Model(db.DB_Setting{}).Select("ItemValue").Where("ItemKey =?", key).First(&value).Error
	return value, err
}
func (s *S_Setting) InfoBatch(tx *gorm.DB, key []string) (map[string]string, error) {
	var value map[string]string
	err := tx.Model(db.DB_Setting{}).Select("ItemValue").Where("ItemKey IN ?", key).Find(&value).Error
	return value, err
}
func (s *S_Setting) Update(tx *gorm.DB, key, value string) error {
	//tx 必须复制一次才可以,不然值是报错的,后面使用都是报错的 也就是说参数不能传指针必须传值,
	var a db.DB_Setting
	err := tx.Model(db.DB_Setting{}).Where("ItemKey = ?", key).First(&a).Error

	if err == nil {
		err = tx.Model(db.DB_Setting{}).Where("ItemKey = ?", key).Update("ItemValue", &value).Error
	} else {
		err = tx.Model(db.DB_Setting{}).Create(db.DB_Setting{ItemKey: key, ItemValue: value}).Error
	}

	return err
}
func (s *S_Setting) Delete(tx *gorm.DB, ItemKey interface{}) (影响行数 int64, error error) {
	switch k := ItemKey.(type) {
	case int64:
		tx2 := tx.Model(db.DB_Setting{}).Where("ItemKey = ?", k).Delete("")
		return tx2.RowsAffected, tx2.Error
	case []int64:
		tx2 := tx.Model(db.DB_Setting{}).Where("ItemKey IN (?)", k).Delete("")
		return tx2.RowsAffected, tx2.Error
	default:
		return 0, errors.New("错误的数据")
	}

}
