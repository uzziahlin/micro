package random

import (
	"github.com/uzziahlin/micro/router"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
	"math/rand"
)

type Balancer struct {
	conns  []*Conn
	filter router.Filter
}

func (b Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {

	scs := make([]balancer.SubConn, 0, len(b.conns))

	for _, c := range b.conns {
		ok := b.filter(info, c.addr)
		if ok {
			scs = append(scs, c.conn)
		}
	}

	if len(scs) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	r := rand.Intn(len(scs))

	return balancer.PickResult{
		SubConn: scs[r],
	}, nil
}

type BalancerBuilder struct {
	Filter router.Filter
}

func (b BalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*Conn, 0, len(info.ReadySCs))

	for sc, info := range info.ReadySCs {
		conns = append(conns, &Conn{
			conn: sc,
			addr: info.Address,
		})
	}

	var filter router.Filter = func(info balancer.PickInfo, addr resolver.Address) bool {
		return true
	}

	if b.Filter != nil {
		filter = b.Filter
	}

	return &Balancer{
		conns:  conns,
		filter: filter,
	}
}

type Conn struct {
	conn balancer.SubConn
	addr resolver.Address
}
