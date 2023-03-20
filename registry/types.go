package registry

import (
	"context"
	"io"
)

type Registry interface {
	Registrable
	Discoverable
	Subscribable
	io.Closer
}

type Registrable interface {
	Register(ctx context.Context, instance *ServiceInstance) error
	Unregister(ctx context.Context, instance *ServiceInstance) error
}

type Discoverable interface {
	Discover(ctx context.Context, serviceName string) ([]*ServiceInstance, error)
}

type Subscribable interface {
	Subscribe(serviceName string) <-chan Event
}

type ServiceInstance struct {
	ServiceName string
	Addr        string
	Weight      int
	Group       string
}

type Event struct {
}
