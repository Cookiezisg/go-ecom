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
		// 从 metadata 中获取 Authorization header
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			// 如果没有 metadata，尝试继续处理（某些接口可能不需要认证）
			return handler(ctx, req)
		}

		// 获取 Authorization header
		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			// 如果没有 Authorization header，尝试继续处理
			return handler(ctx, req)
		}

		// 解析 Bearer token
		token := authHeaders[0]
		token = strings.TrimPrefix(token, "Bearer ")

		// 解析 JWT token
		claims, err := utils.ParseToken(token, jwtSecret)
		if err != nil {
			// Token 解析失败，但不阻止请求（某些接口可能不需要认证）
			// 如果需要强制认证，可以返回错误：
			// return nil, status.Error(codes.Unauthenticated, "token无效")
			return handler(ctx, req)
		}

		// 将 user_id 和 username 添加到 context
		ctx = utils.WithUserID(ctx, claims.UserID)
		ctx = utils.WithUsername(ctx, claims.Username)

		return handler(ctx, req)
	}
}
