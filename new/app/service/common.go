package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/global"
	"server/new/app/models/request"
	"server/utils"
)

// 分批执行IN查询删除，避免MySQL占位符超限(65535)
func batchDeleteIN(db *gorm.DB, model interface{}, tableName string, field string, ids []int) (int64, error) {
	var total int64
	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}
		d := db.Model(model)
		if tableName != "" {
			d = d.Table(tableName)
		}
		tx := d.Where(field+" IN ?", ids[i:end]).Delete("")
		if tx.Error != nil {
			return total, tx.Error
		}
		total += tx.RowsAffected
	}
	return total, nil
}

// 分批执行IN查询删除(int64版本)
func batchDeleteIN64(db *gorm.DB, model interface{}, tableName string, field string, ids []int64) (int64, error) {
	var total int64
	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}
		d := db.Model(model)
		if tableName != "" {
			d = d.Table(tableName)
		}
		tx := d.Where(field+" IN ?", ids[i:end]).Delete("")
		if tx.Error != nil {
			return total, tx.Error
		}
		total += tx.RowsAffected
	}
	return total, nil
}

// 分批执行IN查询，返回结果集(避免占位符超限)
func batchFindIN[T any](db *gorm.DB, model interface{}, tableName string, field string, ids []int) ([]T, error) {
	var result []T
	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}
		var batch []T
		d := db.Model(model)
		if tableName != "" {
			d = d.Table(tableName)
		}
		tx := d.Where(field+" IN ?", ids[i:end]).Find(&batch)
		if tx.Error != nil {
			return result, tx.Error
		}
		result = append(result, batch...)
	}
	return result, nil
}

// 分批执行IN查询(int64版本)
func batchFindIN64[T any](db *gorm.DB, model interface{}, tableName string, field string, ids []int64) ([]T, error) {
	var result []T
	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}
		var batch []T
		d := db.Model(model)
		if tableName != "" {
			d = d.Table(tableName)
		}
		tx := d.Where(field+" IN ?", ids[i:end]).Find(&batch)
		if tx.Error != nil {
			return result, tx.Error
		}
		result = append(result, batch...)
	}
	return result, nil
}

// 分批执行IN更新，避免MySQL占位符超限
func batchUpdateIN(db *gorm.DB, model interface{}, tableName string, field string, ids []int, data map[string]interface{}) (int64, error) {
	var total int64
	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}
		d := db.Model(model)
		if tableName != "" {
			d = d.Table(tableName)
		}
		tx := d.Where(field+" IN ?", ids[i:end]).Updates(data)
		if tx.Error != nil {
			return total, tx.Error
		}
		total += tx.RowsAffected
	}
	return total, nil
}

// 使用泛型定义基础服务
type BaseService[T any] struct {
	db *gorm.DB
	c  *gin.Context
}

func NewBaseService[T any](c *gin.Context, db *gorm.DB) *BaseService[T] {
	return &BaseService[T]{
		db: db,
		c:  c,
	}
}

// 通用创建方法（使用泛型）
func (s *BaseService[T]) Create(record *T) (int64, error) {
	tx := s.db.Create(record)
	return tx.RowsAffected, tx.Error
} // 批量创建
func (s *BaseService[T]) BatchCreate(records *[]T) error {
	return s.db.CreateInBatches(records, 100).Error
}

// 删除 - 支持ID或ID数组
func (s *BaseService[T]) Delete(id interface{}) (int64, error) {
	switch v := id.(type) {
	case int:
		tx := s.db.Delete(new(T), v)
		return tx.RowsAffected, tx.Error
	case []int:
		var total int64
		for i := 0; i < len(v); i += batchSize {
			end := i + batchSize
			if end > len(v) {
				end = len(v)
			}
			tx := s.db.Delete(new(T), v[i:end])
			if tx.Error != nil {
				return total, tx.Error
			}
			total += tx.RowsAffected
		}
		return total, nil
	default:
		return 0, errors.New("不支持的数据类型")
	}
}

// 删除 支持 数组,和id
func (s *BaseService[T]) DeleteWhere(where map[string]interface{}) (影响行数 int64, error error) {

	tx := s.db.Model(new(T)).Where(where).Delete("")

	return tx.RowsAffected, tx.Error
}

// 改
func (s *BaseService[T]) UpdateMap(Id []int, data map[string]interface{}) (row int64, err error) {
	var total int64
	for i := 0; i < len(Id); i += batchSize {
		end := i + batchSize
		if end > len(Id) {
			end = len(Id)
		}
		tx := s.db.Model(new(T)).Where("id in ?", Id[i:end]).Updates(data)
		if tx.Error != nil {
			return total, tx.Error
		}
		total += tx.RowsAffected
	}
	return total, nil
}

// 改
func (s *BaseService[T]) UpdateWhere(where map[string]interface{}, data map[string]interface{}) (int64, error) {
	tx := s.db.Model(new(T)).Where(where).Updates(data)
	return tx.RowsAffected, tx.Error
}

// 查
func (s *BaseService[T]) Info(id int) (info T, err error) {
	tx := s.db.Model(new(T)).Where("Id = ?", id).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 查
func (s *BaseService[T]) Infos(where map[string]interface{}) (info []T, err error) {
	tx := s.db.Model(new(T)).Where(where).Find(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 优化查询链式操作
func (s *BaseService[T]) GetList(请求 request.List) (int64, []T, error) {
	// 创建查询构建器
	db := s.db.Model(new(T))
	if 请求.Page == 0 {
		请求.Page = 1
	}

	// 关键字搜索
	if 请求.Keywords != "" && 请求.Type == 1 {
		db = db.Where("Id = ?", 请求.Keywords)
	}

	// 优化计数逻辑
	var count int64
	if 请求.Count > 0 && 请求.Count <= 500000 {
		count = 请求.Count
	} else {
		if err := db.Count(&count).Error; err != nil {
			return 0, nil, err
		}
	}

	// 排序处理
	order := "Id DESC"
	if 请求.Order == 1 {
		order = "Id ASC"
	}

	// 分页查询
	var results []T
	err := db.Order(order).
		Limit(请求.Size).
		Offset((请求.Page - 1) * 请求.Size).
		Find(&results).Error

	if err != nil {
		global.GVA_LOG.Error(utils.Q取包名结构体方法(s) + ":" + err.Error())
	}

	return count, results, err
}
