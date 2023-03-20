package router

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/resolver"
)

// Filter 对过滤规则的抽象
type Filter func(info balancer.PickInfo, addr resolver.Address) bool
