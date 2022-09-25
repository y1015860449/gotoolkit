package coPool

import (
	"errors"
	"sync/atomic"
	"time"
)

const taskChanCap = 100

type worker struct {
	pool      *Pool
	tasks     chan *task
	runTime   time.Time
	taskCount int32
	id        int64
}

func newWork(id int64, pool *Pool) *worker {
	w := &worker{
		id:    id,
		pool:  pool,
		tasks: make(chan *task, taskChanCap),
	}
	go w.run()
	return w
}

func (w *worker) run() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				w.pool.removeBusyWork(w)
			}
		}()
		for {
			t := <-w.tasks
			// 结束协程
			if t == nil {
				w.pool.removeBusyWork(w)
				return
			}
			_ = t.Execute()
			w.updateTaskCount(-1)
			if atomic.LoadInt32(&w.taskCount) <= 0 {
				if !w.pool.revertWorker(w) {
					w.pool.removeBusyWork(w)
					return
				} else {
					w.runTime = time.Now()
				}
			}
		}
	}()
}

func (w *worker) addTask(t *task) error {
	if atomic.LoadInt32(&w.taskCount) >= taskChanCap {
		return errors.New("worker too busy")
	}
	w.tasks <- t
	if t != nil {
		w.updateTaskCount(1)
	}
	return nil
}

func (w *worker) lastRunTime() time.Time {
	return w.runTime
}

func (w *worker) updateTaskCount(delta int32) {
	atomic.AddInt32(&w.taskCount, delta)
	w.pool.updateTaskCount(delta)
}

func (w *worker) IsFull() bool {
	return atomic.LoadInt32(&w.taskCount) >= taskChanCap
}
