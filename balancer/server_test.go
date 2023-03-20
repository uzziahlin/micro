package micro

import (
	"context"
	"fmt"
	"github.com/uzziahlin/micro"
	"github.com/uzziahlin/micro/registry/etcd"
	"github.com/uzziahlin/micro/testdata/user"
	"golang.org/x/sync/errgroup"
	"testing"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/stretchr/testify/require"
)

func TestServer_Start(t *testing.T) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"159.75.251.57:2379"},
	})
	require.NoError(t, err)

	registry, err := etcd.NewEtcdRegistry(client)
	require.NoError(t, err)

	eg := errgroup.Group{}
	for i := 0; i < 5; i++ {
		ii := i
		eg.Go(func() error {
			server, err := micro.NewServer(
				"user-service",
				micro.ServerWithRegistry(registry),
				micro.ServerWithWeight(ii))
			require.NoError(t, err)

			user.RegisterServiceServer(server, &UserServiceImpl{
				idx: ii + 1,
			})

			return server.Start(fmt.Sprintf(":808%d", ii+1))
		})
	}
	err = eg.Wait()
	require.NoError(t, err)
}

type UserInfo struct {
	id   string
	name string
	age  int32
}

var userInfos = map[string]UserInfo{
	"1101": {
		id:   "1101",
		name: "jack",
		age:  18,
	},
	"1102": {
		id:   "1102",
		name: "tom",
		age:  15,
	},
	"1103": {
		id:   "1103",
		name: "rose",
		age:  19,
	},
}

type UserServiceImpl struct {
	user.UnimplementedServiceServer
	idx int
}

func (u UserServiceImpl) GetUserInfo(ctx context.Context, req *user.GetInfoReq) (*user.GetInfoResp, error) {
	id := req.Id
	info, ok := userInfos[id]

	if !ok {
		return nil, fmt.Errorf("ID为%s的学生不存在", id)
	}

	fmt.Printf("请求收到了，我是服务器：%d \n", u.idx)

	return &user.GetInfoResp{
		Id:   info.id,
		Name: info.name,
		Age:  info.age,
	}, nil
}
