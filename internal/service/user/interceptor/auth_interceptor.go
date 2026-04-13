package interceptor

import (
	"context"
	"strings"

	"ecommerce-system/internal/pkg/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// AuthInterceptor 认证拦截器，从 metadata 中提取 JWT token 并解析 user_id
func AuthInterceptor(jwtSecret string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return handler(ctx, req)
		}

		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return handler(ctx, req)
		}

		token := authHeaders[0]
		token = strings.TrimPrefix(token, "Bearer ")

		claims, err := utils.ParseToken(token, jwtSecret)
		if err != nil {
			// 解析失败不强制拦截：由具体 Handler 决定是否要求登录
			return handler(ctx, req)
		}

		ctx = utils.WithUserID(ctx, claims.UserID)
		ctx = utils.WithUsername(ctx, claims.Username)
		return handler(ctx, req)
	}
}
