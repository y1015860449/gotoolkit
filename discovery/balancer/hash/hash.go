package hash

import (
	"github.com/y1015860449/gotoolkit/utils"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/grpclog"
	"sync"
)

const (
	Name    = "bln_hash"
	HashKey = "hash_key"
)

var logger = grpclog.Component("hash")

// newBuilder creates a new hash balancer builder.
func newBuilder() balancer.Builder {
	return base.NewBalancerBuilder(Name, &hashPickerBuilder{}, base.Config{HealthCheck: true})
}

func init() {
	balancer.Register(newBuilder())
}

type hashPickerBuilder struct{}

func (*hashPickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	logger.Infof("hashPicker: Build called with info: %v", info)
	if len(info.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}
	scs := make([]balancer.SubConn, 0, len(info.ReadySCs))
	for sc := range info.ReadySCs {
		scs = append(scs, sc)
	}
	return &hashPicker{
		subConns: scs,
	}
}

type hashPicker struct {
	subConns []balancer.SubConn
	mu       sync.RWMutex
}

func (p *hashPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	value := info.Ctx.Value(HashKey).(string)
	sum := utils.Hash32([]byte(value))
	p.mu.RLock()
	length := len(p.subConns)
	index := int(sum) % length
	sc := p.subConns[index]
	p.mu.RUnlock()
	return balancer.PickResult{SubConn: sc}, nil
}
