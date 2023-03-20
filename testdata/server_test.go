package testdata

import (
	"context"
	"fmt"
	"github.com/uzziahlin/micro/testdata/user"
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	"google.golang.org/grpc"
)

func Test_Server(t *testing.T) {

	listener, err := net.Listen("tcp", ":8081")

	require.NoError(t, err)

	userService := &UserServiceImpl{}

	server := grpc.NewServer()

	user.RegisterServiceServer(server, userService)

	err = server.Serve(listener)

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
}

func (u UserServiceImpl) GetUserInfo(ctx context.Context, req *user.GetInfoReq) (*user.GetInfoResp, error) {
	id := req.Id
	info, ok := userInfos[id]

	if !ok {
		return nil, fmt.Errorf("ID为%s的学生不存在", id)
	}

	return &user.GetInfoResp{
		Id:   info.id,
		Name: info.name,
		Age:  info.age,
	}, nil
}
