package publicJs

import (
	. "EFunc/utils"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"server/app/global"
	dbm "server/app/models/db"
)

var L_publicJs publicJs

func init() {
	L_publicJs = publicJs{}
}

type publicJs struct {
}

// Z置值2 修改公共JS函数(涉及文件IO+数据库+缓存三重操作)
func (j *publicJs) Z置值2(c *gin.Context, PublicJs dbm.DB_PublicJs) error {
	//注意宝塔写文件 文件会在 /www/server/panel 文件夹
	err := W文件_保存(公共JS_脚本文件路径("/云函数/"+PublicJs.Name+".js"), PublicJs.Value)
	if err != nil {
		return err
	}
	PublicJs.Value = "/云函数/" + PublicJs.Name + ".js"

	m := map[string]interface{}{}
	m["AppId"] = PublicJs.AppId
	m["Name"] = PublicJs.Name
	m["Value"] = PublicJs.Value
	m["IsVip"] = PublicJs.IsVip
	m["Note"] = PublicJs.Note
	err = global.GVA_DB.Model(dbm.DB_PublicJs{}).Where("Id=?", PublicJs.Id).Updates(&m).Error
	if err == nil && global.H缓存 != nil { //删除缓存
		global.H缓存.Delete(公共JS_脚本文件路径(PublicJs.Value))
	}
	return err
}

// C创建 创建公共JS函数(涉及文件IO+数据库)
func (j *publicJs) C创建(c *gin.Context, PublicJs dbm.DB_PublicJs) error {
	//注意宝塔写文件 文件会在 /www/server/panel 文件夹
	err := W文件_保存(公共JS_脚本文件路径("/云函数/"+PublicJs.Name+".js"), PublicJs.Value)
	if err != nil {
		return errors.New("Js写入文件失败:" + err.Error())
	}
	PublicJs.Value = "/云函数/" + PublicJs.Name + ".js"
	err = global.GVA_DB.Model(dbm.DB_PublicJs{}).Create(&PublicJs).Error
	return err
}

// P取值2 按Appid和Name取公共JS函数(涉及文件IO+缓存)
func (j *publicJs) P取值2(c *gin.Context, Appid int, Name string) (dbm.DB_PublicJs, error) {
	var 局_PublicJs dbm.DB_PublicJs
	err := global.GVA_DB.Model(dbm.DB_PublicJs{}).Where("AppId=?", Appid).Where("Name=?", Name).First(&局_PublicJs).Error

	if err != nil {
		return 局_PublicJs, errors.New("[" + Name + "],Hook函数不存在")
	}

	局_文件路径 := 公共JS_脚本文件路径(局_PublicJs.Value)
	if global.H缓存 != nil {
		if 局_缓存值, ok := global.H缓存.Get(局_文件路径); ok {
			if 局_文本, 局_类型正确 := 局_缓存值.(string); 局_类型正确 {
				局_PublicJs.Value = 局_文本
				return 局_PublicJs, nil
			}
			global.H缓存.Delete(局_文件路径)
		}
	}
	if !W文件_是否存在(局_文件路径) {
		return 局_PublicJs, errors.New(Name + ".js文件读取失败可能被删除,请重新编辑公共函数")
	}
	局_PublicJs.Value = string(W文件_读入文件(局_文件路径))
	if global.H缓存 != nil {
		global.H缓存.Set(局_文件路径, 局_PublicJs.Value, time.Hour*720)
	}

	return 局_PublicJs, nil
}

func 公共JS_脚本文件路径(value string) string {
	局_相对路径 := strings.TrimLeft(filepath.FromSlash(value), `/\`)
	return filepath.Join(global.GVA_CONFIG.Q取运行目录, 局_相对路径)
}
