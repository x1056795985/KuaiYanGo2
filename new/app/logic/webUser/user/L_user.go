package user

import (
	"github.com/gin-gonic/gin"
)

var L_user user

func init() {
	L_user = user{}

}

type user struct {
	注册后处理 []func(c *gin.Context, AppId, 邀请人, 被邀请人 int, Referer string)
}

func (j *user) T邀请注册成功后处理(c *gin.Context, AppId, 邀请人, 被邀请人 int, Referer string) {
	for _, v := range j.注册后处理 {
		v(c, AppId, 邀请人, 被邀请人, Referer)
	}
	return
}
func (j *user) Z邀请注册成功后处理(v func(c *gin.Context, AppId, 邀请人, 被邀请人 int, Referer string)) {
	j.注册后处理 = append(j.注册后处理, v)
}
