package utils

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// 子程序名：文件_枚举
// 枚举某个目录下的指定类型文件；成功返回文件数量；
// 返回值类型：整数型
// 参数<1>的名称为“欲寻找的目录”，类型为“文本型”。注明：文件目录。
// 参数<2>的名称为“欲寻找的文件名”，类型为“文本型”。注明：如果寻找全部文件可以填入空白，.txt|.jpg找txt和jpg的文件
// 参数<3>的名称为“文件数组”，类型为“文本型”，接收参数数据时采用参考传递方式，允许接收空参数数据，需要接收数组数据。注明：用于装载文件数组的变量；把寻找到的文件都放在这个数组里，并返回；。
// 参数<4>的名称为“是否带路径”，类型为“逻辑型”，允许接收空参数数据。注明：默认为假； 真=带目录路径，如C:\012.txt； 假=不带，如 012.txt；。
// 参数<6>的名称为“是否遍历子目录”，类型为“逻辑型”，允许接收空参数数据。注明：留空默认为假；为真时文件数组不主动清空。
func W文件_枚举(欲寻找的目录 string, 欲寻找的文件名 string, files *[]string, 是否带路径 bool, 是否遍历子目录 bool) error {
	var ok bool
	欲寻找的文件名arr := strings.Split(欲寻找的文件名, "|")
	l, err := ioutil.ReadDir(欲寻找的目录)
	if err != nil {
		return err
	}

	separator := "/"

	for _, f := range l {
		tmp := 欲寻找的目录 + separator + f.Name()

		if f.IsDir() {
			if 是否遍历子目录 {
				err = W文件_枚举(tmp, 欲寻找的文件名, files, 是否带路径, 是否遍历子目录)
				if err != nil {
					return err
				}
			}
		} else {
			ok = false
			// 目标文件类型被指定
			if !S数组_是否为空(欲寻找的文件名arr) {
				// 属于目标文件类型
				if isInSuffix(欲寻找的文件名arr, f.Name()) {
					ok = true

				}
			} else { // 目标文件类型为空
				ok = true
			}
			if ok {
				if 是否带路径 {
					*files = append(*files, tmp)
				} else {
					*files = append(*files, f.Name())
				}
			}
		}
	}
	return err
}

// 判断目标字符串的末尾是否含有数组中指定的字符串
func isInSuffix(list []string, s string) (isIn bool) {

	isIn = false
	for _, f := range list {

		if strings.TrimSpace(f) != "" && strings.HasSuffix(s, f) {
			isIn = true
			break
		}
	}

	return isIn
}

func W文件_取文件名(路径 string) string {
	return filepath.Base(路径)
}

func W文件_路径合并处理(elem ...string) string {
	return path.Join(elem...)
}

func W文件_取父目录(dirpath string) string {
	return path.Dir(dirpath)
}

func W文件_删除(欲删除的文件名 string) error {
	return os.Remove(欲删除的文件名)

}

// 调用格式： 〈逻辑型〉 文件更名 （文本型 欲更名的原文件或目录名，文本型 欲更改为的现文件或目录名） - 系统核心支持库->磁盘操作
// 英文名称：name
// 重新命名一个文件或目录。成功返回真，失败返回假。本命令为初级命令。
// 参数<1>的名称为“欲更名的原文件或目录名”，类型为“文本型（text）”。
// 参数<2>的名称为“欲更改为的现文件或目录名”，类型为“文本型（text）”。
//
// 操作系统需求： Windows、Linux
func W文件_更名(欲更名的原文件或目录名 string, 欲更改为的现文件或目录名 string) error {
	return os.Rename(欲更名的原文件或目录名, 欲更改为的现文件或目录名)
}

// 路径不存在时自动创建
func W文件_写出(文件名 string, 欲写入文件的数据 interface{}) error {
	fpath := W文件_取父目录(文件名)
	if !W文件_是否存在(fpath) {
		M目录_创建(fpath)
	}
	return ioutil.WriteFile(文件名, D到字节集(欲写入文件的数据), os.ModePerm)
}

// 路径不存在时自动创建
// 为习惯添加的函数
func W文件_写出文件(文件名 string, 欲写入文件的数据 interface{}) error {
	return W文件_写出(文件名, 欲写入文件的数据)
}

// 路径不存在时自动创建
func W文件_追加文本(文件名 string, 欲追加的文本 string) error {
	fpath := W文件_取父目录(文件名)
	if !W文件_是否存在(fpath) {
		M目录_创建(fpath)
	}
	file, err := os.OpenFile(文件名, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
	defer file.Close()

	_, err = file.Write(D到字节集(欲追加的文本 + "\r\n"))
	return err
}

// 从路径中读入文件的文本内容
func W文件_读入文本(文件名 string) string {
	var data []byte
	data, _ = ioutil.ReadFile(文件名)
	return D到文本(data)
}

// 从路径中读入文件的文本内容
func W文件_读入文件(文件名 string) []byte {
	var data []byte
	data, _ = ioutil.ReadFile(文件名)
	return data
}

// 自动检查内容是否一致是否需要写出
func W文件_保存(文件名 string, 欲写入文件的数据 interface{}) error {
	if W文件_是否存在(文件名) {
		data := W文件_读入文件(文件名)
		wdata := D到字节集(欲写入文件的数据)
		if !bytes.Equal(data, wdata) {
			//E调试输出("不相同写出")
			return W文件_写出(文件名, wdata)
		}
	} else {
		return W文件_写出(文件名, 欲写入文件的数据)
	}
	return nil
}

// W文件_是否存在 判断一个文件或文件夹是否存在
// 输入文件路径，根据返回的bool值来判断文件或文件夹是否存在
func W文件_是否存在(路径 string) bool {
	_, err := os.Stat(路径)
	if err == nil {
		return true
	}

	return false
}

// 调用格式： 〈文本型〉 取临时文件名 （［文本型 目录名］） - 系统核心支持库->磁盘操作
// 英文名称：GetTempFileName
// 返回一个在指定目录中确定不存在的 .TMP 全路径文件名称。本命令为初级命令。
// 参数<1>的名称为“目录名”，类型为“文本型（text）”，可以被省略。如果省略本参数，默认将使用系统的标准临时目录。
//
// 操作系统需求： Windows
func W文件_取临时文件名(目录名 string) (f *os.File, filepath string, err error) {
	prefix := ""
	f, err = ioutil.TempFile(目录名, prefix)
	filepath = 目录名 + f.Name()
	return f, filepath, err
}

//调用格式： 〈整数型〉 取文件尺寸 （文本型 文件名） - 系统核心支持库->磁盘操作
//英文名称：FileLen
//返回一个文件的长度，单位是字节。如果该文件不存在，将返回 -1。本命令为初级命令。
//参数<1>的名称为“文件名”，类型为“文本型（text）”。
//
//操作系统需求： Windows、Linux

func W文件_取大小(文件名 string) int64 {
	f, err := os.Stat(文件名)
	if err == nil {
		return f.Size()
	} else {
		return -1
	}
}
