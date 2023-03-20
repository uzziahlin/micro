package broadcastall

import (
	"context"
	"fmt"
	"github.com/uzziahlin/micro/registry"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"reflect"
)

type ClusterBuilder struct {
	registry    registry.Registry
	serviceName string
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
		ok, ch := IsBroadcast(ctx)
		if !ok {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		defer close(ch)

		instances, err := c.registry.Discover(ctx, c.serviceName)
		if err != nil {
			return err
		}

		var eg errgroup.Group
		typ := reflect.TypeOf(reply).Elem()
		for _, instance := range instances {
			addr := instance.Addr
			eg.Go(func() error {
				insCC, er := grpc.Dial(addr, c.dialOpts...)
				if er != nil {
					ch <- Resp{Err: er}
					return er
				}

				nReply := reflect.New(typ).Interface()
				er = invoker(ctx, method, req, nReply, insCC, opts...)

				select {
				case <-ctx.Done():
					er = fmt.Errorf("响应没有人接收, %w", ctx.Err())
				case ch <- Resp{Reply: nReply, Err: er}:
				}

				return er
			})
		}
		err = eg.Wait()
		return err
	}
}

type broadcastKey struct {
}

func UseBroadcast(p context.Context) (context.Context, chan Resp) {
	c := make(chan Resp)
	return context.WithValue(p, broadcastKey{}, c), c
}

func IsBroadcast(ctx context.Context) (bool, chan Resp) {
	val, ok := ctx.Value(broadcastKey{}).(chan Resp)
	return ok, val
}

type Resp struct {
	Reply any
	Err   error
}
