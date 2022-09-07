package ringBuffer

import (
	"encoding/binary"
	"errors"
)

type RingBuffer struct {
	buffer  []byte
	initCap int // 初始
	cap     int
	len     int
	rd      int
	wr      int
}

func NewBuffer(size int) *RingBuffer {
	return &RingBuffer{
		buffer:  make([]byte, size, size),
		initCap: size,
		cap:     size,
		len:     0,
		rd:      0,
		wr:      0,
	}
}

// Length 当前buffer长度
func (b *RingBuffer) Length() int {
	return b.len
}

// Capacity 当前buffer容量
func (b *RingBuffer) Capacity() int {
	return b.cap
}

// IsEmpty 判断当前buffer长度是否为空
func (b *RingBuffer) IsEmpty() bool {
	return b.len == 0
}

// IsFull 判断当前buffer是否已满
func (b *RingBuffer) IsFull() bool {
	return b.len >= b.cap
}

// ReadBuffer 读取buffer
func (b *RingBuffer) ReadBuffer(data []byte) (int, error) {
	if len(data) <= 0 {
		return 0, nil
	}
	if b.IsEmpty() {
		return 0, errors.New("buffer is empty")
	}
	dataLen := len(data)
	if dataLen > b.len {
		dataLen = b.len
	}
	if b.wr > b.rd {
		copy(data, b.buffer[b.rd:b.rd+dataLen])
	} else {
		if b.rd+dataLen <= b.cap {
			copy(data, b.buffer[b.rd:b.rd+dataLen])
		} else {
			// front
			copy(data, b.buffer[b.rd:b.cap])
			// tail
			copy(data[b.cap-b.rd:], b.buffer[0:dataLen-b.cap+b.rd])
		}
	}
	b.rd = (b.rd + dataLen) % b.cap
	b.len = b.len - dataLen
	return dataLen, nil
}

func (b *RingBuffer) ReadAllBuffer() []byte {
	data := b.GetBuffers()
	b.rd = (b.rd + b.len) % b.cap
	b.len = 0
	return data
}

// WriteBuffer 写入buffer
func (b *RingBuffer) WriteBuffer(data []byte) (int, error) {
	if len(data) <= 0 {
		return 0, nil
	}
	dataLen := len(data)
	freeLen := b.cap - b.len
	if freeLen < dataLen {
		b.expandCapacity(dataLen - freeLen)
	}
	if b.wr >= b.rd {
		if b.cap-b.wr >= dataLen {
			copy(b.buffer[b.wr:], data)
		} else {
			copy(b.buffer[b.wr:], data[0:b.cap-b.wr])
			copy(b.buffer[0:], data[b.cap-b.wr:])
		}
	} else {
		copy(b.buffer[b.wr:], data)
	}
	b.len = b.len + dataLen
	b.wr = (b.wr + dataLen) % b.cap
	return dataLen, nil
}

func (b *RingBuffer) expandCapacity(len int) {
	newCap := 0
	for newCap < len+b.cap {
		if b.cap < 1024 {
			newCap = newCap + 2*b.cap
		} else {
			newCap = newCap + b.cap/2
		}
	}
	b.makeCapacity(newCap)
}

func (b *RingBuffer) makeCapacity(cap int) {
	newBuffer := make([]byte, cap, cap)
	bufLen := b.len
	_, _ = b.ReadBuffer(newBuffer)
	b.buffer = newBuffer
	b.wr = bufLen
	b.rd = 0
	b.len = bufLen
	b.cap = cap
}

// RetrieveAll 移动位置到初始值
func (b *RingBuffer) RetrieveAll() {
	b.rd = 0
	b.wr = 0
	b.len = 0
}

// Retrieve 移动读取位置
func (b *RingBuffer) Retrieve(len int) {
	if b.IsEmpty() || len <= 0 {
		return
	}
	if len < b.len {
		b.rd = (b.rd + len) % b.cap
		b.len = b.len - len
	} else {
		b.RetrieveAll()
	}
}

// PeekBuffer 预读buffer，不移动读取位置
func (b *RingBuffer) PeekBuffer(len int) ([]byte, []byte, int) {
	if b.IsEmpty() || len <= 0 {
		return nil, nil, 0
	}
	var (
		first  []byte = nil
		second []byte = nil
	)
	if len > b.len {
		len = b.len
	}
	if b.rd < b.wr {
		first = b.buffer[b.rd : b.rd+len]
	} else {
		if b.rd+len <= b.cap {
			first = b.buffer[b.rd : b.rd+len]
		} else {
			// front
			first = b.buffer[b.rd:b.cap]
			// tail
			second = b.buffer[0 : len-b.cap+b.rd]
		}
	}
	return first, second, len
}

// PeekAllBuffer 预读buffer，不移动读取位置
func (b *RingBuffer) PeekAllBuffer() ([]byte, []byte) {
	if b.IsEmpty() {
		return nil, nil
	}
	var (
		first  []byte = nil
		second []byte = nil
	)

	if b.rd < b.wr {
		first = b.buffer[b.rd:b.wr]
	} else {
		// front
		first = b.buffer[b.rd:b.cap]
		// tail
		second = b.buffer[0:b.wr]
	}
	return first, second
}

func (b *RingBuffer) PeekUint8() uint8 {
	if b.IsEmpty() {
		return 0
	}
	f, s, _ := b.PeekBuffer(1)
	if len(f) > 0 {
		return f[0]
	} else {
		return s[0]
	}
}

func (b *RingBuffer) PeekUint16() uint16 {
	if b.len < 2 {
		return 0
	}
	f, s, _ := b.PeekBuffer(2)
	if len(s) > 0 {
		f = append(f, s...)
	}
	return binary.BigEndian.Uint16(f)
}

func (b *RingBuffer) PeekUint32() uint32 {
	if b.len < 4 {
		return 0
	}
	f, s, _ := b.PeekBuffer(4)
	if len(s) > 0 {
		f = append(f, s...)
	}
	return binary.BigEndian.Uint32(f)
}

func (b *RingBuffer) PeekUint64() uint64 {
	if b.len < 8 {
		return 0
	}
	f, s, _ := b.PeekBuffer(8)
	if len(s) > 0 {
		f = append(f, s...)
	}
	return binary.BigEndian.Uint64(f)
}

// GetBuffers 返回buffer所有可读数据, 不移动读位置，仅仅是拷贝全部数据
func (b *RingBuffer) GetBuffers() []byte {
	if b.IsEmpty() {
		return nil
	}
	buf := make([]byte, b.len)
	if b.wr > b.rd {
		copy(buf, b.buffer[b.rd:b.wr])
	} else {
		copy(buf, b.buffer[b.rd:b.cap])
		copy(buf[b.cap-b.rd:], b.buffer[0:b.wr])
	}
	return buf
}
