// 自动扩缩容的环形队列

package ringQueue

import "errors"

type RingQueue struct {
	queues  []interface{}
	front   int // 队头
	tail    int // 队尾
	len     int // 队列长度
	cap     int // 队列容量
	initCap int // 初始化的队列容量
}

// InitRingQueue 初始化环形队列
func InitRingQueue(size int) *RingQueue {
	ring := &RingQueue{
		queues:  make([]interface{}, size, size),
		front:   0,
		tail:    0,
		len:     0,
		cap:     size,
		initCap: size,
	}
	return ring
}

// Length 当前队列长度
func (p *RingQueue) Length() int {
	return p.len
}

// Capacity 当前队列容量
func (p *RingQueue) Capacity() int {
	return p.cap
}

// IsEmpty 判断当前队列长度是否为空
func (p *RingQueue) IsEmpty() bool {
	return p.len == 0
}

// IsFull 判断当前队列是否已满
func (p *RingQueue) IsFull() bool {
	return p.len >= p.cap
}

// Enqueue 入队
func (p *RingQueue) Enqueue(elem interface{}) {
	if p.IsFull() {
		newCap := 2 * p.cap
		if p.cap >= 1024 {
			newCap = p.cap + p.cap/2
		}
		p.makeCapacity(newCap)
	}
	p.queues[p.tail] = elem
	p.tail = (p.tail + 1) % p.cap
	p.len++
}

// Dequeue 出队
func (p *RingQueue) Dequeue() (interface{}, error) {
	if p.IsEmpty() {
		return nil, errors.New("queue is empty")
	}
	// 当队列长度小于队列1/4容量 且 队列容量大于队列初始化容量
	if p.len <= p.cap/4 && p.cap > p.initCap {
		newCap := p.initCap
		if p.cap/2 >= p.initCap {
			newCap = p.cap / 2
		}
		p.makeCapacity(newCap)
	}
	elem := p.queues[p.front]
	p.front = (p.front + 1) % p.cap
	p.len--
	return elem, nil
}

func (p *RingQueue) makeCapacity(newCap int) {
	queues := make([]interface{}, newCap, newCap)
	if p.tail > p.front {
		copy(queues, p.queues[p.front:p.tail])
	} else {
		if p.front+p.len <= p.cap {
			copy(queues, p.queues[p.front:p.front+p.len])
		} else {
			// front
			copy(queues, p.queues[p.front:p.cap])
			// tail
			copy(queues[p.cap-p.front:], p.queues[0:p.tail])
		}
	}
	p.queues = queues
	p.front = 0
	p.tail = p.len
	p.cap = newCap
}
