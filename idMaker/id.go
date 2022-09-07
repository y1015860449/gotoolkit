package idMaker

import (
	"github.com/y1015860449/gotoolkit/utils"
	"sync"
	"time"
)

const (
	MaxTick       = 1024
	Epoch   int64 = 1644768000000 // 2022-02-14 00:00:00
)

var (
	idMtx       sync.Mutex
	idTick      int64
	idTimestamp int64
	svcNode     int64
)

func init() {
	tmp := utils.GetUUID()
	hash := utils.Hash64([]byte(tmp))
	svcNode = int64(hash % 8192)
}

/*
GenerateId  可使用到 2056 年
|---------------64位----------------------|
|-1-|----40---------|----13-----|----10---|
|填充|--时间戳差值-----|-自定填充---|--自增数--|
*/
func GenerateId() int64 {
	idMtx.Lock()
	defer idMtx.Unlock()

RETRY:
	now := utils.GetMillisecond()
	if idTimestamp == now {
		idTick++
		if idTick > MaxTick {
			time.Sleep(time.Duration(1) * time.Millisecond)
			goto RETRY
		}
	} else {
		idTick = 0
	}
	idTimestamp = now
	return (now-Epoch)<<23 | svcNode<<10 | idTick
}
