package middleware

import (
	"context"
	"strings"

	"ecommerce-system/internal/pkg/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor gRPC 一元拦截器：从 metadata 解析 JWT，将 user_id/username 注入 context。
// 解析失败时不强制拦截——由具体 handler 决定是否要求登录。
// 需要强制鉴权的接口，请使用 RequireAuthInterceptor。
func AuthInterceptor(jwtSecret string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx = injectUserFromMeta(ctx, jwtSecret)
		return handler(ctx, req)
	}
}

// RequireAuthInterceptor gRPC 一元拦截器：强制要求 JWT 有效，否则返回 Unauthenticated。
// 白名单（skipMethods）中的方法名不做校验，格式如 "/user.v1.UserService/Login"。
func RequireAuthInterceptor(jwtSecret string, skipMethods ...string) grpc.UnaryServerInterceptor {
	skip := make(map[string]struct{}, len(skipMethods))
	for _, m := range skipMethods {
		skip[m] = struct{}{}
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if _, ok := skip[info.FullMethod]; ok {
			return handler(ctx, req)
		}

		ctx = injectUserFromMeta(ctx, jwtSecret)

		if _, ok := utils.GetUserID(ctx); !ok {
			return nil, status.Error(codes.Unauthenticated, "未授权，请先登录")
		}
		return handler(ctx, req)
	}
}

// injectUserFromMeta 从 gRPC metadata 解析 Authorization header，写入 context
func injectUserFromMeta(ctx context.Context, jwtSecret string) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}
	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return ctx
	}
	token := strings.TrimPrefix(authHeaders[0], "Bearer ")
	claims, err := utils.ParseToken(token, jwtSecret)
	if err != nil {
		return ctx
	}
	ctx = utils.WithUserID(ctx, claims.UserID)
	ctx = utils.WithUsername(ctx, claims.Username)
	return ctx
}
