package round_robin

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"sync/atomic"
)

type Balancer struct {
	conns []balancer.SubConn
	idx   uint64
	cLen  int
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(b.conns) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	nIdx := atomic.AddUint64(&b.idx, 1)
	sc := b.conns[int(nIdx)%b.cLen]
	return balancer.PickResult{
		SubConn: sc,
		Done: func(info balancer.DoneInfo) {

		},
	}, nil
}

type BalancerBuilder struct {
}

func (b BalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]balancer.SubConn, 0, len(info.ReadySCs))

	for sc := range info.ReadySCs {
		conns = append(conns, sc)
	}

	return &Balancer{
		conns: conns,
		idx:   0,
		cLen:  len(conns),
	}
}
