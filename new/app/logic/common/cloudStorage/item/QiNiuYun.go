package item

import (
	. "EFunc/utils"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	"github.com/qiniu/go-sdk/v7/storagev2/credentials"
	"github.com/qiniu/go-sdk/v7/storagev2/downloader"
	"github.com/qiniu/go-sdk/v7/storagev2/http_client"
	"github.com/qiniu/go-sdk/v7/storagev2/objects"
	"github.com/qiniu/go-sdk/v7/storagev2/uptoken"
	"net/url"
	"server/config"
	"server/new/app/models/common"
	"strings"
	"time"
)

type Q七牛云 struct {
	配置     config.Q七牛云对象存储
	bucket *objects.Bucket
}

func (j *Q七牛云) Q取云存储名称() string {
	return "七牛云"
}

func (j *Q七牛云) C初始化数据(配置 config.Y云存储配置) bool {
	j.配置 = 配置.Q七牛云对象存储
	if j.配置.RootPath == "" {
		j.配置.RootPath = "fnkuaiyan/"
	}

	mac := credentials.NewCredentials(j.配置.AccessKey, j.配置.SecretKey)
	objectsManager := objects.NewObjectsManager(&objects.ObjectsManagerOptions{
		Options: http_client.Options{Credentials: mac},
	})
	j.bucket = objectsManager.Bucket(j.配置.Bucket)
	return j.配置.AccessKey != "" && j.配置.SecretKey != "" && j.配置.Bucket != ""
}

func (j *Q七牛云) H获取文件列表(c *gin.Context, 前缀 string, 分隔符 string) (列表 []common.W文件对象详情, err error) {
	//删除左边的  /
	路径前缀 := strings.TrimLeft(j.配置.RootPath+前缀, "/")
	局_标志 := ""
	var list = make([]common.W文件对象详情, 0, 100)
	var objectInfo objects.ObjectDetails
	iter := j.bucket.List(c, &objects.ListObjectsOptions{Prefix: 路径前缀, Marker: 局_标志, NeedParts: true})
	defer iter.Close()
	局_目录信息 := ""
	for iter.Next(&objectInfo) {
		if objectInfo.Name == 路径前缀 { //跳过目录自身
			continue
		}
		局_临时目录 := W文本_取出中间文本(objectInfo.Name, j.配置.RootPath+前缀, "/")
		if 局_临时目录 != "" && strings.Index(局_目录信息, "\n"+前缀+局_临时目录+"/") == -1 {
			局_目录信息 += "\n" + 局_临时目录 + "/"
		}

		//只获取子级目录 1次是目录, 0次是文件
		if W文本_取右边(objectInfo.Name, 1) == "/" || strings.Count(W文本_取文本右边(objectInfo.Name, 路径前缀), "/") >= 1 {
			continue
		}

		list = append(list, common.W文件对象详情{
			Name:   W文件_取文件名(objectInfo.Name),
			Path:   objectInfo.Name[len(j.配置.RootPath):],
			MD5:    Z字节集_字节集到十六进制(objectInfo.MD5[:]),
			Size:   objectInfo.Size,
			Type:   2,
			UpTime: objectInfo.UploadedAt.Unix(),
		})
	}
	if err = iter.Error(); err != nil {
		return
	}
	//开始处理 直属子级目录
	局_目录信息_数组 := strings.Split(局_目录信息, "\n")
	for _, 目录 := range 局_目录信息_数组 {
		if 目录 == "" {
			continue
		}
		list = append(list, common.W文件对象详情{
			Name:   W文本_取文本左边(目录, "/"),
			Path:   前缀 + 目录,
			MD5:    "d41d8cd98f00b204e9800998ecf8427e",
			Size:   0,
			Type:   1,
			UpTime: 0,
		})
	}

	列表 = list
	return
}

func (j *Q七牛云) Q取文件上传授权(c *gin.Context, 要上传的路径 string) (common.W文件上传凭证, error) {

	mac := credentials.NewCredentials(j.配置.AccessKey, j.配置.SecretKey)
	bucket := j.配置.Bucket
	// 需要覆盖的文件名
	keyToOverwrite := j.配置.RootPath + 要上传的路径
	putPolicy, err := uptoken.NewPutPolicyWithKey(bucket, keyToOverwrite, time.Now().Add(1*time.Hour))
	if err != nil {
		return common.W文件上传凭证{}, err
	}
	upToken, err := uptoken.NewSigner(putPolicy, mac).GetUpToken(c)
	if err != nil {
		return common.W文件上传凭证{}, err
	}

	//up-cn-east-2.qiniup.com
	// 初始化 BucketManager
	Cfg, _ := j.Q基础信息2(c)
	//https://blog.csdn.net/Coin_Collecter/article/details/129813929
	局_上传地址 := ""
	switch Cfg.BucketInfo.Zone {
	case "z0":
		局_上传地址 = "https://upload.qiniup.com"
	case "cn-east-2":
		局_上传地址 = "https://upload-cn-east-2.qiniup.com"
	case "z1":
		局_上传地址 = "https:///upload-z1.qiniup.com"
	case "z2":
		局_上传地址 = "https://upload-z2.qiniup.com"
	case "na0":
		局_上传地址 = "https://upload-na0.qiniup.com"
	case "as0":
		局_上传地址 = "https://upload-as0.qiniup.com"
	case "ap-northeast-1":
		局_上传地址 = "https://upload-ap-northeast-1.qiniup.com"
	default:
		局_上传地址 = "https://pload-" + Cfg.BucketInfo.Zone + ".qiniup.com" // 默认地址
	} //up-cn-east-2.qiniup.com   cn-east-2

	return common.W文件上传凭证{Path: keyToOverwrite, Type: 2, Url: 局_上传地址, UpToken: upToken}, nil
}

// 移动文件也是重命名文件
func (j *Q七牛云) W文件移动(c *gin.Context, 文件路径, 新文件路径 string) error {
	key := j.配置.RootPath + 文件路径
	//目标空间可以和源空间相同，但是不能为跨机房的空间
	destBucket := j.配置.Bucket
	//目标文件名可以和源文件名相同，也可以不同
	destKey := j.配置.RootPath + 新文件路径
	err := j.bucket.Object(key).MoveTo(destBucket, destKey).Call(c)
	if err != nil {
		return err
	}
	return nil
}

func (j *Q七牛云) W文件删除(c *gin.Context, 文件路径 []string) (局_失败计数 int, err error) {

	mac := credentials.NewCredentials(j.配置.AccessKey, j.配置.SecretKey)
	objectsManager := objects.NewObjectsManager(&objects.ObjectsManagerOptions{
		Options: http_client.Options{Credentials: mac},
	})

	onError := func(err2 error) {
		局_失败计数++
		err = err2
		//fmt.Printf("%s\n", err)
	}

	operations := make([]objects.Operation, 0, len(文件路径))
	for _, path := range 文件路径 {
		key := j.配置.RootPath + path
		operations = append(operations, j.bucket.Object(key).Delete().OnError(onError))
	}

	err = objectsManager.Batch(c, operations, nil)
	if err != nil {
		return
	}
	return
}

func (j *Q七牛云) X下载(c *gin.Context, 文件路径 string) (下载地址 string, err error) {
	mac := credentials.NewCredentials(j.配置.AccessKey, j.配置.SecretKey)
	urlsProvider := downloader.SignURLsProvider(downloader.NewDefaultSrcURLsProvider(mac.AccessKey, nil), downloader.NewCredentialsSigner(mac), nil)
	var iter downloader.URLsIter
	iter, err = urlsProvider.GetURLsIter(c, j.配置.RootPath+文件路径, &downloader.GenerateOptions{
		BucketName:          j.配置.Bucket,
		UseInsecureProtocol: true,
	})
	if err != nil {
		return
	}
	defer iter.Clone()
	urls := make([]*url.URL, 0, 16)
	for {
		u := new(url.URL)
		var ok bool
		ok, err = iter.Peek(u)
		if err != nil || !ok {
			break
		}
		下载地址 = u.String()
		urls = append(urls, u)
		iter.Next()
	}
	return
}

func (j *Q七牛云) Q取外链地址(c *gin.Context, 文件路径 string, 有效秒数 int64) (下载地址 string, err error) {
	domain := j.配置.W外链域名

	if domain == "" {
		info, err2 := j.Q基础信息2(c)
		if err2 == nil {
			if len(info.DomainInfo) > 0 {
				domain = info.DomainInfo[0].Domain
			}
		}
	}
	mac := credentials.NewCredentials(j.配置.AccessKey, j.配置.SecretKey)

	urlsProvider := downloader.SignURLsProvider(downloader.NewStaticDomainBasedURLsProvider([]string{domain}), downloader.NewCredentialsSigner(mac), &downloader.SignOptions{
		TTL: time.Duration(有效秒数) * time.Second, // 有效期
	})

	var iter downloader.URLsIter
	iter, err = urlsProvider.GetURLsIter(c, j.配置.RootPath+文件路径, &downloader.GenerateOptions{
		BucketName:          j.配置.Bucket,
		UseInsecureProtocol: true,
	})
	if err != nil {
		return
	}
	defer iter.Clone()
	urls := make([]*url.URL, 0, 16)
	for {
		u := new(url.URL)
		var ok bool
		ok, err = iter.Peek(u)
		if err != nil || !ok {
			break
		}
		下载地址 = u.String()
		urls = append(urls, u)
		iter.Next()
	}
	return
}
func (j *Q七牛云) Q基础信息(c *gin.Context) (响应json信息 string, err error) {
	基础信息, err2 := j.Q基础信息2(c)
	if err2 != nil {
		return "", err
	}
	marshal, err := json.Marshal(基础信息)
	if err != nil {
		return "", err
	}
	响应json信息 = string(marshal)
	return
}
func (j *Q七牛云) Q基础信息2(c *gin.Context) (基础信息 struct {
	Cfg        storage.Config // 创建配置对象.
	BucketInfo storage.BucketInfo
	DomainInfo []storage.DomainInfo
}, err error) {
	var mac *qbox.Mac
	mac = qbox.NewMac(j.配置.AccessKey, j.配置.SecretKey) // 初始化 Mac 对象

	// 初始化 BucketManager
	bucketManager := storage.NewBucketManager(mac, &基础信息.Cfg)
	基础信息.BucketInfo, err = bucketManager.GetBucketInfo(j.配置.Bucket)
	// 获取域名列表
	基础信息.DomainInfo, err = bucketManager.ListBucketDomains(j.配置.Bucket)
	return
}
