package cron

import (
	"github.com/robfig/cron/v3"
	"server/new/app/models/db"
	"sync"
)

// 为集群情况,设计的定时任务 自动同步任务,同任务,不会重复处理,多次强占任务失败自动校准时间
type D定时任务 struct {
	Cron        cron.Cron
	Map本机任务列表   map[string]cron.EntryID //刷新不会清除
	Map集群任务列表   map[string]D定时任务_集群节点   //刷新会清除,  集群存数据库id为键名
	Map集群任务Hash string                  //所有任务id+状态+cron 集合的md5    redis 也存储一个这个值,只要更新数据库,就会变,然后判断是否刷新本机的集群任务列表
	L临界         sync.Mutex              //集群刷新任务列表使用,防止线程冲突
	Q抢占任务失败次数   int
}

type D定时任务_集群节点 struct {
	cron.EntryID
	Job Job
}

func (c *D定时任务) Init() *D定时任务 {
	//放弃指针操作方式,
	c.Cron = *cron.New(cron.WithSeconds()) //这里设置的 6位表达式 秒级
	c.Map本机任务列表 = map[string]cron.EntryID{}
	c.Map集群任务列表 = map[string]D定时任务_集群节点{}
	return c
}
func (c *D定时任务) T添加本机任务(任务名称, 表达式 string, cmd func()) error {

	EntryID, err := c.Cron.AddFunc(表达式, cmd)
	if err == nil {
		c.Map本机任务列表[任务名称] = EntryID
	} else {
		return err
	}
	return nil
}

func (c *D定时任务) T添加集群任务(任务数据 db.DB_Cron, 函数 func(执行时间戳 int64, 任务数据 db.DB_Cron)) error {
	Job1 := Job{
		R任务数据: 任务数据,
		H函数:   函数,
	}
	//mark 待验证&Job是否会发生内存溢出的情况, 比如,不断添加,然后不断删除在添加任务,这个Job会被不断创建,但是不知道删除定时任务后是否会被回收内存
	EntryID, err := c.Cron.AddJob(任务数据.Cron, Job1)
	if err == nil {
		c.Map集群任务列表[任务数据.Name] = D定时任务_集群节点{EntryID, Job1}
	} else {
		return err
	}
	return nil
}
