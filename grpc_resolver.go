package micro

import (
	"context"
	"github.com/uzziahlin/micro/registry"
	"google.golang.org/grpc/resolver"
	"time"
)

const scheme = "registry"

type RegisterBuilder struct {
	registry registry.Registry
	timeout  time.Duration
}

func (r RegisterBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {

	res := &RegisterResolver{
		cc:          cc,
		registry:    r.registry,
		timeout:     r.timeout,
		serviceName: target.Endpoint(),
	}

	res.ResolveNow(resolver.ResolveNowOptions{})

	go res.watch()

	return res, nil
}

func (r RegisterBuilder) Scheme() string {
	return scheme
}

type RegisterResolver struct {
	cc          resolver.ClientConn
	registry    registry.Registry
	timeout     time.Duration
	serviceName string
	closeC      chan struct{}
}

func (r RegisterResolver) ResolveNow(options resolver.ResolveNowOptions) {
	r.resolve()
}

func (r RegisterResolver) resolve() {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	instances, err := r.registry.Discover(ctx, r.serviceName)

	if err != nil {
		r.cc.ReportError(err)
		return
	}

	addresses := make([]resolver.Address, 0, len(instances))

	for _, instance := range instances {
		addr := resolver.Address{
			Addr: instance.Addr,
		}

		if instance.Weight != 0 {
			addr.Attributes.WithValue("weight", instance.Weight)
		}

		if instance.Group != "" {
			addr.Attributes.WithValue("group", instance.Group)
		}

		addresses = append(addresses, addr)
	}

	err = r.cc.UpdateState(resolver.State{
		Addresses: addresses,
	})

	if err != nil {
		r.cc.ReportError(err)
		return
	}
}

func (r RegisterResolver) watch() {
	watchC := r.registry.Subscribe(r.serviceName)
	for {
		select {
		case <-watchC:
			r.resolve()
		case <-r.closeC:
			return
		}
	}
}

func (r RegisterResolver) Close() {
	r.closeC <- struct{}{}
}
