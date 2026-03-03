package cart

import (
	"context"
	v1 "ecommerce-system/api/cart/v1"
	apperrors "ecommerce-system/internal/pkg/errors"
	"ecommerce-system/internal/service/cart/model"
	"ecommerce-system/internal/service/cart/service"
	"time"

	"ecommerce-system/internal/pkg/utils"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CartService struct {
	v1.UnimplementedCartServiceServer
	svcCtx *ServiceContext
	logic  *service.CartLogic
}

func NewCartService(svcCtx *ServiceContext) *CartService {
	logic := service.NewCartLogic(svcCtx.CartRepo)

	return &CartService{
		svcCtx: svcCtx,
		logic:  logic,
	}
}

func (s *CartService) GetCart(ctx context.Context, req *v1.GetCartRequest) (*v1.GetCartResponse, error) {
	userID, ok := utils.GetUserID(ctx)
	if !ok {
		if req.UserId > 0 {
			userID = uint64(req.UserId)
		} else {
			return nil, status.Error(codes.Unauthenticated, "当前没有用户登陆")
		}
	}

	getReq := &service.GetCartRequest{
		UserID: userID,
	}
	resp, err := s.logic.GetCart(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	items := make([]*v1.CartItem, 0, len(resp.Items))
	for _, item := range resp.Items {
		items = append(items, convertCartItemToProto(item))
	}

	return &v1.GetCartResponse{
		Code:    0,
		Message: "成功",
		Data:    items,
	}, nil
}

func (s *CartService) AddItem(ctx context.Context, req *v1.AddItemRequest) (*v1.AddItemResponse, error) {
	userID, ok := utils.GetUserID(ctx)
	if !ok {
		if req.UserId > 0 {
			userID = uint64(req.UserId)
		} else {
			return nil, status.Error(codes.Unauthenticated, "当前没有用户登陆")
		}
	}

	addReq := &service.AddItemRequest{
		UserID:   userID,
		SkuID:    uint64(req.SkuId),
		Quantity: int(req.Quantity),
	}

	resp, err := s.logic.AddItem(ctx, addReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.AddItemResponse{
		Code:    0,
		Message: "成功",
		Data:    convertCartItemToProto(resp.Cart),
	}, nil
}

func convertCartItemToProto(item *model.Cart) *v1.CartItem {
	if item == nil {
		return nil
	}

	createdAt := ""
	updatedAt := ""
	if !item.CreatedAt.IsZero() {
		createdAt = item.CreatedAt.Format(time.RFC3339)
	}
	if !item.UpdatedAt.IsZero() {
		updatedAt = item.UpdatedAt.Format(time.RFC3339)
	}

	return &v1.CartItem{
		Id:         int64(item.ID),
		UserId:     int64(item.UserID),
		SkuId:      int64(item.SkuID),
		Quantity:   int32(item.Quantity),
		IsSelected: int32(item.IsSelected),
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}
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
		default:
			grpcCode = codes.Internal
		}
		return status.Error(grpcCode, bizErr.Error())
	}

	return status.Error(codes.Internal, err.Error())
}
