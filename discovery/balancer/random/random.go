package random

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/grpclog"
	"math/rand"
	"sync"
	"time"
)

// Name is the name of random balancer.
const Name = "bln_random"

var logger = grpclog.Component("random")

// newBuilder creates a new random balancer builder.
func newBuilder() balancer.Builder {
	return base.NewBalancerBuilder(Name, &randomPickerBuilder{}, base.Config{HealthCheck: true})
}

func init() {
	balancer.Register(newBuilder())
	rand.Seed(time.Now().UnixNano())
}

type randomPickerBuilder struct{}

func (*randomPickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	logger.Infof("randomPicker: Build called with info: %v", info)
	if len(info.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}
	scs := make([]balancer.SubConn, 0, len(info.ReadySCs))
	for sc := range info.ReadySCs {
		scs = append(scs, sc)
	}
	return &randomPicker{
		subConns: scs,
	}
}

type randomPicker struct {
	subConns []balancer.SubConn
	mu       sync.Mutex
}

func (p *randomPicker) Pick(balancer.PickInfo) (balancer.PickResult, error) {
	p.mu.Lock()
	length := len(p.subConns)
	sc := p.subConns[rand.Int()%length]
	p.mu.Unlock()
	return balancer.PickResult{SubConn: sc}, nil
}
