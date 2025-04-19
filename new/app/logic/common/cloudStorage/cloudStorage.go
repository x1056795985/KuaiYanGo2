package cloudStorage

import (
	"errors"
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
	H获取文件列表(c *gin.Context, 路径前缀, 分隔符 string) (列表 []common.W文件对象详情, err error)
	Q取文件上传授权(c *gin.Context, 要上传的路径 string) (common.W文件上传凭证, error)
	W文件删除(c *gin.Context, 要上传的路径 []string) (局_失败计数 int, err error)
	W文件移动(c *gin.Context, 文件路径, 新文件路径 string) (err error)
	X下载(c *gin.Context, 文件路径 string) (下载地址 string, err error)
	Q取外链地址(c *gin.Context, 文件路径 string, 有效时间 int64) (下载地址 string, err error)
	Q基础信息(c *gin.Context) (响应json信息 string, err error)
}

func (j *Item) Q取通道(序号 int) (存储接口 StorageItem, err error) {
	局_配置 := setting.Q云存储配置()
	if 序号 == 0 {
		序号 = 局_配置.D当前选择
	}
	switch 序号 {
	default:
		return nil, errors.New("序号错误")
	case 1:
		存储接口 = &item.S3Api{}
	case 2:
		存储接口 = &item.Q七牛云{}
	}
	if !存储接口.C初始化数据(setting.Q云存储配置()) {
		err = errors.New("云存储配置初始化失败,请检查系统设置->云存储配置->[" + 存储接口.Q取云存储名称() + "]参数配置是否正确")
	}

	return

}

func (j *Item) H获取文件列表(c *gin.Context, 路径前缀, 分隔符 string) (列表 []common.W文件对象详情, err error) {
	var 存储空间 StorageItem
	存储空间, err = j.Q取通道(0)
	if err != nil {
		return
	}
	return 存储空间.H获取文件列表(c, 路径前缀, 分隔符)
}

func (j *Item) Q取文件上传授权(c *gin.Context, 要上传的路径 string) (common.W文件上传凭证, error) {
	var 存储空间 StorageItem
	var err error

	存储空间, err = j.Q取通道(0)
	if err != nil {
		return common.W文件上传凭证{}, err
	}
	return 存储空间.Q取文件上传授权(c, 要上传的路径)
}

func (j *Item) W文件删除(c *gin.Context, 要上传的路径 []string) (局_失败计数 int, err error) {
	var 存储空间 StorageItem

	存储空间, err = j.Q取通道(0)
	if err != nil {
		return 0, err
	}
	return 存储空间.W文件删除(c, 要上传的路径)
}

func (j *Item) W文件移动(c *gin.Context, 文件路径, 新文件路径 string) (err error) {
	var 存储空间 StorageItem

	存储空间, err = j.Q取通道(0)
	if err != nil {
		return err
	}
	return 存储空间.W文件移动(c, 文件路径, 新文件路径)
}

func (j *Item) X下载(c *gin.Context, 文件路径 string) (下载地址 string, err error) {
	var 存储空间 StorageItem

	存储空间, err = j.Q取通道(0)
	if err != nil {
		return
	}
	return 存储空间.X下载(c, 文件路径)
}

func (j *Item) Q取外链地址(c *gin.Context, 文件路径 string, 有效秒数 int64) (下载地址 string, err error) {
	var 存储空间 StorageItem

	存储空间, err = j.Q取通道(0)
	if err != nil {
		return
	}

	if 有效秒数 == 0 {
		有效秒数 = 604800
	}
	return 存储空间.Q取外链地址(c, 文件路径, 有效秒数)
}

func (j *Item) Q取基础信息(c *gin.Context) (json string, err error) {
	var 存储空间 StorageItem
	存储空间, err = j.Q取通道(0)
	if err != nil {
		return
	}
	return 存储空间.Q基础信息(c)
}
