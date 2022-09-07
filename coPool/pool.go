package coPool

import (
	"errors"
	"github.com/y1015860449/gotoolkit/idMaker"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type Pool struct {
	// 当前正在运行的工作协程
	workingCount int32
	// 超过空闲时间清楚空闲的协程
	expiryDuration time.Duration
	// 最大协程数
	capacity int32
	// 最小协程数
	min int32
	// 当前协程数
	current int32
	// 空闲协程
	idleWorkers map[int64]*worker
	// 忙碌协程
	busyWorkers map[int64]*worker
	lock        sync.Mutex
	// 关闭标识
	isClose int32
	// 总任务数
	totalCount int32
	// 当前任务数
	taskCount int32
}

func NewPool(size, min int32, expire time.Duration) *Pool {
	p := &Pool{
		capacity:       size,
		expiryDuration: expire,
		min:            min,
		idleWorkers:    make(map[int64]*worker),
		busyWorkers:    make(map[int64]*worker),
	}
	if min <= 0 {
		p.min = 1
	}
	if size <= 0 {
		p.capacity = 4
	}
	if expire <= 0 {
		p.expiryDuration = 60 * time.Second
	}

	for i := int32(0); i < min; i++ {
		id := idMaker.GenerateId()
		w := newWork(id, p)
		p.busyWorkers[id] = w
	}
	atomic.SwapInt32(&p.current, min)
	atomic.SwapInt32(&p.workingCount, min)
	atomic.AddInt32(&p.totalCount, min*taskChanCap)
	go p.periodicallyPurgeWorker()
	return p
}

func (p *Pool) Close() {
	if atomic.LoadInt32(&p.isClose) == 0 {
		atomic.StoreInt32(&p.isClose, 1)
		p.lock.Lock()
		for _, w := range p.idleWorkers {
			_ = w.addTask(nil)
		}
		p.lock.Unlock()
		return
	}
}

func (p *Pool) Submit(data interface{}, f func(interface{}) error) error {
	t := &task{
		param:      data,
		handleFunc: f,
	}
	return p.AddTask(t)
}

func (p *Pool) AddTask(t *task) error {
	w, err := p.getWorker()
	if err != nil {
		return err
	}
	return w.addTask(t)
}

func (p *Pool) getWorker() (*worker, error) {
	var (
		w   *worker
		err error
	)
	p.lock.Lock()
	if atomic.LoadInt32(&p.taskCount) < atomic.LoadInt32(&p.totalCount) {
		for _, v := range p.busyWorkers {
			if !v.IsFull() {
				w = v
				break
			}
		}
	} else {
		if len(p.idleWorkers) > 0 {
			for k, v := range p.idleWorkers {
				w = v
				p.busyWorkers[k] = v
				delete(p.idleWorkers, k)
				atomic.AddInt32(&p.totalCount, taskChanCap)
				atomic.AddInt32(&p.workingCount, 1)
				break
			}
		} else {
			if atomic.LoadInt32(&p.current) < atomic.LoadInt32(&p.capacity) {
				id := idMaker.GenerateId()
				w = newWork(id, p)
				p.busyWorkers[id] = w
				atomic.AddInt32(&p.totalCount, taskChanCap)
				atomic.AddInt32(&p.workingCount, 1)
				atomic.AddInt32(&p.current, 1)
			} else {
				err = errors.New("pool too busy")
			}
		}
	}
	p.lock.Unlock()
	return w, err
}

// 定期清除空闲协程
func (p *Pool) periodicallyPurgeWorker() {
	heartbeat := time.NewTicker(p.expiryDuration)
	defer heartbeat.Stop()

	var expiredWorkers []*worker
	for range heartbeat.C {
		if atomic.LoadInt32(&p.isClose) == 1 {
			break
		}
		currentTime := time.Now()
		p.lock.Lock()
		idleCount := len(p.idleWorkers)
		if idleCount <= int(p.min) {
			p.lock.Unlock()
			continue
		}
		for i, w := range p.idleWorkers {
			if currentTime.Sub(w.lastRunTime()) > p.expiryDuration {
				expiredWorkers = append(expiredWorkers, w)
				delete(p.idleWorkers, i)
			}
			if len(p.idleWorkers) <= int(p.min) {
				break
			}
		}
		p.lock.Unlock()
		for _, w := range expiredWorkers {
			log.Printf("timer close")
			_ = w.addTask(nil)
			atomic.AddInt32(&p.current, -1)
		}
		expiredWorkers = expiredWorkers[0:0]
	}
}

func (p *Pool) removeBusyWork(w *worker) {
	p.lock.Lock()
	_, ok := p.busyWorkers[w.id]
	if ok {
		delete(p.busyWorkers, w.id)
		atomic.AddInt32(&p.workingCount, -1)
		atomic.AddInt32(&p.totalCount, -1*taskChanCap)
	} else {
		delete(p.idleWorkers, w.id)
	}
	p.lock.Unlock()
	atomic.AddInt32(&p.current, -1)
}

func (p *Pool) revertWorker(w *worker) bool {
	if atomic.LoadInt32(&p.isClose) == 1 {
		return false
	}
	p.lock.Lock()
	p.idleWorkers[w.id] = w
	delete(p.busyWorkers, w.id)
	p.lock.Unlock()
	atomic.AddInt32(&p.workingCount, -1)
	atomic.AddInt32(&p.totalCount, -1*taskChanCap)
	return true
}

func (p *Pool) updateTaskCount(delta int32) {
	atomic.AddInt32(&p.taskCount, delta)
}
