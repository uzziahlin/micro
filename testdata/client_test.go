package testdata

import (
	"context"
	"fmt"
	"github.com/uzziahlin/micro/testdata/user"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func Test_Client(t *testing.T) {

	cc, err := grpc.Dial("localhost:8081", grpc.WithInsecure())

	require.NoError(t, err)

	client := user.NewServiceClient(cc)

	ctx := context.Background()

	resp, err := client.GetUserInfo(ctx, &user.GetInfoReq{
		Id: "1102",
	})

	require.NoError(t, err)

	fmt.Printf("resp is %v", resp)

}
