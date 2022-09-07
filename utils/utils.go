package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/spaolacci/murmur3"
	"math/rand"
	"net"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GetUUID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

func GetMillisecond() int64 {
	return time.Now().UnixNano() / 1e6
}

func GetGenerateRandom(length int) string {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	randTab := []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	size := len(randTab)
	bytes := make([]byte, length)
	for i := 0; i < length; i++ {
		bytes[i] = randTab[r.Intn(size)]
	}
	return string(bytes)
}

func Hash64(data []byte) uint64 {
	return murmur3.Sum64(data)
}

func Hash32(data []byte) uint32 {
	return murmur3.Sum32(data)
}

func Md5(data []byte) []byte {
	digest := md5.New()
	digest.Write(data)
	return digest.Sum(nil)
}

// Md5Hex returns the md5 hex string of data.
func Md5Hex(data []byte) string {
	return fmt.Sprintf("%x", Md5(data))
}

// RemoveRepeatInt64Array int64 数字切片去重
func RemoveRepeatInt64Array(arr []int64) []int64 {
	if len(arr) <= 0 {
		return arr
	}
	rest := make([]int64, 0)
	tmp := make(map[int64]interface{})
	for _, v := range arr {
		if _, ok := tmp[v]; !ok {
			rest = append(rest, v)
			tmp[v] = 0
		}
	}
	return rest
}

// SubstrInt64Array int64 数字切片差集
func SubstrInt64Array(src, dst []int64) []int64 {
	rest := make([]int64, 0)
	tmp := make(map[int64]interface{})
	for _, v := range src {
		if _, ok := tmp[v]; !ok {
			tmp[v] = 0
		}
	}
	for _, v := range dst {
		if _, ok := tmp[v]; !ok {
			rest = append(rest, v)
		}
	}
	return rest
}

// DeleteInt64Array 删除 int64 切片
func DeleteInt64Array(arr []int64, del []int64) []int64 {
	rest := make([]int64, 0)
	tmp := make(map[int64]interface{})
	for _, v := range arr {
		tmp[v] = 0
	}
	for _, v := range del {
		delete(tmp, v)
	}

	for k, _ := range tmp {
		rest = append(rest, k)
	}
	return rest
}

// CopyStruct 相同的数据对象赋值
func CopyStruct(src, dst interface{}) {
	sData, _ := json.Marshal(src)
	_ = json.Unmarshal(sData, dst)
}

// RandomInt 获取 int 随机数
func RandomInt(begin, end int) int {
	return rand.Intn(end-begin+1) + begin
}

func RandomInt64() int64 {
	return rand.Int63()
}

// JsonDecoder json解析int64缺精度问题
func JsonDecoder(data []byte, result interface{}) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(result); err != nil {
		return err
	}
	return nil
}

// InternalIP return internal ip.
func InternalIP() string {
	inters, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, inter := range inters {
		if inter.Flags&net.FlagUp != 0 && !strings.HasPrefix(inter.Name, "lo") {
			addrs, err := inter.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						return ipnet.IP.String()
					}
				}
			}
		}
	}
	return ""
}
