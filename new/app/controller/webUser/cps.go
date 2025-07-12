package controller

import (
	"github.com/gin-gonic/gin"
	"server/new/app/controller/Common"
	"server/new/app/controller/Common/response"
)

type Cps struct {
	Common.Common
}

func NewCpsController() *Cps {
	return &Cps{}
}

/*
手动设置邀请码  解决客户已经注册,但是邀请码没有的情况
不同应用可能是不同的邀请者,要区分开来, 用户在该应用充值时,给对应的邀请者
还有种情况,账号模式应用 a 第一次邀请的c注册应用1,这时候,b就无法邀请注册了,但是又确实是b邀请的注册应用2,只能让C登陆后,手动填写邀请人
*/
func (C *Cps) SetVisitRelation(c *gin.Context) {
	//mark待完善
	response.Ok(c)
	return
}
