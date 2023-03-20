package random

import (
	"github.com/uzziahlin/micro/router"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
	"math/rand"
)

type WeightBalancer struct {
	conns  []*WeightConn
	filter router.Filter
}

func (w WeightBalancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {

	scs := make([]*WeightConn, 0, len(w.conns))

	total := 0

	for _, c := range w.conns {
		if w.filter(info, c.addr) {
			scs = append(scs, c)
			total += c.weight
		}
	}

	if len(scs) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	r := rand.Intn(total + 1)

	for _, c := range scs {
		r = r - c.weight
		if r < 0 {
			return balancer.PickResult{
				SubConn: c.conn,
				Done: func(info balancer.DoneInfo) {

				},
			}, nil
		}
	}

	panic("inaccessible code")
}

type WeightBalancerBuilder struct {
	Filter router.Filter
}

func (w WeightBalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*WeightConn, 0, len(info.ReadySCs))

	for sc, info := range info.ReadySCs {
		w := info.Address.Attributes.Value("weight").(int)

		conns = append(conns, &WeightConn{
			conn:   sc,
			weight: w,
			addr:   info.Address,
		})
	}

	var filter router.Filter = func(info balancer.PickInfo, addr resolver.Address) bool {
		return true
	}

	if w.Filter != nil {
		filter = w.Filter
	}

	return &WeightBalancer{
		conns:  conns,
		filter: filter,
	}
}

type WeightConn struct {
	conn   balancer.SubConn
	weight int
	addr   resolver.Address
}
