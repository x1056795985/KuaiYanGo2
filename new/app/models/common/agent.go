package common

type D代理分成结构 struct {
	Uid      int
	User     string
	F分成百分比   int     //就是设置的自身百分比
	F分给下级百分比 int     //下级的百分比
	S实际自身百分比 int     //自身百分比-下级的百分比  实际会分到的百分比
	S实际分成金额  float64 //
}
