package Common

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"reflect"
	"server/app/global"
	"server/app/models/old/response"
)

type Common struct {
}

// 统一反序列化参数
func (C *Common) ToJSON(c *gin.Context, obj any) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		// 获取validator.ValidationErrors类型的errors
		//20250411 发现检测有个问题 如果是逻辑型值为false的参数   参数开启了required 必填 他也会报错 参数不存在 解决办法,不校验required
		errs, ok := err.(validator.ValidationErrors)
		errStr := ""
		if !ok {
			errStr = "参数错误:" + err.Error() //	// 非validator.ValidationErrors类型错误直接返回
		} else {
			for _, v := range errs.Translate(global.Trans) { // validator.ValidationErrors类型错误则进行翻译
				errStr += v + ","
			}
		}
		response.FailWithMessage(errStr, c)
		return false
	}
	F钳制分页参数(obj)
	return true
}

// F钳制分页参数 统一分页参数钳制:Page>=1, 1<=Size<=10000
// 防止负Size直通SQL(MySQL下等于无LIMIT全量拉取)、超大Size拖库、负Page产生负Offset
func F钳制分页参数(obj any) {
	if obj == nil {
		return
	}
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return
	}
	钳制结构体字段(v)
}

func 钳制结构体字段(v reflect.Value) {
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		字段 := v.Field(i)
		字段类型 := t.Field(i)
		if 字段类型.Anonymous && 字段.Kind() == reflect.Struct {
			钳制结构体字段(字段) //递归处理匿名嵌入(如request.List)
			continue
		}
		if 字段.Kind() != reflect.Int && 字段.Kind() != reflect.Int64 {
			continue
		}
		if !字段.CanSet() {
			continue
		}
		switch 字段类型.Name {
		case "Page", "page":
			if 字段.Int() < 1 {
				字段.SetInt(1)
			}
		case "Size", "size":
			if 字段.Int() <= 0 {
				字段.SetInt(10)
			} else if 字段.Int() > 10000 {
				字段.SetInt(10000)
			}
		}
	}
}
