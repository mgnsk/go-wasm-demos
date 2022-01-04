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

// NewLock creates a priority-assigned lock. The returned locker
// must only be locked and unlocked once.
func (l *PriorityLocker) NewLock(priority uint64) sync.Locker {
	return locker{
		priority: priority,
		pl:       l,
	}
}

type locker struct {
	priority uint64
	pl       *PriorityLocker
}

func (l locker) Lock() {
	var waiters []func()
	l.pl.internalMu.Lock()
	// Get all waitgroups up to this priority included.
	for i := atomic.LoadUint64(&l.pl.lowestPriority); i <= l.priority; i++ {
		wg, ok := l.pl.waitgroups[i]
		if !ok {
			// Waitgroup doesn't exist yet, create it.
			wg = &sync.WaitGroup{}
			wg.Add(1)
			l.pl.waitgroups[i] = wg
		}
		waiters = append(waiters, wg.Wait)
	}
	l.pl.internalMu.Unlock()

	// Wait for them.
	for _, wait := range waiters {
		wait()
	}

	// It is safe to remove all lower waitgroups since all are done.
	l.pl.internalMu.Lock()
	for pr := range l.pl.waitgroups {
		if pr <= l.priority {
			delete(l.pl.waitgroups, pr)
		}
	}
	l.pl.internalMu.Unlock()

	// Now lock the lock.
	l.pl.lock.Lock()
}

func (l locker) Unlock() {
	l.pl.lock.Unlock()

	l.pl.internalMu.RLock()
	// Unblock callers with higher priority.
	l.pl.waitgroups[l.priority].Done()
	l.pl.internalMu.RUnlock()

	// The immediate next caller, increase lowestPriority as we have
	// unlocked everything up to this index.
	if l.priority == atomic.LoadUint64(&l.pl.lowestPriority)+1 {
		atomic.AddUint64(&l.pl.lowestPriority, 1)
	}
}
