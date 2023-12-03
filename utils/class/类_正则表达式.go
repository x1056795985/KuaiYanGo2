package class

import "regexp"

type L_正则表达式 struct {
	//正则
	r *regexp.Regexp
	//返回结果
	res [][]string
	//批量数量
	Count int
	//子匹配数量
	SubmatchCount2 int
}

func New正则表达式类(正则表达式文本 string, 被搜索的文本 string) (*L_正则表达式, bool) {
	t := new(L_正则表达式)
	b := t.E创建(正则表达式文本, 被搜索的文本)
	return t, b
}
func (this *L_正则表达式) E创建(正则表达式文本 string, 被搜索的文本 string) bool {
	this.r = regexp.MustCompile(正则表达式文本) //` `表示使用原生字符串
	this.res = this.r.FindAllStringSubmatch(被搜索的文本, -1)
	this.Count = len(this.res)
	if this.Count == 0 {
		return false
	}
	this.SubmatchCount2 = len(this.res[0])
	return true
}
func (this *L_正则表达式) Q取匹配数量() int {
	return len(this.res)
}
func (this *L_正则表达式) Q取匹配文本(匹配索引 int) string {
	return this.res[匹配索引][0]
}
func (this *L_正则表达式) Q取子匹配文本(匹配索引 int, 子表达式索引 int) string {
	if -1 >= 匹配索引 || -1 >= 子表达式索引 {
		return ""
	}
	if this.Count <= 匹配索引 || this.SubmatchCount2 <= 子表达式索引 {
		return ""
	}

	return this.res[匹配索引][子表达式索引]
}
func (this *L_正则表达式) Q取子匹配数量() int {
	return this.SubmatchCount2
}

func (this *L_正则表达式) GetResult() [][]string {
	return this.res
}
