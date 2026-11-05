package sync

import "sync"

type Quit struct {
	done  chan struct{}
	mutex sync.RWMutex
	tasks sync.WaitGroup
}

func NewQuit() *Quit {
	return &Quit{
		done: make(chan struct{}),
	}
}

func (q *Quit) Signal() <-chan struct{} {
	return q.done
}

func (q *Quit) Add(delta int) bool {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	select {
	case <-q.done:
		return false
	default:
		q.tasks.Add(delta)
		return true
	}
}

func (q *Quit) Done() {
	q.tasks.Done()
}

func (q *Quit) Wait() bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	select {
	case <-q.done:
		return false
	default:
		close(q.done)
		q.tasks.Wait()
		return true
	}
}

func (q *Quit) Close() bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	select {
	case <-q.done:
		return false
	default:
		close(q.done)
		return true
	}
}
