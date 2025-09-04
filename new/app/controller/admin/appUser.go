package controller

import (
	. "EFunc/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/Service/Ser_LinkUser"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/controller/Common/response"

	"server/new/app/service"
	DB "server/structs/db"
	utils2 "server/utils"
	"strconv"
	"strings"
	"time"
)

type AppUser struct {
	Common.Common
}

func NewAppUserController() *AppUser {
	return &AppUser{}
}

// 批量添加用户
func (C *AppUser) BatchAddUser(c *gin.Context) {
	//{"AppId":2,"Type":2,"Size":10,"Page":1,"Status":1,"keywords":"1"}
	var 请求 struct {
		AppId int    `json:"AppId"`
		Note  string `json:"Note"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	db := *global.GVA_DB
	var err error
	var info struct {
		AppInfo   DB.DB_AppInfo
		用户数组      []string
		数组ka      []DB.DB_Ka
		数组User    []DB.DB_User
		数组AppUser []DB.DB_AppUser
	}
	局_制卡人 := Ser_LinkUser.Token取Name(c.Request.Header.Get("Token"))
	局_时间戳 := time.Now().Unix()
	info.AppInfo, err = service.NewAppInfo(c, &db).Info(请求.AppId)
	if err != nil {
		response.FailWithMessage(c, "AppId不存在")
		return
	}
	info.用户数组 = strings.Split(请求.Note, "\n")
	info.数组ka = make([]DB.DB_Ka, len(info.用户数组))
	info.数组User = make([]DB.DB_User, len(info.用户数组))
	info.数组AppUser = make([]DB.DB_AppUser, len(info.用户数组))

	//账号|密码|到期时间|积分|绑定信息
	//卡号|到期时间|积分|绑定信息
	//aaaaaadwad|2023-10-01 12:00:00|66|绑定信息

	for i, v := range info.用户数组 {
		局_临时数组 := strings.Split(v, "|")
		if len(局_临时数组) != S三元(info.AppInfo.AppType < 3, 5, 4) {
			response.FailWithMessage(c, "第"+strconv.Itoa(i+1)+"行数据格式错误,|符号数量不正确")
			return
		}
		var 局_账号, 局_密码 string
		var 局_积分 float64
		var 局_VipTime int64

		if info.AppInfo.AppType < 3 {
			局_账号 = strings.TrimSpace(局_临时数组[0])
			局_密码 = strings.TrimSpace(局_临时数组[1])
			局_积分, _ = strconv.ParseFloat(局_临时数组[3], 64)
		} else {
			局_账号 = strings.TrimSpace(局_临时数组[0])
			局_密码 = ""
			局_积分, _ = strconv.ParseFloat(局_临时数组[2], 64)
		}

		if len(局_临时数组[len(局_临时数组)-3]) == 10 {
			局_VipTime, _ = strconv.ParseInt(局_临时数组[len(局_临时数组)-3], 10, 64)
		} else {
			t, err2 := time.Parse("2006-01-02 15:04:05", 局_临时数组[len(局_临时数组)-3])
			if err2 != nil {
				response.FailWithMessage(c, "第"+strconv.Itoa(i+1)+"行数据格式错误,|到期时间格式错误,")
				return
			}
			局_VipTime = t.Unix()
		}

		info.数组AppUser[i] = DB.DB_AppUser{
			Uid:          0,
			Status:       1,
			Key:          局_临时数组[len(局_临时数组)-1],
			VipTime:      局_VipTime,
			VipNumber:    局_积分,
			MaxOnline:    1,
			RegisterTime: 局_时间戳,
		}

		if info.AppInfo.AppType < 3 {
			info.数组User[i] = DB.DB_User{
				User:          局_账号,
				PassWord:      S三元(len(局_临时数组[1]) == 32, 局_密码, utils2.Md5String(局_密码)),
				SuperPassWord: utils2.Md5String(局_密码),
				RegisterTime:  局_时间戳,
				RegisterIp:    "",
			}
		} else {
			info.数组ka[i] = DB.DB_Ka{
				AppId:        请求.AppId,
				Name:         局_账号,
				Status:       1,
				RegisterUser: 局_制卡人,
				RegisterTime: 局_时间戳,
				AdminNote:    "批量导入用户关联卡号",
				AgentNote:    "",
				VipTime:      0,
				InviteCount:  0,
				RMb:          0,
				VipNumber:    0,
				Money:        0,
				AgentMoney:   0,
				UserClassId:  0,
				NoUserClass:  2,
				KaType:       1,
				MaxOnline:    1,
				Num:          1,
				NumMax:       1,
				User:         局_临时数组[0],
				UserTime:     D到文本(局_时间戳),
				UseTime:      局_时间戳,
				InviteUser:   "",
				EndTime:      9999999999,
			}
		}
	}
	//执行批量插入 事务处理
	err = db.Transaction(func(tx *gorm.DB) error {
		if info.AppInfo.AppType < 3 {
			err = tx.Model(DB.DB_User{}).CreateInBatches(&info.数组User, len(info.数组User)).Error
			if err != nil {
				return err
			}
			for i, v := range info.数组User {
				info.数组AppUser[i].Uid = v.Id
			}

		} else {
			err = tx.Model(DB.DB_Ka{}).CreateInBatches(&info.数组ka, len(info.数组ka)).Error
			if err != nil {
				return err
			}
			for i, v := range info.数组ka {
				info.数组AppUser[i].Uid = v.Id
			}
		}
		err = tx.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(请求.AppId)).CreateInBatches(&info.数组AppUser, len(info.数组AppUser)).Error
		return err
	})

	if err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}
	response.OkWithMessage(c, "添加成功,总数:"+strconv.Itoa(len(info.数组AppUser)))
	return
}
