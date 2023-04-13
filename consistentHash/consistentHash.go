package consistentHash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type ConsistentHash struct {
	hashFunc    Hash           // hash函数
	virtualNode int            // 虚拟节点倍数
	hashRing    []int          // 哈希环
	hashNodes   map[int]string // 节点与key的映射表
}

func NewConsistentHash(virtualNode int, fn Hash) *ConsistentHash {
	m := &ConsistentHash{
		virtualNode: virtualNode,
		hashFunc:    fn,
		hashNodes:   make(map[int]string),
	}
	if m.hashFunc == nil {
		m.hashFunc = crc32.ChecksumIEEE
	}
	return m
}

// 添加节点
func (m *ConsistentHash) AddNodes(keys ...string) {
	// 对一个物理节点添加多个虚拟节点
	for _, key := range keys {
		for i := 0; i < m.virtualNode; i++ {
			hash := int(m.hashFunc([]byte(strconv.Itoa(i) + key)))
			m.hashRing = append(m.hashRing, hash)
			m.hashNodes[hash] = key
		}
	}
	// 对哈希环排序
	sort.Ints(m.hashRing)
}

// 删除节点
func (m *ConsistentHash) DeleteNodes(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.virtualNode; i++ {
			hash := int(m.hashFunc([]byte(strconv.Itoa(i) + key)))
			delete(m.hashNodes, hash)
		}
	}
	// 重新建立哈希环
	m.hashRing = m.hashRing[:0]
	for hs := range m.hashNodes {
		m.hashRing = append(m.hashRing, hs)
	}
	sort.Ints(m.hashRing)
}

func (m *ConsistentHash) Get(key string) string {
	if len(m.hashRing) == 0 {
		return ""
	}
	hash := int(m.hashFunc([]byte(key)))
	// 顺时针方向遍历哈希环，找到第一个哈希值比 key 的哈希值大的节点，返回该节点
	idx := sort.Search(len(m.hashRing), func(i int) bool { return m.hashRing[i] >= hash })
	if idx == len(m.hashRing) {
		idx = 0
	}
	return m.hashNodes[m.hashRing[idx]]
}
