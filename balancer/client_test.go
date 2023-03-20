package micro

import (
	"context"
	"fmt"
	"github.com/uzziahlin/micro"
	"github.com/uzziahlin/micro/balancer/random"
	"github.com/uzziahlin/micro/registry/etcd"
	"github.com/uzziahlin/micro/testdata/user"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestClient_Dial(t *testing.T) {
	c, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"159.75.251.57:2379"},
	})
	require.NoError(t, err)

	registry, err := etcd.NewEtcdRegistry(c)
	require.NoError(t, err)

	client, err := micro.NewClient(micro.ClientWithRegistry(registry), micro.ClientWithInsecure(),
		micro.ClientWithBalancer("DEMO_ROUND_ROBIN", &random.BalancerBuilder{}))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cc, err := client.Dial(ctx, "user-service")
	require.NoError(t, err)

	serviceClient := user.NewServiceClient(cc)

	for i := 0; i < 200; i++ {
		resp, err := serviceClient.GetUserInfo(ctx, &user.GetInfoReq{
			Id: "1101",
		})
		require.NoError(t, err)

		fmt.Printf("resp is %v \n", resp)
	}

}

type User struct {
	name string
	age  int
}

func (u *User) talk() {
	if u == nil {
		fmt.Println("user is nil")
		return
	}
	fmt.Printf("user is not nil, his name is %s, age is %d", u.name, u.age)
}

func TestNil(t *testing.T) {
	var u = (*User)(nil)
	u.talk()
	u.talk()
}
