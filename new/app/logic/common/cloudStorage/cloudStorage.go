package cloudStorage

import (
	系统错误 "errors"
	"github.com/gin-gonic/gin"
	"server/config"
	"server/new/app/logic/common/cloudStorage/item"
	"server/new/app/logic/common/setting"
	"server/new/app/models/common"
)

var L_云存储 Item

type Item struct {
}

// 注册通道接口
type StorageItem interface {
	C初始化数据(配置 config.Y云存储配置) bool
	Q取云存储名称() string
	H获取文件列表(c *gin.Context, 路径前缀 string) (列表 []common.W文件对象详情, err error)
}

func (j *Item) Q取通道(序号 int) (存储接口 StorageItem, err error) {
	局_配置 := setting.Q云存储配置()
	if 序号 == 0 {
		序号 = 局_配置.D当前选择
	}
	switch 序号 {
	default:
		return nil, 系统错误.New("序号错误")
	case 2:
		存储接口 = &item.Q七牛云{}
	}

	存储接口.C初始化数据(setting.Q云存储配置())
	return

}

func (j *Item) H获取文件列表(c *gin.Context, 路径前缀 string) (列表 []common.W文件对象详情, err error) {
	var 存储空间 StorageItem
	存储空间, err = j.Q取通道(0)
	if err != nil {
		return
	}
	return 存储空间.H获取文件列表(c, 路径前缀)
}
