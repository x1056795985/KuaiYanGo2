package cloudStorage

import (
	"errors"
	"github.com/gin-gonic/gin"
	"server/app/logic/common/cloudStorage/item"
	"server/app/logic/common/setting"
	"server/app/models/common"
	"sort"
	"strings"
	"sync"
	"time"
)

var L_云存储 Item

type Item struct {
}

// ETag缓存结构
type 缓存_ETag项 struct {
	ETag      string
	写入时间戳 int64
}

// 全局ETag缓存, key=文件路径
var 集_ETag缓存 = struct {
	sync.RWMutex
	数据 map[string]缓存_ETag项
}{
	数据: make(map[string]缓存_ETag项),
}

// ETag缓存有效期(秒), 超时后自动重新获取
const ETag缓存有效期 = 300

// Q取ETag 获取指定路径文件的ETag, 优先从全局缓存读取, 缓存过期或不存在时从云存储获取
func (j *Item) Q取ETag(c *gin.Context, 文件路径 string) (ETag string, err error) {
	// 先查缓存
	集_ETag缓存.RLock()
	局_项, 局_存在 := 集_ETag缓存.数据[文件路径]
	集_ETag缓存.RUnlock()

	if 局_存在 && (time.Now().Unix()-局_项.写入时间戳) < ETag缓存有效期 {
		return 局_项.ETag, nil
	}

	// 缓存不存在或已过期, 从云存储获取
	局_文件信息, 局_错误 := j.Q取文件信息(c, 文件路径)
	if 局_错误 != nil {
		return "", 局_错误
	}

	ETag = 局_文件信息.MD5

	// 写入缓存
	集_ETag缓存.Lock()
	集_ETag缓存.数据[文件路径] = 缓存_ETag项{
		ETag:      ETag,
		写入时间戳: time.Now().Unix(),
	}
	集_ETag缓存.Unlock()

	return ETag, nil
}

// G更新ETag缓存 在文件上传后调用, 删除旧缓存使下次获取时重新从云存储拉取
func (j *Item) G更新ETag缓存(文件路径 string) {
	集_ETag缓存.Lock()
	delete(集_ETag缓存.数据, 文件路径)
	集_ETag缓存.Unlock()
}

// 注册通道接口
type StorageItem interface {
	C初始化数据(配置 common.Y云存储配置) bool
	Q取云存储名称() string
	H获取文件列表(c *gin.Context, 路径前缀, 分隔符 string) (列表 []common.W文件对象详情, err error)
	Q取文件上传授权(c *gin.Context, 要上传的路径 string) (common.W文件上传凭证, error)
	Q取文件信息(c *gin.Context, 文件路径 string) (common.W文件对象详情, error)
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

// H获取文件分页列表 统一处理不同存储通道的目录排序、关键词过滤和分页。
func (j *Item) H获取文件分页列表(c *gin.Context, 路径前缀, 分隔符, 关键词, 排序字段 string, 页码, 每页数量, 排序方式 int) ([]common.W文件对象详情, int64, error) {
	局_全部列表, 局_错误 := j.H获取文件列表(c, 路径前缀, 分隔符)
	if 局_错误 != nil {
		return nil, 0, 局_错误
	}

	局_关键词 := strings.ToLower(strings.TrimSpace(关键词))
	局_筛选列表 := make([]common.W文件对象详情, 0, len(局_全部列表))
	for _, 局_文件 := range 局_全部列表 {
		if 局_关键词 == "" || strings.Contains(strings.ToLower(局_文件.Name), 局_关键词) {
			局_筛选列表 = append(局_筛选列表, 局_文件)
		}
	}

	局_是否升序 := 排序方式 == 1
	局_排序字段 := strings.ToLower(strings.TrimSpace(排序字段))
	sort.SliceStable(局_筛选列表, func(局_左序号, 局_右序号 int) bool {
		局_左文件 := 局_筛选列表[局_左序号]
		局_右文件 := 局_筛选列表[局_右序号]
		if 局_左文件.Type != 局_右文件.Type {
			return 局_左文件.Type == 1
		}

		局_比较结果 := 0
		switch 局_排序字段 {
		case "size":
			if 局_左文件.Size < 局_右文件.Size {
				局_比较结果 = -1
			} else if 局_左文件.Size > 局_右文件.Size {
				局_比较结果 = 1
			}
		case "uptime":
			if 局_左文件.UpTime < 局_右文件.UpTime {
				局_比较结果 = -1
			} else if 局_左文件.UpTime > 局_右文件.UpTime {
				局_比较结果 = 1
			}
		default:
			局_比较结果 = strings.Compare(strings.ToLower(局_左文件.Name), strings.ToLower(局_右文件.Name))
		}
		if 局_比较结果 == 0 {
			局_比较结果 = strings.Compare(局_左文件.Path, 局_右文件.Path)
		}
		if 局_是否升序 {
			return 局_比较结果 < 0
		}
		return 局_比较结果 > 0
	})

	if 页码 < 1 {
		页码 = 1
	}
	if 每页数量 < 1 {
		每页数量 = 20
	}
	if 每页数量 > 100 {
		每页数量 = 100
	}

	局_总数 := int64(len(局_筛选列表))
	局_开始序号 := (页码 - 1) * 每页数量
	if 局_开始序号 >= len(局_筛选列表) {
		return []common.W文件对象详情{}, 局_总数, nil
	}
	局_结束序号 := 局_开始序号 + 每页数量
	if 局_结束序号 > len(局_筛选列表) {
		局_结束序号 = len(局_筛选列表)
	}

	return 局_筛选列表[局_开始序号:局_结束序号], 局_总数, nil
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

func (j *Item) Q取文件信息(c *gin.Context, 文件路径 string) (common.W文件对象详情, error) {
	var 存储空间 StorageItem
	var err error

	存储空间, err = j.Q取通道(0)
	if err != nil {
		return common.W文件对象详情{}, err
	}
	return 存储空间.Q取文件信息(c, 文件路径)
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
		有效秒数 = 604799
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
