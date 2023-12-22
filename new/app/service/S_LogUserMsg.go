package service

import (
	"gorm.io/gorm"
	DB "server/structs/db"
)

type S_LogUserMsg struct {
}

func (s *S_LogUserMsg) S删除重复消息(tx *gorm.DB) error {
	/*	DELETE from db_Log_UserMsg where id not IN (
		SELECT
		    min(id) id
		FROM
			db_Log_UserMsg
		GROUP BY
			Note
		)
	*/
	var ids []int
	//不能删除子级的子查询表,所以需要两步 先聚合id,再删除
	err := tx.Raw("SELECT min(id) id FROM db_Log_UserMsg GROUP BY Note").Scan(&ids).Error
	if err != nil {
		return err
	}
	err = tx.Debug().Model(DB.DB_LogUserMsg{}).Where("id not IN ?", ids).Delete("").Error
	return err
}
