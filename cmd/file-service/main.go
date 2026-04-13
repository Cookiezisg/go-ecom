package main

import (
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	filepb "ecommerce-system/api/file/v1"
	"ecommerce-system/internal/service/file"
)

var configFile = flag.String("f", "configs/dev/file-config.yaml", "配置文件路径")

func main() {
	flag.Parse()

	var c file.Config
	conf.MustLoad(*configFile, &c)

	svcCtx := file.NewServiceContext(c)
	fileSvc := file.NewFileService(svcCtx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		// 注册服务
		filepb.RegisterFileServiceServer(grpcServer, fileSvc)

		// 开发/测试环境开启 gRPC 反射（用于调试工具如 grpcurl 和 Gateway）
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("文件服务启动在 %s\\n", c.ListenOn)
	s.Start()
}
