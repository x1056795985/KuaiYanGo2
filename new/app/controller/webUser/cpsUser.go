package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/controller/Common/response"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	DB "server/structs/db"
	"strconv"
	"time"
)

type CpsUser struct {
	Common.Common
}

func NewCpsUserController() *CpsUser {
	return &CpsUser{}
}

func (C *CpsUser) Info(c *gin.Context) {
	var err error
	var info = struct {
		appInfo  DB.DB_AppInfo
		likeInfo DB.DB_LinksToken
		cpsUser  dbm.DB_CpsUser
		cpsCode  dbm.DB_CpsCode
	}{}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	//查询是否拥有邀请人   如果已设置过,需要删除,因为有唯一索引
	tx := *global.GVA_DB
	info.cpsUser, err = service.NewCpsUser(c, &tx).Info(info.appInfo.AppId, info.likeInfo.Uid)
	//判断是否存在,如果不存在,插入默认数据
	if err != nil && err.Error() == "record not found" {
		tx = *global.GVA_DB
		info.cpsUser.UserId = info.likeInfo.Uid
		info.cpsUser.AppId = info.appInfo.AppId
		info.cpsUser.CreatedAt = time.Now().Unix()
		info.cpsUser.UpdatedAt = info.cpsUser.CreatedAt
		_, err = service.NewCpsUser(c, &tx).Create(&info.cpsUser)
	}

	info.cpsCode, err = service.NewCpsCode(c, &tx).InfoUserId(info.likeInfo.Uid)
	//判断是否存在,如果不存在,插入默认数据
	if err != nil && err.Error() == "record not found" {
		tx = *global.GVA_DB
		info.cpsCode.UserId = info.likeInfo.Uid
		info.cpsCode.CpsCode = GetInvCodeByUIDUniqueNew(strconv.Itoa(info.likeInfo.Uid + 1000000))
		info.cpsCode.CreatedAt = time.Now().Unix()
		info.cpsCode.UpdatedAt = info.cpsUser.CreatedAt
		_, err = service.NewCpsCode(c, &tx).Create(&info.cpsCode)
	}

	局_临时 := struct {
		dbm.DB_CpsUser
		CpsCode string `json:"cpsCode"`
	}{info.cpsUser, info.cpsCode.CpsCode}

	response.OkWithData(c, 局_临时)
}

func GetInvCodeByUIDUniqueNew(uid string) string {
	var crc16Table [256]uint16
	const poly = 0xA001
	for i := 0; i < 256; i++ {
		crc := uint16(i)
		for j := 0; j < 8; j++ {
			if crc&0x0001 != 0 {
				crc = (crc >> 1) ^ poly
			} else {
				crc >>= 1
			}
		}
		crc16Table[i] = crc
	}

	// 使用查找表计算CRC16并返回hex格式
	crc16WithTable := func(data []byte) string {
		crc := uint16(0xFFFF)

		for _, b := range data {
			index := uint8(crc ^ uint16(b))
			crc = (crc >> 8) ^ crc16Table[index]
		}

		return fmt.Sprintf("%04x", crc)
	}

	// 使用IEEE多项式

	result := crc16WithTable([]byte(uid))
	return result
}
