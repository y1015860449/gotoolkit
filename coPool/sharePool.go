package coPool

import (
	"context"
	"errors"
	"github.com/y1015860449/gotoolkit/idMaker"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type SharePool struct {
	// 超过空闲时间清楚空闲的协程
	expiryDuration time.Duration
	// 最大协程数
	capacity int32
	// 最小协程数
	min int32
	// 当前协程数
	current int32
	// 当前写正在运行的所有协程
	workers map[int64]*shareWorker
	lock    sync.Mutex
	// 任务队列
	tasks    chan *task
	taskSize int32

	ctx    context.Context
	cancel func()
}

func NewSharePool(size, min, taskSize int32, expire time.Duration) *SharePool {
	p := &SharePool{
		capacity:       size,
		expiryDuration: expire,
		min:            min,
		taskSize:       taskSize,
		workers:        make(map[int64]*shareWorker),
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
	if taskSize <= 0 {
		p.taskSize = 1024
	}
	p.ctx, p.cancel = context.WithCancel(context.Background())
	p.tasks = make(chan *task, p.taskSize)
	for i := int32(0); i < min; i++ {
		id := idMaker.GenerateId()
		w := newShareWork(id, p)
		p.workers[id] = w
	}
	atomic.SwapInt32(&p.current, min)
	go p.periodicallyPurgeWorker()
	return p
}

func (p *SharePool) Close() {
	if atomic.LoadInt32(&p.current) > 0 {
		atomic.SwapInt32(&p.current, 0)
		p.cancel()
		p.workers = make(map[int64]*shareWorker)
	}
}

func (p *SharePool) Submit(data interface{}, f func(interface{}) error) error {
	t := &task{
		param:      data,
		handleFunc: f,
	}
	return p.AddTask(t)
}

func (p *SharePool) AddTask(t *task) error {
	for {
		select {
		case p.tasks <- t:
			return nil
		default:
			if atomic.LoadInt32(&p.current) > p.capacity {
				return errors.New("task full")
			}
			id := idMaker.GenerateId()
			w := newShareWork(id, p)
			p.lock.Lock()
			p.workers[id] = w
			p.lock.Unlock()
			atomic.AddInt32(&p.current, 1)
			time.Sleep(500 * time.Microsecond)
		}
	}
}

// 定期清除空闲协程
func (p *SharePool) periodicallyPurgeWorker() {
	heartbeat := time.NewTicker(p.expiryDuration)
	defer heartbeat.Stop()

	for range heartbeat.C {
		select {
		case <-p.ctx.Done():
			return
		default:
			var expiredWorkers []*shareWorker
			currentTime := time.Now()
			if atomic.LoadInt32(&p.current) <= p.min {
				return
			}
			p.lock.Lock()
			for i, w := range p.workers {
				if currentTime.Sub(w.lastRunTime()) > p.expiryDuration {
					expiredWorkers = append(expiredWorkers, w)
					delete(p.workers, i)
					atomic.AddInt32(&p.current, -1)
				}
				if atomic.LoadInt32(&p.current) <= p.min {
					break
				}
			}
			p.lock.Unlock()
			for _, w := range expiredWorkers {
				log.Printf("timer close")
				w.done()
			}
			expiredWorkers = expiredWorkers[0:0]
		}
	}
}
