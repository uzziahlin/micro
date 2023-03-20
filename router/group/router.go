package group

import (
	"github.com/uzziahlin/micro/router"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/resolver"
)

type FilterBuilder struct {
}

func (g FilterBuilder) Build() router.Filter {
	return func(info balancer.PickInfo, addr resolver.Address) bool {
		group := info.Ctx.Value("group").(string)
		aGroup := addr.Attributes.Value("group").(string)
		return group == aGroup
	}
}
