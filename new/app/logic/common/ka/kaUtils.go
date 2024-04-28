package ka

import (
	"math/rand"
	"strings"
)

func 生成校验字符(str string) string {

	ArrInt := []byte(str)
	Int := 0
	for _, 值 := range ArrInt {
		Int += int(值)
	}
	Int = Int % len(str)

	return string(str[Int])
}
func Ka校验卡号(str string) bool {
	if len(str) < 2 {
		return false
	}
	局_待校验文本 := str[0 : len(str)-1]
	局_校验值 := string(str[len(str)-1])

	return 生成校验字符(局_待校验文本) == 局_校验值
}

func 生成随机字符串(lenNum int, 类型 int) string {

	var CHARS []string
	switch 类型 {
	case 2:
		CHARS = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
			"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}
	case 3:
		CHARS = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
			"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}
	default:
		CHARS = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "m", "n", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
			"A", "B", "C", "D", "E", "F", "G", "H", "J", "K", "L", "M", "N", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
			"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"} //删除一些容易误会的字符,大写的i 小写的l o O
	}

	str := strings.Builder{}
	length := len(CHARS)
	for i := 0; i < lenNum; i++ {
		str.WriteString(CHARS[rand.Intn(length)])
	}
	return str.String()
}
