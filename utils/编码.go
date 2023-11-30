package utils

import "encoding/base64"

func B编码_base64编码(待编码数据 []byte) string {
	return base64.StdEncoding.EncodeToString(待编码数据)
}
func B编码_base64解码(待解码文本 string) []byte {
	decodeData, err := base64.StdEncoding.DecodeString(待解码文本)
	if err != nil {
		return []byte{}
	}

	return decodeData
}
