package class

import "sync"

type L_临界许可 struct {
	lock sync.Mutex
}

func (l *L_临界许可) J进入许可区() {
	l.lock.Lock()
}

func (l *L_临界许可) T退出许可区() {
	l.lock.Unlock()
}

func (l *L_临界许可) C尝试进入() bool {
	return l.lock.TryLock()
}
