package ringQueue

import (
	"fmt"
	"testing"
)

func TestRingQueue(t *testing.T) {
	queue := InitRingQueue(33)
	for i := 0; i < 54; i++ {
		elem := fmt.Sprintf("test == %d", i)
		queue.Enqueue(elem)
	}

	for i := 0; i < 100; i++ {
		elem, err := queue.Dequeue()
		if err != nil {
			t.Logf("err (%+v)", err)
			break
		}
		t.Logf("elem(%v) len(%v) cap(%v)", elem, queue.len, queue.cap)
	}
}

func TestRingQueue22(t *testing.T) {
	type tmpStr struct {
		key   string
		value int
	}
	queue := InitRingQueue(17)
	for i := 0; i < 54; i++ {
		elem := &tmpStr{
			key:   fmt.Sprintf("test%d", i),
			value: i,
		}
		queue.Enqueue(elem)
		t.Logf("elem(%+v) len(%v) cap(%v) front(%v) tail(%v)", elem, queue.len, queue.cap, queue.front, queue.tail)
	}
	t.Log("\n\n\n\n")
	for i := 0; i < 100; i++ {
		elem, err := queue.Dequeue()
		if err != nil {
			t.Logf("err (%+v)", err)
			break
		}
		//tmp := elem.(*tmpStr)
		t.Logf("elem(%+v) len(%v) cap(%v) front(%v) tail(%v)", elem, queue.len, queue.cap, queue.front, queue.tail)
	}
}
