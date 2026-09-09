package item

import (
	"EFunc/utils"
	"context"
	"encoding/json"
	"errors"
	"path"
	"server/app/models/common"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// 兼容S3API协议
type S3Api struct {
	配置       common.S3兼容协议
	minio客户端 *minio.Client
}

func (j *S3Api) Q取云存储名称() string {
	return "S3兼容协议"
}

// 云存储_取地域 从Endpoint自动提取常见云厂商的地域标识, 提取不到返回空
// 腾讯COS: cos.ap-beijing.myqcloud.com | AWS: s3.us-east-1.amazonaws.com | 阿里OSS: oss-cn-hangzhou.aliyuncs.com
func 云存储_取地域(endpoint string) string {
	局_域名 := strings.ToLower(strings.TrimSpace(endpoint))
	// Cloudflare R2官方S3地域固定为auto
	if strings.Contains(局_域名, ".r2.cloudflarestorage.com") {
		return "auto"
	}
	局_排除词 := map[string]bool{"amazonaws": true, "aliyuncs": true, "myqcloud": true, "accelerate": true, "www": true}
	for _, 局_前缀 := range []string{"cos.", "s3.", "s3-", "oss.", "oss-"} {
		局_起点 := strings.Index(局_域名, 局_前缀)
		if 局_起点 < 0 {
			continue
		}
		局_剩余 := 局_域名[局_起点+len(局_前缀):]
		局_终点 := strings.Index(局_剩余, ".")
		if 局_终点 > 0 {
			局_地域 := 局_剩余[:局_终点]
			if !局_排除词[局_地域] && !strings.ContainsAny(局_地域, ":/") {
				return 局_地域
			}
		}
	}
	return ""
}

func (j *S3Api) C初始化数据(配置 common.Y云存储配置) bool {
	j.配置 = 配置.S3兼容协议
	if j.配置.RootPath == "" {
		j.配置.RootPath = "fnkuaiyan/"
	}
	// 根目录统一去掉前导/ (S3对象Key不允许以/开头)
	j.配置.RootPath = strings.TrimLeft(j.配置.RootPath, "/")

	// minio.New要求Endpoint为 host:port 格式, 不能携带协议前缀, 因此必须先剥离
	// 并根据原协议设置Secure; 未填写协议时默认启用HTTPS (R2等S3兼容服务的HTTP会301重定向且不跟随)
	局_是否HTTPS := true
	局_endpoint := strings.TrimSpace(j.配置.Endpoint)
	if strings.HasPrefix(局_endpoint, "http://") {
		局_是否HTTPS = false
		局_endpoint = strings.TrimPrefix(局_endpoint, "http://")
	} else {
		局_endpoint = strings.TrimPrefix(局_endpoint, "https://")
	}

	// 显式指定Region: 未指定时minio-go会先发 GET /?location 自动探测,
	// 但COS等S3兼容服务对非AWS端点的该请求会误报 "The specified key does not exist",
	// 因此优先取配置, 其次从Endpoint自动识别; 自建MinIO识别不到则保持自动探测
	局_地域 := strings.TrimSpace(j.配置.Region)
	if 局_地域 == "" {
		局_地域 = 云存储_取地域(局_endpoint)
	}
	局_选项 := &minio.Options{
		Creds:  credentials.NewStaticV4(j.配置.AccessKey, j.配置.SecretKey, ""),
		Secure: 局_是否HTTPS,
	}
	if 局_地域 != "" {
		局_选项.Region = 局_地域
	}

	// 初始化Minio客户端
	client, err := minio.New(局_endpoint, 局_选项)
	if err != nil {
		return false
	}
	j.minio客户端 = client
	return true
}

func (j *S3Api) H获取文件列表(c *gin.Context, 前缀 string, 分隔符 string) (列表 []common.W文件对象详情, err error) {
	路径前缀 := strings.TrimLeft(j.配置.RootPath+前缀, "/")

	ctx := context.Background()

	// 处理目录分隔符

	对象通道 := j.minio客户端.ListObjects(ctx, j.配置.Bucket, minio.ListObjectsOptions{
		Prefix:       路径前缀,
		Recursive:    true,
		WithVersions: true,
	})

	var list []common.W文件对象详情
	var 目录过滤 = make(map[string]int, len(对象通道))
	for 对象 := range 对象通道 {
		if 对象.Err != nil {
			return nil, 对象.Err
		}
		//aaaa/aaa.jpg  //这种子级文件夹的文件, 就添加为目录,
		局_path := strings.TrimPrefix(对象.Key, 路径前缀)
		if strings.Index(局_path, "/") == -1 { //没有/ 说明是文件
			list = append(list, common.W文件对象详情{
				Name:   path.Base(对象.Key),
				Path:   strings.TrimPrefix(对象.Key, j.配置.RootPath),
				Type:   2,
				UpTime: 对象.LastModified.Unix(),
				Size:   对象.Size,
				MD5:    对象.ETag,
			})
		} else { //有斜杠说明是路径
			局_path = utils.W文本_取文本左边(局_path, "/")
			if _, ok := 目录过滤[局_path]; !ok {
				目录过滤[局_path] = 1
			} else {
				continue
			}

			list = append(list, common.W文件对象详情{
				Name:   局_path,
				Path:   前缀 + 局_path + "/",
				MD5:    对象.ETag,
				Size:   0,
				Type:   1,
				UpTime: 对象.LastModified.Unix(),
			})
		}
	}

	return list, nil
}

func (j *S3Api) Q取文件上传授权(c *gin.Context, 要上传的路径 string) (common.W文件上传凭证, error) {
	对象名称 := path.Join(j.配置.RootPath, 要上传的路径)
	过期时间 := time.Hour

	预签名URL, err := j.minio客户端.PresignedPutObject(context.Background(), j.配置.Bucket, 对象名称, 过期时间)
	if err != nil {
		return common.W文件上传凭证{}, err
	}

	return common.W文件上传凭证{
		Path:    对象名称,
		Type:    1,
		Url:     预签名URL.String(),
		UpToken: "",
	}, nil
}

func (j *S3Api) Q取文件信息(c *gin.Context, 文件路径 string) (common.W文件对象详情, error) {
	对象名称 := path.Join(j.配置.RootPath, 文件路径)
	对象信息, err := j.minio客户端.StatObject(context.Background(), j.配置.Bucket, 对象名称, minio.StatObjectOptions{})
	if err != nil {
		return common.W文件对象详情{}, err
	}
	return common.W文件对象详情{
		Name:   path.Base(对象信息.Key),
		Path:   strings.TrimPrefix(对象信息.Key, j.配置.RootPath),
		Type:   2,
		UpTime: 对象信息.LastModified.Unix(),
		Size:   对象信息.Size,
		MD5:    对象信息.ETag,
	}, nil
}

func (j *S3Api) W文件移动(c *gin.Context, 文件路径, 新文件路径 string) error {
	源对象 := path.Join(j.配置.RootPath, 文件路径)
	目标对象 := path.Join(j.配置.RootPath, 新文件路径)
	ctx := context.Background()

	// Minio需要先复制再删除
	_, err := j.minio客户端.CopyObject(ctx, minio.CopyDestOptions{
		Bucket: j.配置.Bucket,
		Object: 目标对象,
	}, minio.CopySrcOptions{
		Bucket: j.配置.Bucket,
		Object: 源对象,
	})
	if err != nil {
		return err
	}

	return j.minio客户端.RemoveObject(ctx, j.配置.Bucket, 源对象, minio.RemoveObjectOptions{})
}

func (j *S3Api) W文件删除(c *gin.Context, 文件路径 []string) (局_失败计数 int, err error) {
	ctx := context.Background()

	// 创建对象删除通道（使用minio.ObjectInfo替代）
	删除队列 := make(chan minio.ObjectInfo, len(文件路径))

	// 异步填充删除队列
	go func() {
		defer close(删除队列)
		for _, 路径 := range 文件路径 {
			对象名称 := path.Join(j.配置.RootPath, 路径)
			删除队列 <- minio.ObjectInfo{
				Key: 对象名称,
			}
		}
	}()

	// 执行批量删除操作（使用RemoveObjectsV2）
	错误通道 := j.minio客户端.RemoveObjects(ctx, j.配置.Bucket, 删除队列, minio.RemoveObjectsOptions{
		GovernanceBypass: true,
	})

	// 处理错误结果
	for 错误信息 := range 错误通道 {
		if 错误信息.Err != nil {
			局_失败计数++
			err = 错误信息.Err
		}
	}

	return
}

func (j *S3Api) X下载(c *gin.Context, 文件路径 string) (下载地址 string, err error) {
	return j.Q取外链地址(c, 文件路径, 3600)
}

func (j *S3Api) Q取外链地址(c *gin.Context, 文件路径 string, 有效秒数 int64) (下载地址 string, err error) {
	对象名称 := path.Join(j.配置.RootPath, 文件路径)

	if 有效秒数 > 604799 {
		err = errors.New("S3兼容协议,官方限制最长有限期7天(604799秒),如有长时间需求,请公共函数动态生成.")
		return
	}

	过期时间 := time.Duration(有效秒数) * time.Second

	预签名URL, err := j.minio客户端.PresignedGetObject(context.Background(), j.配置.Bucket, 对象名称, 过期时间, nil)
	if err != nil {
		return "", err
	}

	// 使用自定义域名  http://www.baidu.com  替换域名部分
	if j.配置.W外链域名 != "" {
		return j.配置.W外链域名 + 预签名URL.Path + "?" + 预签名URL.RawQuery, nil
	}

	return 预签名URL.String(), nil
}

func (j *S3Api) Q基础信息(c *gin.Context) (响应json信息 string, err error) {
	信息 := map[string]interface{}{
		"bucket": j.配置.Bucket,
		"region": "自定义配置", // Minio需要额外接口获取
		"domain": j.配置.W外链域名,
	}
	json数据, _ := json.Marshal(信息)
	return string(json数据), nil
}

func (j *S3Api) Q基础信息2(c *gin.Context) (基础信息 struct {
	BucketInfo interface{}
	DomainInfo interface{}
}, err error) {
	// Minio需要额外实现具体信息获取
	return struct {
		BucketInfo interface{}
		DomainInfo interface{}
	}{
		BucketInfo: j.配置.Bucket,
		DomainInfo: j.配置.W外链域名,
	}, nil
}
