package cart

import (
	v1 "ecommerce-system/api/cart/v1"
	"ecommerce-system/internal/service/cart/service"
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
