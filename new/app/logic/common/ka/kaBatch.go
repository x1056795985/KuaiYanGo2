package ka

import (
	. "EFunc/utils"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/new/app/global"
	"server/new/app/models/constant"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	"strconv"
	"time"
)

// Ka批量创建 批量创建卡号
func (j *ka) Ka批量创建(c *gin.Context, 卡信息切片 []dbm.DB_Ka, 卡类id, 制卡人id int, 制卡人账号 string, 管理员备注 string, 代理备注 string, 有效期时间戳 int64) error {
	局_db := *global.GVA_DB
	if len(卡信息切片) > 2400 {
		return errors.New("每批次最大数量不能超过2400")
	}

	KaClass详细信息, err := service.NewKaClass(c, &局_db).KaClass取详细信息(卡类id)
	if err != nil {
		return err
	}

	type 卡号记录 struct {
		Index   int
		Name    string
		AutoGen bool
	}
	var 待检查卡号 []卡号记录

	for i := range 卡信息切片 {
		if 卡信息切片[i].Name == "" {
			name := KaClass详细信息.Prefix + 生成随机字符串(KaClass详细信息.KaLength-len(KaClass详细信息.Prefix), KaClass详细信息.KaStringType)
			卡信息切片[i].Name = name
			待检查卡号 = append(待检查卡号, 卡号记录{Index: i, Name: name, AutoGen: true})
		} else {
			待检查卡号 = append(待检查卡号, 卡号记录{Index: i, Name: 卡信息切片[i].Name, AutoGen: false})
		}
	}

	nameSet := make(map[string]struct{})
	for _, rec := range 待检查卡号 {
		if _, ok := nameSet[rec.Name]; ok {
			return errors.New("本批次内卡号重复: " + rec.Name)
		}
		nameSet[rec.Name] = struct{}{}
	}

	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		maxRetries := 10
		for retry := 0; retry <= maxRetries; retry++ {
			nameList := make([]string, len(待检查卡号))
			for i, rec := range 待检查卡号 {
				nameList[i] = rec.Name
			}
			var existingNames []string
			err = tx.Model(dbm.DB_Ka{}).Select("Name").Where("Name IN (?)", nameList).Find(&existingNames).Error
			if err != nil {
				return err
			}
			existingSet := make(map[string]bool)
			for _, n := range existingNames {
				existingSet[n] = true
			}

			var conflictIndices []int
			for i, rec := range 待检查卡号 {
				if existingSet[rec.Name] {
					if rec.AutoGen {
						conflictIndices = append(conflictIndices, i)
					} else {
						return errors.New("卡号:" + rec.Name + "已存在无法使用")
					}
				}
			}

			if len(conflictIndices) == 0 {
				break
			}
			if retry == maxRetries {
				return errors.New("创建失败,连续多次随机到重复卡号,请尝试删除无用卡号,再重新制卡")
			}

			for _, idx := range conflictIndices {
				rec := &待检查卡号[idx]
				newName := KaClass详细信息.Prefix + 生成随机字符串(KaClass详细信息.KaLength-len(KaClass详细信息.Prefix), KaClass详细信息.KaStringType)
				rec.Name = newName
				卡信息切片[rec.Index].Name = newName
			}

			nameSet = make(map[string]struct{})
			for _, rec := range 待检查卡号 {
				if _, ok := nameSet[rec.Name]; ok {
					return errors.New("本批次内卡号重复: " + rec.Name)
				}
				nameSet[rec.Name] = struct{}{}
			}
		}

		for i := range 卡信息切片 {
			卡信息切片[i].AppId = KaClass详细信息.AppId
			卡信息切片[i].KaClassId = KaClass详细信息.Id
			卡信息切片[i].Status = 1
			卡信息切片[i].RegisterId = 制卡人id
			卡信息切片[i].RegisterUser = 制卡人账号
			卡信息切片[i].RegisterTime = time.Now().Unix()
			卡信息切片[i].AdminNote = 管理员备注
			卡信息切片[i].AgentNote = 代理备注
			卡信息切片[i].VipTime = KaClass详细信息.VipTime
			卡信息切片[i].InviteCount = KaClass详细信息.InviteCount
			卡信息切片[i].RMb = KaClass详细信息.RMb
			卡信息切片[i].VipNumber = KaClass详细信息.VipNumber
			卡信息切片[i].Money = KaClass详细信息.Money
			卡信息切片[i].AgentMoney = KaClass详细信息.AgentMoney
			卡信息切片[i].UserClassId = KaClass详细信息.UserClassId
			卡信息切片[i].NoUserClass = KaClass详细信息.NoUserClass
			卡信息切片[i].KaType = KaClass详细信息.KaType
			卡信息切片[i].MaxOnline = KaClass详细信息.MaxOnline
			卡信息切片[i].Num = 0
			卡信息切片[i].NumMax = KaClass详细信息.Num
			卡信息切片[i].User = ""
			卡信息切片[i].UserTime = ""
			卡信息切片[i].UseTime = 0
			卡信息切片[i].InviteUser = ""
			卡信息切片[i].EndTime = 9999999999
			if 有效期时间戳 != 0 {
				卡信息切片[i].EndTime = 有效期时间戳
			}
		}

		err = tx.Model(dbm.DB_Ka{}).Create(&卡信息切片).Error
		return err
	})
}

// Ka修改状态_同步卡号模式软件用户 修改卡号状态并同步卡号模式软件用户状态
func (j *ka) Ka修改状态_同步卡号模式软件用户(c *gin.Context, id []int, status int) error {
	局_sql := `SELECT DISTINCT AppId  FROM db_App_Info  WHERE AppId IN (SELECT DISTINCT AppId  FROM db_Ka  WHERE Id IN ?) AND AppType IN (3,4)`

	var 局数组_卡号Appid []int
	global.GVA_DB.Raw(局_sql, id).Scan(&局数组_卡号Appid)
	if len(局数组_卡号Appid) == 0 {
		return global.GVA_DB.Model(dbm.DB_Ka{}).Where("Id IN ? ", id).Update("Status", status).Error
	}
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		局_ka := make([]dbm.DB_Ka, 0, len(id))
		err := tx.Model(dbm.DB_Ka{}).Select("Id,AppId").Where("Id IN ?", id).Scan(&局_ka).Error
		if err != nil {
			return err
		}

		局_map := make(map[int][]int, len(id)+1)
		for _, 值 := range 局_ka {
			局_最终AppId := 值.AppId
			if !S数组_整数是否存在(局数组_卡号Appid, 局_最终AppId) {
				局_最终AppId = 1
			}
			if _, ok := 局_map[局_最终AppId]; !ok {
				局_map[局_最终AppId] = make([]int, 0, len(id))
			}
			局_map[局_最终AppId] = append(局_map[局_最终AppId], 值.Id)
		}

		for AppId := range 局_map {
			err = tx.Model(dbm.DB_Ka{}).Where("Id IN ? ", 局_map[AppId]).Update("Status", status).Error
			if err != nil {
				return err
			}
			if AppId >= 10000 {
				err = tx.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Uid IN ? ", 局_map[AppId]).Update("Status", status).Error
				if err != nil {
					return err
				}
			}
			if status == 2 {
				_ = service.NewLinksToken(c, global.GVA_DB).Set批量注销Uid数组(局_map[AppId], AppId, constant.Z注销_管理员手动注销)
			}
		}
		return nil
	})
}
