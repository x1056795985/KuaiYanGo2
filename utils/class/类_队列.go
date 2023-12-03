package class

import (
	"container/list"
	"fmt"
	"sync"
)

// 先入先出队列,线程安全
type L_队列 struct {
	l    sync.Mutex
	data list.List
}

func (j *L_队列) init() {
	j.data.Init()
}

func (q *L_队列) J加入队列(v interface{}) {
	q.l.Lock()
	defer q.l.Unlock()
	q.data.PushFront(v)
}

func (q *L_队列) T弹出队列() interface{} {
	q.l.Lock()
	defer q.l.Unlock()
	iter := q.data.Back()
	if nil == iter {
		return nil
	}
	v := iter.Value
	q.data.Remove(iter)
	return v
}
func (q *L_队列) T弹出队列文本(值 *string) bool {
	q.l.Lock()
	defer q.l.Unlock()
	iter := q.data.Back()
	if nil == iter {
		return false
	}
	v := iter.Value
	局_临时, ok := v.(string)
	if !ok {
		return false
	}
	q.data.Remove(iter)
	*值 = 局_临时
	return true
}

func (q *L_队列) Q取队列长度() int {
	return q.data.Len()
}

func (q *L_队列) Q清空队列() interface{} {
	q.l.Lock()
	defer q.l.Unlock()
	return q.data.Init()
}

func (q *L_队列) Dump() {
	for iter := q.data.Back(); iter != nil; iter = iter.Prev() {
		fmt.Println("item:", iter.Value)
	}
}
