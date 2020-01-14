package audio

import (
	"sync"
	"sync/atomic"
)

// PriorityLocker hands out lock/unlock wrappers to a single mutex that when locked,
// first waits until the lock has been locked and unlocked by all callers with lower priority
// before attempting to lock it.
type PriorityLocker struct {
	lock sync.Mutex

	lowestPriority uint64
	internalMu     sync.RWMutex
	waitgroups     map[uint64]*sync.WaitGroup
}

// NewPriorityLocker constructor.
func NewPriorityLocker(lowestPriority uint64) *PriorityLocker {
	return &PriorityLocker{
		lowestPriority: lowestPriority,
		waitgroups:     make(map[uint64]*sync.WaitGroup),
	}
}

// GetLock registers a priority-assigned lock.
func (l *PriorityLocker) GetLock(priority uint64) (lock, unlock func()) {
	lock = func() {
		var waiters []func()
		l.internalMu.Lock()
		// Get all waitgroups up to this priority included.
		for i := atomic.LoadUint64(&l.lowestPriority); i <= priority; i++ {
			wg, ok := l.waitgroups[i]
			if !ok {
				// Waitgroup doesn't exist yet, create it.
				wg = &sync.WaitGroup{}
				wg.Add(1)
				l.waitgroups[i] = wg
			}
			waiters = append(waiters, wg.Wait)
		}
		l.internalMu.Unlock()

		// Wait for them.
		for _, wait := range waiters {
			wait()
		}

		// It is safe to remove all lower waitgroups since all are done.
		l.internalMu.Lock()
		for pr := range l.waitgroups {
			if pr <= priority {
				delete(l.waitgroups, pr)
			}
		}
		l.internalMu.Unlock()

		// Now lock the lock.
		l.lock.Lock()
	}

	unlock = func() {
		l.lock.Unlock()

		l.internalMu.RLock()
		// Unblock callers with higher priority.
		l.waitgroups[priority].Done()
		l.internalMu.RUnlock()

		// The immediate next caller, increase lowestPriority as we have
		// unlocked everything up to this index.
		if priority == atomic.LoadUint64(&l.lowestPriority)+1 {
			atomic.AddUint64(&l.lowestPriority, 1)
		}
	}

	return
}
