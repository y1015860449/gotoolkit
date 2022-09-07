package ringBuffer

import (
	"encoding/binary"
	"testing"
)

func TestRingBuffer(t *testing.T) {
	ring := NewBuffer(46)
	for i := 0; i < 111; i++ {
		data := make([]byte, 2, 2)
		binary.BigEndian.PutUint16(data, uint16(i))
		ring.WriteBuffer(data)
	}
	t.Logf("len(%v) cap(%v) front(%v) tail(%v)", ring.len, ring.cap, ring.rd, ring.wr)
	t.Logf("uint8(%v)", ring.PeekUint8())
	t.Logf("uint16(%v)", ring.PeekUint16())
	t.Logf("uint32(%v)", ring.PeekUint32())
	t.Logf("uint64(%v)", ring.PeekUint64())
	t.Logf("buffer(%v)", ring.GetBuffers())
	data := make([]byte, 22)
	ring.ReadBuffer(data)
	t.Logf("read(%v)", data)
	t.Logf("len(%v) cap(%v) front(%v) tail(%v)", ring.len, ring.cap, ring.rd, ring.wr)
	t.Logf("buffer(%v)", ring.ReadAllBuffer())
	t.Logf("len(%v) cap(%v) front(%v) tail(%v)", ring.len, ring.cap, ring.rd, ring.wr)
}
