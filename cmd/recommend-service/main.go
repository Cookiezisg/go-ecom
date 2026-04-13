package main

import (
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	recommendpb "ecommerce-system/api/recommend/v1"
	"ecommerce-system/internal/service/recommend"
)

var configFile = flag.String("f", "configs/dev/recommend-config.yaml", "配置文件路径")

func main() {
	flag.Parse()

	var c recommend.Config
	conf.MustLoad(*configFile, &c)

	svcCtx := recommend.NewServiceContext(c)
	recommendSvc := recommend.NewRecommendService(svcCtx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		// 注册服务
		recommendpb.RegisterRecommendServiceServer(grpcServer, recommendSvc)

		// 开发/测试环境开启 gRPC 反射（用于调试工具如 grpcurl 和 Gateway）
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("推荐服务启动在 %s\\n", c.ListenOn)
	s.Start()
}
