package broadcast

import (
	"context"
	"github.com/uzziahlin/micro/registry"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type ClusterBuilder struct {
	serviceName string
	registry    registry.Registry
	dialOpts    []grpc.DialOption
}

func NewClusterBuilder(r registry.Registry, service string, dialOptions ...grpc.DialOption) *ClusterBuilder {
	return &ClusterBuilder{
		registry:    r,
		serviceName: service,
		dialOpts:    dialOptions,
	}
}

func (c ClusterBuilder) BuildUnaryInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if !IsBroadcast(ctx) {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		instances, err := c.registry.Discover(ctx, c.serviceName)

		if err != nil {
			return err
		}

		var eg errgroup.Group
		for _, instance := range instances {
			addr := instance.Addr
			eg.Go(func() error {
				insCC, er := grpc.Dial(addr, c.dialOpts...)
				if er != nil {
					return er
				}
				return invoker(ctx, method, req, reply, insCC, opts...)
			})
		}

		return eg.Wait()
	}
}

type broadcastKey struct {
}

func UseBroadcast(p context.Context) context.Context {
	return context.WithValue(p, broadcastKey{}, true)
}

func IsBroadcast(ctx context.Context) bool {
	val, ok := ctx.Value(broadcastKey{}).(bool)

	return ok && val
}
