package main

import (
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	productpb "ecommerce-system/api/product/v1"
	"ecommerce-system/internal/service/product"
)

var configFile = flag.String("f", "configs/dev/product-config.yaml", "配置文件路径")

func main() {
	flag.Parse()

	// 加载配置（包含 RpcServerConf 和自定义配置）
	var c product.Config
	conf.MustLoad(*configFile, &c)

	// 创建服务上下文
	svcCtx := product.NewServiceContext(c)

	// 创建商品服务
	productSvc := product.NewProductService(svcCtx)

	// 创建 gRPC 服务器
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		// 注册服务
		productpb.RegisterProductServiceServer(grpcServer, productSvc)

		// 开发/测试环境开启 gRPC 反射（用于调试工具如 grpcurl 和 Gateway）
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("商品服务启动在 %s\\n", c.ListenOn)
	s.Start()
}
