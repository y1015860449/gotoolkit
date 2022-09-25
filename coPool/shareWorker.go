package coPool

import (
	"context"
	"log"
	"runtime"
	"time"
)

type shareWorker struct {
	pool    *SharePool
	id      int64
	runTime time.Time
	ctx     context.Context
	cancel  func()
}

func newShareWork(id int64, pool *SharePool) *shareWorker {
	w := &shareWorker{
		id:   id,
		pool: pool,
	}
	w.ctx, w.cancel = context.WithCancel(context.Background())
	go w.run()
	return w
}

func (w *shareWorker) run() {
	for {
		select {
		case t := <-w.pool.tasks:
			w.exec(t)
		case <-w.ctx.Done():
			return
		case <-w.pool.ctx.Done():
			return
		}
	}
}

func (w *shareWorker) exec(t *task) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 64*1024)
			buf = buf[:runtime.Stack(buf, false)]
			log.Println(buf)
		}
	}()
	w.runTime = time.Now()
	_ = t.Execute()
}

func (w *shareWorker) done() {
	w.cancel()
}

func (w *shareWorker) lastRunTime() time.Time {
	return w.runTime
}
