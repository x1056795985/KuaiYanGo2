package lastTimeBuffer

import (
	"github.com/gin-gonic/gin"
	"server/app/global"
	"server/app/service"
	"sync"
)

// 集_心跳集合 本机内存收集本回写周期内有心跳的在线id,去重
// 用 map[int]struct{} 而非 H缓存 存 N 个 key:
//   - 写入/去重 O(1),回写只扫这一个集合,不用 Range 全缓存再过滤前缀
//   - 回写后一次 make 清空,不用逐 key Delete
//   - 后期切 Redis:把 map 换成 Redis Set,X心跳_记录→SADD,X心跳_回写→SMEMBERS+DEL,接口不变
var 集_心跳集合 = struct {
	sync.Mutex
	m map[int]struct{}
}{m: make(map[int]struct{})}

// X心跳_记录 记录某在线Id在本次回写周期内有心跳活动
// 同一 id 多次请求只写同一个 map key,天然去重
func X心跳_记录(Id int) {
	if Id <= 0 {
		return
	}
	集_心跳集合.Lock()
	集_心跳集合.m[Id] = struct{}{}
	集_心跳集合.Unlock()
}

// X心跳_回写 定时任务(每30秒)调用:把本周期有心跳的id统一用当前时间戳批量UPDATE db_links_Token.LastTime
// 失败时把id放回集合下次重试,避免丢失心跳更新
func X心跳_回写() {
	if global.GVA_DB == nil {
		return
	}
	// 1. 加锁取出id并清空集合(重新分配map),锁外执行DB更新避免持锁阻塞中间件写入
	集_心跳集合.Lock()
	if len(集_心跳集合.m) == 0 {
		集_心跳集合.Unlock()
		return
	}
	局_ids := make([]int, 0, len(集_心跳集合.m))
	for 局_id := range 集_心跳集合.m {
		局_ids = append(局_ids, 局_id)
	}
	集_心跳集合.m = make(map[int]struct{})
	集_心跳集合.Unlock()

	// 2. 一条UPDATE批量更新,统一用当前时间戳,WHERE Id IN 限定锁范围
	局_db := *global.GVA_DB
	// 复用 service 层,遵循 db := *global.GVA_DB 约定
	局_c := &gin.Context{}

	err := service.NewLinksToken(局_c, &局_db).G更新最后活动时间(局_ids)
	if err != nil {
		if global.GVA_LOG != nil {
			global.GVA_LOG.Println("在线心跳LastTime批量回写失败:" + err.Error())
		}
		// 失败:把id放回集合,下次重试
		集_心跳集合.Lock()
		for _, 局_id := range 局_ids {
			集_心跳集合.m[局_id] = struct{}{}
		}
		集_心跳集合.Unlock()
	}

}
