package micro

import (
	"context"
	"github.com/uzziahlin/micro/registry"
	"google.golang.org/grpc"
	"net"
	"time"
)

type ServerOption func(server *Server)

// Server 对RPC服务端的抽象， 对grpc.Server进行封装
// 启动的时候会自身的实例信息注册给注册中心(如果指定了注册中心的话)
// 如果没指定注册中心，则会直接监听端口
type Server struct {
	*grpc.Server
	serviceName     string
	addr            string
	registry        registry.Registry
	si              *registry.ServiceInstance
	registerTimeout time.Duration
	listener        net.Listener
	weight          int
	group           string
}

func NewServer(serviceName string, opts ...ServerOption) (*Server, error) {

	server := grpc.NewServer()

	res := &Server{
		Server:          server,
		serviceName:     serviceName,
		registerTimeout: 3 * time.Second,
	}

	for _, opt := range opts {
		opt(res)
	}

	return res, nil
}

func ServerWithRegistry(registry registry.Registry) ServerOption {
	return func(server *Server) {
		server.registry = registry
	}
}

func ServerWithWeight(w int) ServerOption {
	return func(server *Server) {
		server.weight = w
	}
}

func ServerWithGroup(g string) ServerOption {
	return func(server *Server) {
		server.group = g
	}
}

// Start 启动服务，如果指定了注册中心，则会将自身的实例信息注册到注册中心
// 否则则会直接启动服务
func (s *Server) Start(addr string) error {
	// 监听端口
	listener, err := net.Listen("tcp", addr)

	s.addr = addr
	s.listener = listener

	if err != nil {
		return err
	}

	if s.registry != nil {
		err := s.registerSelf()
		if err != nil {
			return err
		}
	}

	return s.Serve(listener)
}

func (s *Server) registerSelf() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.registerTimeout)
	defer cancel()

	s.si = &registry.ServiceInstance{
		ServiceName: s.serviceName,
		Addr:        s.addr,
		Weight:      s.weight,
		Group:       s.group,
	}

	// 将自己注册到注册中心
	return s.registry.Register(ctx, s.si)
}

func (s *Server) Close() error {

	if s.registry != nil {
		ctx, cancel := context.WithTimeout(context.Background(), s.registerTimeout)
		defer cancel()
		err := s.registry.Unregister(ctx, s.si)
		if err != nil {
			return err
		}
	}

	if s.listener != nil {
		err := s.listener.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
