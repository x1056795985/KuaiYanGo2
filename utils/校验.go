package utils

import (
	"crypto/md5"
	"encoding/hex"
)

// 返回小写32为md5值
func J校验_取md5(data []byte) string {
	h := md5.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
