package item

import (
	. "EFunc/utils"
	"github.com/gin-gonic/gin"
	"github.com/qiniu/go-sdk/v7/storagev2/credentials"
	"github.com/qiniu/go-sdk/v7/storagev2/http_client"
	"github.com/qiniu/go-sdk/v7/storagev2/objects"
	"server/config"
	"server/new/app/models/common"
)

var 云存储_七牛云 Q七牛云

type Q七牛云 struct {
	配置     config.Q七牛云对象存储
	bucket *objects.Bucket
}

func (j *Q七牛云) Q取云存储名称() string {
	return "七牛云"
}

func (j *Q七牛云) C初始化数据(配置 config.Y云存储配置) bool {
	j.配置 = 配置.Q七牛云对象存储

	mac := credentials.NewCredentials(j.配置.AccessKey, j.配置.SecretKey)
	objectsManager := objects.NewObjectsManager(&objects.ObjectsManagerOptions{
		Options: http_client.Options{Credentials: mac},
	})
	j.bucket = objectsManager.Bucket(j.配置.Bucket)

	return j.配置.AccessKey != "" && j.配置.SecretKey != ""
}

func (j *Q七牛云) H获取文件列表(c *gin.Context, 路径前缀 string) (列表 []common.W文件对象详情, err error) {
	iter := j.bucket.List(c, &objects.ListObjectsOptions{Prefix: 路径前缀})
	defer iter.Close()
	var objectInfo objects.ObjectDetails
	var list = make([]common.W文件对象详情, 0, 100)
	for iter.Next(&objectInfo) {
		list = append(list, common.W文件对象详情{
			Name:         W文件_取文件名(objectInfo.Name),
			Path:         objectInfo.Name,
			MD5:          Z字节集_字节集到十六进制(objectInfo.MD5[:]),
			Size:         objectInfo.Size,
			Type:         1,
			UploadedTime: objectInfo.UploadedAt.Unix(),
		})
	}
	if err = iter.Error(); err != nil {
		return
	}
	列表 = list
	return
}
