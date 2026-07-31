package common

import (
	"github.com/gogf/gf/v2/encoding/gjson"
	"server/app/models/db"
)

type Q请求_上下文 struct {
	AppInfo      db.DB_AppInfo    // 应用信息
	Z在线信息        db.DB_LinksToken // 在线Token信息(无Token请求如GetToken时为零值)
	Q请求明文        *gjson.Json      // 解密后的请求明文JSON
	X响应明文        *gjson.Json      // 解密后的请求明文JSON
	Api          string           // 解密后的Api名称
	C成功状态码       int              // 客户端期望的成功状态码
	CryptoKeyAes string           // 通讯AES密钥(响应加密用)
	B值长度         int
	RSA强制        bool // 是否强制RSA加密响应
	W无Token请求    bool // 是否为无Token请求(GetToken等)
}
