package cart

import (
	"context"
	"fmt"
	"time"

	v1 "ecommerce-system/api/cart/v1"
	apperrors "ecommerce-system/internal/pkg/errors"
	"ecommerce-system/internal/pkg/utils"
	"ecommerce-system/internal/service/cart/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CartService 实现 gRPC 服务接口
type CartService struct {
	v1.UnimplementedCartServiceServer
	svcCtx *ServiceContext
	logic  *service.CartLogic
}

// NewCartService 创建购物车服务
func NewCartService(svcCtx *ServiceContext) *CartService {
	logic := service.NewCartLogic(svcCtx.CartRepo, svcCtx.ProductClient, svcCtx.InvClient)

	return &CartService{
		svcCtx: svcCtx,
		logic:  logic,
	}
}

// GetCart 获取购物车
func (s *CartService) GetCart(ctx context.Context, req *v1.GetCartRequest) (*v1.GetCartResponse, error) {
	userID, ok := utils.GetUserID(ctx)
	if !ok {
		if req.UserId > 0 {
			userID = uint64(req.UserId)
		} else {
			return nil, status.Error(codes.Unauthenticated, "未授权，请先登录")
		}
	}

	resp, err := s.logic.GetCart(ctx, &service.GetCartRequest{UserID: userID})
	if err != nil {
		return nil, convertError(err)
	}

	items := make([]*v1.CartItem, 0, len(resp.Items))
	for _, item := range resp.Items {
		items = append(items, convertDetailToProto(item))
	}

	return &v1.GetCartResponse{
		Code:    0,
		Message: "成功",
		Data:    items,
	}, nil
}

// AddItem 添加商品到购物车
func (s *CartService) AddItem(ctx context.Context, req *v1.AddItemRequest) (*v1.AddItemResponse, error) {
	userID, ok := utils.GetUserID(ctx)
	if !ok {
		if req.UserId > 0 {
			userID = uint64(req.UserId)
		} else {
			return nil, status.Error(codes.Unauthenticated, "未授权，请先登录")
		}
	}

	resp, err := s.logic.AddItem(ctx, &service.AddItemRequest{
		UserID:   userID,
		SkuID:    uint64(req.SkuId),
		Quantity: int(req.Quantity),
	})
	if err != nil {
		return nil, convertError(err)
	}

	item := &v1.CartItem{
		Id:          int64(resp.Cart.ID),
		UserId:      int64(resp.Cart.UserID),
		SkuId:       int64(resp.Cart.SkuID),
		Quantity:    int32(resp.Cart.Quantity),
		IsSelected:  int32(resp.Cart.IsSelected),
		CreatedAt:   formatTime(&resp.Cart.CreatedAt),
		UpdatedAt:   formatTime(&resp.Cart.UpdatedAt),
		ProductId:   resp.ProductID,
		ProductName: resp.ProductName,
		SkuName:     resp.SkuName,
		SkuImage:    resp.SkuImage,
		Price:       fmt.Sprintf("%.2f", resp.Price),
		StockStatus: resp.StockStatus,
	}

	return &v1.AddItemResponse{
		Code:    0,
		Message: "添加成功",
		Data:    item,
	}, nil
}

// UpdateQuantity 更新商品数量
func (s *CartService) UpdateQuantity(ctx context.Context, req *v1.UpdateQuantityRequest) (*v1.UpdateQuantityResponse, error) {
	userID, ok := utils.GetUserID(ctx)
	if !ok {
		if req.UserId > 0 {
			userID = uint64(req.UserId)
		} else {
			return nil, status.Error(codes.Unauthenticated, "未授权，请先登录")
		}
	}

	err := s.logic.UpdateQuantity(ctx, &service.UpdateQuantityRequest{
		UserID:   userID,
		SkuID:    uint64(req.SkuId),
		Quantity: int(req.Quantity),
	})
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.UpdateQuantityResponse{Code: 0, Message: "更新成功"}, nil
}

// RemoveItem 删除商品
func (s *CartService) RemoveItem(ctx context.Context, req *v1.RemoveItemRequest) (*v1.RemoveItemResponse, error) {
	userID, ok := utils.GetUserID(ctx)
	if !ok {
		if req.UserId > 0 {
			userID = uint64(req.UserId)
		} else {
			return nil, status.Error(codes.Unauthenticated, "未授权，请先登录")
		}
	}

	skuIDs := make([]uint64, 0, len(req.SkuIds))
	for _, id := range req.SkuIds {
		skuIDs = append(skuIDs, uint64(id))
	}

	err := s.logic.RemoveItem(ctx, &service.RemoveItemRequest{UserID: userID, SkuIDs: skuIDs})
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.RemoveItemResponse{Code: 0, Message: "删除成功"}, nil
}

// ClearCart 清空购物车
func (s *CartService) ClearCart(ctx context.Context, req *v1.ClearCartRequest) (*v1.ClearCartResponse, error) {
	err := s.logic.ClearCart(ctx, &service.ClearCartRequest{UserID: uint64(req.UserId)})
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.ClearCartResponse{Code: 0, Message: "清空成功"}, nil
}

// SelectItem 选择/取消选择商品
func (s *CartService) SelectItem(ctx context.Context, req *v1.SelectItemRequest) (*v1.SelectItemResponse, error) {
	err := s.logic.SelectItem(ctx, &service.SelectItemRequest{
		UserID:     uint64(req.UserId),
		SkuID:      uint64(req.SkuId),
		IsSelected: int8(req.IsSelected),
	})
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.SelectItemResponse{Code: 0, Message: "操作成功"}, nil
}

// BatchSelect 批量选择/取消选择
func (s *CartService) BatchSelect(ctx context.Context, req *v1.BatchSelectRequest) (*v1.BatchSelectResponse, error) {
	skuIDs := make([]uint64, 0, len(req.SkuIds))
	for _, id := range req.SkuIds {
		skuIDs = append(skuIDs, uint64(id))
	}

	err := s.logic.BatchSelect(ctx, &service.BatchSelectRequest{
		UserID:     uint64(req.UserId),
		SkuIDs:     skuIDs,
		IsSelected: int8(req.IsSelected),
	})
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.BatchSelectResponse{Code: 0, Message: "操作成功"}, nil
}

// convertError 转换业务错误为 gRPC 错误
func convertError(err error) error {
	if err == nil {
		return nil
	}

	if bizErr, ok := err.(*apperrors.BusinessError); ok {
		var grpcCode codes.Code
		switch bizErr.Code {
		case apperrors.CodeNotFound:
			grpcCode = codes.NotFound
		case apperrors.CodeInvalidParam:
			grpcCode = codes.InvalidArgument
		case apperrors.CodeUnauthorized:
			grpcCode = codes.Unauthenticated
		case apperrors.CodeForbidden:
			grpcCode = codes.PermissionDenied
		case apperrors.CodeStockInsufficient:
			grpcCode = codes.ResourceExhausted
		default:
			grpcCode = codes.Internal
		}
		return status.Error(grpcCode, bizErr.Error())
	}

	return status.Error(codes.Internal, err.Error())
}

// convertDetailToProto 将 CartItemDetail 转为 proto CartItem
func convertDetailToProto(item *service.CartItemDetail) *v1.CartItem {
	if item == nil || item.Cart == nil {
		return nil
	}
	return &v1.CartItem{
		Id:          int64(item.ID),
		UserId:      int64(item.UserID),
		SkuId:       int64(item.SkuID),
		Quantity:    int32(item.Quantity),
		IsSelected:  int32(item.IsSelected),
		CreatedAt:   formatTime(&item.CreatedAt),
		UpdatedAt:   formatTime(&item.UpdatedAt),
		ProductId:   item.ProductID,
		ProductName: item.ProductName,
		SkuName:     item.SkuName,
		SkuImage:    item.SkuImage,
		Price:       fmt.Sprintf("%.2f", item.Price),
		StockStatus: item.StockStatus,
	}
}

// formatTime 格式化时间为字符串
func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}
