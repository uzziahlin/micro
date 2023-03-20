package micro

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/uzziahlin/micro/registry/etcd"
	"github.com/uzziahlin/micro/testdata/user"
	clientv3 "go.etcd.io/etcd/client/v3"
	"testing"
	"time"
)

func TestClient_Dial(t *testing.T) {
	c, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"159.75.251.57:2379"},
	})
	require.NoError(t, err)

	registry, err := etcd.NewEtcdRegistry(c)
	require.NoError(t, err)

	client, err := NewClient(ClientWithRegistry(registry), ClientWithInsecure())
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cc, err := client.Dial(ctx, "user-service")
	require.NoError(t, err)

	serviceClient := user.NewServiceClient(cc)

	resp, err := serviceClient.GetUserInfo(ctx, &user.GetInfoReq{
		Id: "1101",
	})

	require.NoError(t, err)

	fmt.Printf("resp is %v", resp)
}
