package repository

import (
	"context"
	"ecommerce-system/internal/service/promotion/model"

	"gorm.io/gorm"
)

// CouponRepository 优惠券仓库接口
type CouponRepository interface {
	// GetList 获取优惠券列表
	GetList(ctx context.Context, page, pageSize int) ([]*model.Coupon, int64, error)
	// GetByID 根据ID获取
	GetByID(ctx context.Context, id uint64) (*model.Coupon, error)
	// Create 创建优惠券
	Create(ctx context.Context, coupon *model.Coupon) error
	// Update 更新优惠券
	Update(ctx context.Context, coupon *model.Coupon) error
	// IncrUsedCount 原子加 1 used_count（仅当 total_count=0 或 used_count < total_count 时生效）
	// 返回受影响行数；0 表示已售完，调用方应返回「已领完」错误。
	IncrUsedCount(ctx context.Context, id uint64) (int64, error)
}

type couponRepository struct {
	db *gorm.DB
}

// NewCouponRepository 创建优惠券仓库
func NewCouponRepository(db *gorm.DB) CouponRepository {
	return &couponRepository{db: db}
}

// GetList 获取优惠券列表
func (r *couponRepository) GetList(ctx context.Context, page, pageSize int) ([]*model.Coupon, int64, error) {
	var coupons []*model.Coupon
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Coupon{}).Where("status = ?", 1)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&coupons).Error
	return coupons, total, err
}

// GetByID 根据ID获取
func (r *couponRepository) GetByID(ctx context.Context, id uint64) (*model.Coupon, error) {
	var coupon model.Coupon
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&coupon).Error
	if err != nil {
		return nil, err
	}
	return &coupon, nil
}

// Create 创建优惠券
func (r *couponRepository) Create(ctx context.Context, coupon *model.Coupon) error {
	return r.db.WithContext(ctx).Create(coupon).Error
}

// Update 更新优惠券
func (r *couponRepository) Update(ctx context.Context, coupon *model.Coupon) error {
	return r.db.WithContext(ctx).Save(coupon).Error
}

// IncrUsedCount 原子递增 used_count（WHERE 确保不超发）
func (r *couponRepository) IncrUsedCount(ctx context.Context, id uint64) (int64, error) {
	result := r.db.WithContext(ctx).Exec(
		`UPDATE coupons SET used_count = used_count + 1 WHERE id = ? AND (total_count = 0 OR used_count < total_count)`,
		id,
	)
	return result.RowsAffected, result.Error
}
