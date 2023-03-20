package etcd

import (
	"context"
	"fmt"
	"github.com/uzziahlin/micro/registry"
	"github.com/uzziahlin/micro/serialize"
	"github.com/uzziahlin/micro/serialize/json"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"sync"
)

type Registry struct {
	namespace   string
	client      *clientv3.Client
	sess        *concurrency.Session
	serializer  serialize.Serializer
	closeC      chan struct{}
	watchCancel []func()
	mu          sync.Mutex
}

func NewEtcdRegistry(client *clientv3.Client) (*Registry, error) {

	sess, err := concurrency.NewSession(client)

	if err != nil {
		return nil, err
	}

	res := &Registry{
		client:     client,
		sess:       sess,
		namespace:  "default",
		serializer: &json.Serializer{},
	}

	return res, nil
}

func (r *Registry) Register(ctx context.Context, instance *registry.ServiceInstance) error {
	bytes, err := r.serializer.Serialize(instance)

	if err != nil {
		return err
	}

	_, err = r.client.Put(ctx, r.instanceKey(instance), string(bytes), clientv3.WithLease(r.sess.Lease()))

	return err
}

func (r *Registry) Unregister(ctx context.Context, instance *registry.ServiceInstance) error {
	_, err := r.client.Delete(ctx, r.instanceKey(instance))

	return err
}

func (r *Registry) Discover(ctx context.Context, serviceName string) ([]*registry.ServiceInstance, error) {
	resp, err := r.client.Get(ctx, r.serviceKey(serviceName), clientv3.WithPrefix())

	if err != nil {
		return nil, err
	}

	res := make([]*registry.ServiceInstance, 0, len(resp.Kvs))

	for _, kv := range resp.Kvs {
		instance := new(registry.ServiceInstance)
		err := r.serializer.Deserialize(kv.Value, instance)
		if err != nil {
			return nil, err
		}
		res = append(res, instance)
	}

	return res, nil
}

func (r *Registry) Subscribe(serviceName string) <-chan registry.Event {
	ctx, cancel := context.WithCancel(context.Background())

	r.mu.Lock()
	r.watchCancel = append(r.watchCancel, cancel)
	r.mu.Unlock()

	ctx = clientv3.WithRequireLeader(ctx)
	watchC := r.client.Watch(ctx, r.serviceKey(serviceName), clientv3.WithPrefix())

	eventC := make(chan registry.Event)

	go func() {

		for {
			select {
			case resp := <-watchC:
				for range resp.Events {
					eventC <- registry.Event{}
				}
			case <-r.closeC:
				close(eventC)
				return
			}
		}

	}()

	return eventC
}

func (r *Registry) Close() error {
	r.mu.Lock()
	for _, cancel := range r.watchCancel {
		cancel()
	}
	r.mu.Unlock()
	close(r.closeC)
	return r.sess.Close()
}

// instanceKey 返回实例唯一标识键，规则为/namespace/serviceName/instanceName, 暂时以Addr作为instanceName
func (r *Registry) instanceKey(instance *registry.ServiceInstance) string {
	return fmt.Sprintf("/%s/%s/%s", r.namespace, instance.ServiceName, instance.Addr)
}

// serviceKey 返回服务唯一标识键，规则为/namespace/serviceName
func (r *Registry) serviceKey(serviceName string) string {
	return fmt.Sprintf("/%s/%s", r.namespace, serviceName)
}
