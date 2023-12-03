package class

import "sync"

// 读共享,写独享
type L_读写锁 struct {
	lock sync.RWMutex
}

func (l *L_读写锁) K开始读() {
	l.lock.RLock()
}
func (l *L_读写锁) J结束读() {
	l.lock.RUnlock()
}
func (l *L_读写锁) K开始写() {
	l.lock.Lock()
}
func (l *L_读写锁) J结束写() {
	l.lock.Unlock()
}
