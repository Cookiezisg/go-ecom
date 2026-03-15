package repository

import (
	"context"
	"ecommerce-system/internal/service/product/model"

	"gorm.io/gorm"
)

type ListSkusRequest struct {
	ProductID uint64
	// ProductIDs: 当需要"点击主分类展示其所有子分类商品"时使用（product_id IN (...)）
	// 若该字段非空，则优先使用该字段过滤，并忽略 ProductID。
	Status   int8 // -1-全部, 0-否, 1-是
	Page     int
	PageSize int
}

// SkuRepository SKU数据访问接口
type SkuRepository interface {
	Create(ctx context.Context, sku *model.Sku) error
	GetByID(ctx context.Context, id uint64) (*model.Sku, error)
	GetBySkuCode(ctx context.Context, skuCode string) (*model.Sku, error)
	GetBySkuCodeUnscoped(ctx context.Context, skuCode string) (*model.Sku, error)
	GetByProductID(ctx context.Context, productID uint64) ([]*model.Sku, error)
	List(ctx context.Context, req *ListSkusRequest) ([]*model.Sku, int64, error)
	GetAggByProductIDs(ctx context.Context, productIDs []uint64, status int8) (map[uint64]SkuAgg, error)
	RestoreAndUpdateByID(ctx context.Context, id uint64, updates map[string]any) error
	Update(ctx context.Context, sku *model.Sku) error
	Delete(ctx context.Context, id uint64) error
}

// SkuAgg SKU 聚合信息（用于列表页展示最低价/总库存）
type SkuAgg struct {
	MinPrice   float64
	TotalStock int64
	// MinOriginalPrice 如果需要在商品列表展示"原价"，可用它兜底；为空表示所有 SKU 都没有原价
	MinOriginalPrice *float64
}

type skuRepository struct {
	db *gorm.DB
}

func NewSkuRepository(db *gorm.DB) SkuRepository {
	return &skuRepository{db: db}
}

func (r *skuRepository) Create(ctx context.Context, sku *model.Sku) error {
	return r.db.WithContext(ctx).Create(sku).Error
}

func (r *skuRepository) GetByID(ctx context.Context, id uint64) (*model.Sku, error) {
	var sku model.Sku
	err := r.db.WithContext(ctx).First(&sku, id).Error
	if err != nil {
		return nil, err
	}
	return &sku, nil
}

func (r *skuRepository) GetBySkuCode(ctx context.Context, skuCode string) (*model.Sku, error) {
	var sku model.Sku
	err := r.db.WithContext(ctx).Where("sku_code = ?", skuCode).First(&sku).Error
	if err != nil {
		return nil, err
	}
	return &sku, nil
}

func (r *skuRepository) GetBySkuCodeUnscoped(ctx context.Context, skuCode string) (*model.Sku, error) {
	var sku model.Sku
	err := r.db.WithContext(ctx).Unscoped().Where("sku_code = ?", skuCode).First(&sku).Error
	if err != nil {
		return nil, err
	}
	return &sku, nil
}

func (r *skuRepository) GetByProductID(ctx context.Context, productID uint64) ([]*model.Sku, error)

func (r *skuRepository) List(ctx context.Context, req *ListSkusRequest) ([]*model.Sku, int64, error)

func (r *skuRepository) GetAggByProductIDs(ctx context.Context, productIDs []uint64, status int8) (map[uint64]SkuAgg, error)

func (r *skuRepository) RestoreAndUpdateByID(ctx context.Context, id uint64, updates map[string]any) error

func (r *skuRepository) Update(ctx context.Context, sku *model.Sku) error

func (r *skuRepository) Delete(ctx context.Context, id uint64) error
