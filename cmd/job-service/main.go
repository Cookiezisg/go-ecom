package main

import (
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	jobpb "ecommerce-system/api/job/v1"
	"ecommerce-system/internal/service/job"
)

var configFile = flag.String("f", "configs/dev/job-config.yaml", "配置文件路径")

func main() {
	flag.Parse()

	var c job.Config
	conf.MustLoad(*configFile, &c)

	svcCtx := job.NewServiceContext(c)
	jobSvc := job.NewJobService(svcCtx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		// 注册服务
		jobpb.RegisterJobServiceServer(grpcServer, jobSvc)

		// 开发/测试环境开启 gRPC 反射（用于调试工具如 grpcurl 和 Gateway）
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("定时任务服务启动在 %s\\n", c.ListenOn)
	s.Start()
}
