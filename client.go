package micro

import (
	"context"
	"fmt"
	"github.com/uzziahlin/micro/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"time"
)

type ClientOption func(client *Client)

func ClientWithRegistry(registry registry.Registry) ClientOption {
	return func(client *Client) {
		client.registry = registry
	}
}

func ClientWithInsecure() ClientOption {
	return func(client *Client) {
		client.insecure = true
	}
}

func ClientWithTimeout(timeout time.Duration) ClientOption {
	return func(client *Client) {
		client.resolverTimeout = timeout
	}
}

func ClientWithBalancer(name string, builder base.PickerBuilder) ClientOption {
	return func(client *Client) {
		client.balancer = &PickerBuilder{
			PickerBuilder: builder,
			name:          name,
		}
	}
}

func NewClient(opts ...ClientOption) (*Client, error) {
	res := &Client{
		resolverTimeout: 60 * time.Second,
	}

	for _, opt := range opts {
		opt(res)
	}

	return res, nil
}

type Client struct {
	registry        registry.Registry
	insecure        bool
	resolverTimeout time.Duration
	balancer        *PickerBuilder
}

func (c *Client) Dial(ctx context.Context, service string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {

	dialOpts := make([]grpc.DialOption, 0)

	address := service

	if c.registry != nil {
		dialOpts = append(dialOpts, grpc.WithResolvers(RegisterBuilder{
			registry: c.registry,
			timeout:  c.resolverTimeout,
		}))
		address = fmt.Sprintf("%s:///%s", scheme, service)
	}

	if pb := c.balancer; pb != nil {
		balancer.Register(base.NewBalancerBuilder(pb.name, pb.PickerBuilder, base.Config{HealthCheck: true}))
		dialOpts = append(dialOpts, grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy":"%s"}`, pb.name)))
	}

	if c.insecure {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	}

	if len(opts) > 0 {
		dialOpts = append(dialOpts, opts...)
	}

	return grpc.DialContext(ctx, address, dialOpts...)
}

type PickerBuilder struct {
	base.PickerBuilder
	name string
}
