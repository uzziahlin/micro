# 一个基于grpc开发的go语言微服务框架

首先需要使用protobuf生成客户端和服务端的代码，这里使用的是protoc-gen-go-grpc插件，使用方法如下：
```shell
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative user.proto
```

客户端使用示例如下(以etcd作为注册中心进行演示，用户也可以扩展自己的注册中心)：
```go
c, err := clientv3.New(clientv3.Config{
    Endpoints: []string{"your etcd address"},
})
if err != nil {
    // todo handle error
}

registry, err := etcd.NewEtcdRegistry(c)
if err != nil {
    // todo handle error
}

client, err := NewClient(ClientWithRegistry(registry), ClientWithInsecure())
if err != nil {
    // todo handle error
}

ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

cc, err := client.Dial(ctx, "user-service")
if err != nil {
    // todo handle error
}

serviceClient := user.NewServiceClient(cc)

resp, err := serviceClient.GetUserInfo(ctx, &user.GetInfoReq{
    Id: "1101",
})
if err != nil {
    // todo handle error
}
// todo handle response
```

服务端示例代码如下：
```go
client, err := clientv3.New(clientv3.Config{
    Endpoints: []string{"your etcd address"},
})
if err != nil {
    // todo handle error
}

registry, err := etcd.NewEtcdRegistry(client)
if err != nil {
    // todo handle error
}

server, err := NewServer("user-service", ServerWithRegistry(registry))
if err != nil {
    // todo handle error
}

user.RegisterServiceServer(server, &UserServiceImpl{})

err = server.Start("serve address")
if err != nil {
    // todo handle error
}
```