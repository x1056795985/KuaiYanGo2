package userClass

import (
	"server/global"
	DB "server/structs/db"
	"strconv"
)

var L_userClass userClass

func init() {
	L_userClass = userClass{}

}

type userClass struct {
}

func (j *userClass) UserClass取map列表String(Appid int) map[string]string {

	var DB_UserClass = []DB.DB_UserClass{}
	tx := *global.GVA_DB
	_ = tx.Model(DB.DB_UserClass{}).Select("Id", "Name").Where("Appid=?", Appid).Find(&DB_UserClass).Error
	var AppName = make(map[string]string, len(DB_UserClass))
	//吧 id 和 app名字 放入map
	for 索引 := range DB_UserClass {
		AppName[strconv.Itoa(int(DB_UserClass[索引].Id))] = DB_UserClass[索引].Name
	}
	return AppName
}
