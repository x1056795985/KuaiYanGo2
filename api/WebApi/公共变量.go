package WebApi

import (
	"github.com/gin-gonic/gin"
	"github.com/valyala/fastjson"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"server/Service/Ser_PublicData"
	"server/global"
	"server/structs/Http/response"
	DB "server/structs/db"
	"strings"
)

func Q取公共变量(c *gin.Context) {
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"GetPublicData","Name":"会员数据a"}
	局_变量名 := string(请求json.GetStringBytes("Name"))
	局_云变量数据, err := Ser_PublicData.P取值2(1, 局_变量名) //1>所以有软件公共读
	if err != nil {
		response.FailWithMessage("变量不存在", c)
		return
	}
	//队列类型的单独处理,加锁读取,避免队列数据被修改
	if 局_云变量数据.Type == 4 {
		db := *global.GVA_DB
		err = db.Transaction(func(tx *gorm.DB) error {
			err = tx.Model(DB.DB_PublicData{}).
				Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("AppId=?", 1).Where("Name=?", 局_变量名).
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
				Where("AppId=?", 1).Where("Name=?", 局_变量名).
				Update("Value", 局_云变量数据.Value).Error //加锁再查一次
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
			response.FailWithMessage(err.Error(), c)
			return
		}

	}
	response.OkWithDetailed(局_云变量数据.Value, "获取成功", c)
	return
}
func Z置公共变量(c *gin.Context) {
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"GetPublicData","Name":"会员数据a","Value":"aaaaa"}
	局_变量名 := string(请求json.GetStringBytes("Name"))
	局_变量值 := string(请求json.GetStringBytes("Value"))

	局_云变量数据, err := Ser_PublicData.P取值2(1, 局_变量名) //1>所以有软件公共读
	if err != nil || 局_云变量数据.Type > 4 {           //只允许变量  不允许读取云函数
		response.FailWithMessage("变量不存在", c)
		return
	}
	//队列类型的单独处理,加锁读取,避免队列数据被修改
	db := *global.GVA_DB
	if 局_云变量数据.Type <= 3 {
		err = db.Model(DB.DB_PublicData{}).
			Where("AppId=?", 1).Where("Name=?", 局_变量名).
			Update("Value", 局_变量值).Error //加锁再查一次
		if err != nil {
			response.FailWithMessage(err.Error(), c)
			return
		}
	} else if 局_云变量数据.Type == 4 {
		err = db.Transaction(func(tx *gorm.DB) error {
			err = tx.Model(DB.DB_PublicData{}).
				Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("AppId=?", 1).Where("Name=?", 局_变量名).
				First(&局_云变量数据).Error //加锁再查一次
			if err != nil {
				return err
			}
			if 局_云变量数据.Value != "" {
				局_云变量数据.Value += "\n"
			}
			局_云变量数据.Value += 局_变量值
			err = tx.Model(DB.DB_PublicData{}).
				Where("AppId=?", 1).Where("Name=?", 局_变量名).
				Update("Value", 局_云变量数据.Value).Error //加锁再查一次
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			response.FailWithMessage(err.Error(), c)
			return
		}
	}

	response.Ok(c)
	return
}
