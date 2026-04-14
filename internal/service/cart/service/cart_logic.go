package service

import (
	"context"
	"fmt"

	"ecommerce-system/internal/pkg/client"
	apperrors "ecommerce-system/internal/pkg/errors"
	"ecommerce-system/internal/service/cart/model"
	"ecommerce-system/internal/service/cart/repository"
)

// CartLogic 购物车业务逻辑
type CartLogic struct {
	cartRepo      repository.CartRepository
	productClient *client.ProductClient
	invClient     *client.InventoryClient
}

// NewCartLogic 创建购物车业务逻辑
func NewCartLogic(
	cartRepo repository.CartRepository,
	productClient *client.ProductClient,
	invClient *client.InventoryClient,
) *CartLogic {
	return &CartLogic{
		cartRepo:      cartRepo,
		productClient: productClient,
		invClient:     invClient,
	}
}

// CartItemDetail 购物车商品（含商品富信息）
type CartItemDetail struct {
	*model.Cart
	ProductID   int64
	ProductName string
	SkuName     string
	SkuImage    string
	Price       float64
	StockStatus string // "in_stock" | "low_stock" | "out_of_stock"
}

// GetCartRequest 获取购物车请求
type GetCartRequest struct {
	UserID uint64
}

// GetCartResponse 获取购物车响应
type GetCartResponse struct {
	Items []*CartItemDetail
}

// GetCart 获取购物车（批量查 SKU 信息，返回带价格的商品列表）
func (l *CartLogic) GetCart(ctx context.Context, req *GetCartRequest) (*GetCartResponse, error) {
	carts, err := l.cartRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return nil, apperrors.NewInternalError("获取购物车失败")
	}

	details := make([]*CartItemDetail, 0, len(carts))
	for _, c := range carts {
		detail := &CartItemDetail{Cart: c}
		if l.productClient != nil {
			if sku, err := l.productClient.GetSku(ctx, int64(c.SkuID)); err == nil && sku != nil {
				detail.ProductID = sku.ProductId
				detail.SkuName = sku.Name
				detail.SkuImage = sku.Image
				detail.Price = sku.Price
			}
		}
		detail.StockStatus = l.resolveStockStatus(ctx, c.SkuID)
		details = append(details, detail)
	}

	return &GetCartResponse{Items: details}, nil
}

// AddItemRequest 添加商品请求
type AddItemRequest struct {
	UserID   uint64
	SkuID    uint64
	Quantity int
}

// AddItemResponse 添加商品响应（含 SKU 富信息）
type AddItemResponse struct {
	Cart        *model.Cart
	ProductID   int64
	ProductName string
	SkuName     string
	SkuImage    string
	Price       float64
	StockStatus string
}

// AddItem 添加商品到购物车（校验库存，获取 SKU 信息）
func (l *CartLogic) AddItem(ctx context.Context, req *AddItemRequest) (*AddItemResponse, error) {
	if req.Quantity <= 0 {
		return nil, apperrors.NewInvalidParamError("商品数量必须大于0")
	}

	resp := &AddItemResponse{StockStatus: "in_stock"}

	// 1. 获取 SKU 信息（同时校验 SKU 存在且在售）
	if l.productClient != nil {
		sku, err := l.productClient.GetSku(ctx, int64(req.SkuID))
		if err != nil {
			return nil, apperrors.NewError(apperrors.CodeSkuNotFound, "商品不存在或已下架")
		}
		if sku.Status != 1 {
			return nil, apperrors.NewError(apperrors.CodeSkuOffline, "商品已下架，无法加入购物车")
		}
		resp.ProductID = sku.ProductId
		resp.SkuName = sku.Name
		resp.SkuImage = sku.Image
		resp.Price = sku.Price

		// 获取商品名
		if sku.ProductId > 0 {
			if product, err := l.productClient.GetProduct(ctx, sku.ProductId); err == nil && product != nil {
				resp.ProductName = product.Name
			}
		}
	}

	// 2. 校验库存
	if l.invClient != nil {
		inventories, err := l.invClient.BatchGetInventory(ctx, []int64{int64(req.SkuID)})
		if err == nil && len(inventories) > 0 {
			inv := inventories[0]
			available := int(inv.AvailableStock)
			if available <= 0 {
				return nil, apperrors.NewError(apperrors.CodeStockInsufficient, "库存不足，无法加入购物车")
			}
			if available < req.Quantity {
				return nil, apperrors.NewError(apperrors.CodeStockInsufficient,
					fmt.Sprintf("库存不足，当前可用库存 %d 件", available))
			}
			switch {
			case available <= 10:
				resp.StockStatus = "low_stock"
			default:
				resp.StockStatus = "in_stock"
			}
		}
	}

	// 3. 加入购物车（upsert：已有则累加数量）
	cart := &model.Cart{
		UserID:     req.UserID,
		SkuID:      req.SkuID,
		Quantity:   req.Quantity,
		IsSelected: 1,
	}
	if err := l.cartRepo.AddItem(ctx, cart); err != nil {
		return nil, apperrors.NewInternalError("添加商品失败")
	}

	resp.Cart = cart
	return resp, nil
}

// UpdateQuantityRequest 更新数量请求
type UpdateQuantityRequest struct {
	UserID   uint64
	SkuID    uint64
	Quantity int
}

// UpdateQuantity 更新商品数量
func (l *CartLogic) UpdateQuantity(ctx context.Context, req *UpdateQuantityRequest) error {
	if req.Quantity <= 0 {
		return apperrors.NewInvalidParamError("数量必须大于0")
	}

	// 校验库存
	if l.invClient != nil {
		inventories, err := l.invClient.BatchGetInventory(ctx, []int64{int64(req.SkuID)})
		if err == nil && len(inventories) > 0 {
			available := int(inventories[0].AvailableStock)
			if available < req.Quantity {
				return apperrors.NewError(apperrors.CodeStockInsufficient,
					fmt.Sprintf("库存不足，当前可用库存 %d 件", available))
			}
		}
	}

	if err := l.cartRepo.UpdateQuantity(ctx, req.UserID, req.SkuID, req.Quantity); err != nil {
		return apperrors.NewInternalError("更新数量失败")
	}
	return nil
}

// RemoveItemRequest 删除商品请求
type RemoveItemRequest struct {
	UserID uint64
	SkuIDs []uint64
}

// RemoveItem 删除商品
func (l *CartLogic) RemoveItem(ctx context.Context, req *RemoveItemRequest) error {
	if err := l.cartRepo.RemoveItem(ctx, req.UserID, req.SkuIDs); err != nil {
		return apperrors.NewInternalError("删除商品失败")
	}
	return nil
}

// ClearCartRequest 清空购物车请求
type ClearCartRequest struct {
	UserID uint64
}

// ClearCart 清空购物车
func (l *CartLogic) ClearCart(ctx context.Context, req *ClearCartRequest) error {
	if err := l.cartRepo.ClearCart(ctx, req.UserID); err != nil {
		return apperrors.NewInternalError("清空购物车失败")
	}
	return nil
}

// SelectItemRequest 选择商品请求
type SelectItemRequest struct {
	UserID     uint64
	SkuID      uint64
	IsSelected int8
}

// SelectItem 选择/取消选择商品
func (l *CartLogic) SelectItem(ctx context.Context, req *SelectItemRequest) error {
	if err := l.cartRepo.SelectItem(ctx, req.UserID, req.SkuID, req.IsSelected); err != nil {
		return apperrors.NewInternalError("操作失败")
	}
	return nil
}

// BatchSelectRequest 批量选择请求
type BatchSelectRequest struct {
	UserID     uint64
	SkuIDs     []uint64
	IsSelected int8
}

// BatchSelect 批量选择/取消选择
func (l *CartLogic) BatchSelect(ctx context.Context, req *BatchSelectRequest) error {
	if err := l.cartRepo.BatchSelect(ctx, req.UserID, req.SkuIDs, req.IsSelected); err != nil {
		return apperrors.NewInternalError("批量操作失败")
	}
	return nil
}

// resolveStockStatus 根据库存查询结果返回库存状态
func (l *CartLogic) resolveStockStatus(ctx context.Context, skuID uint64) string {
	if l.invClient == nil {
		return "in_stock"
	}
	inventories, err := l.invClient.BatchGetInventory(ctx, []int64{int64(skuID)})
	if err != nil || len(inventories) == 0 {
		return "in_stock"
	}
	available := int(inventories[0].AvailableStock)
	switch {
	case available <= 0:
		return "out_of_stock"
	case available <= 10:
		return "low_stock"
	default:
		return "in_stock"
	}
}
