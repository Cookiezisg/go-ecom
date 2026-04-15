package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"ecommerce-system/internal/service/seckill/model"
)

type ListSeckillActivitiesRequest struct {
	Page     int
	PageSize int
	// Status: 0-未开始，1-进行中，2-已结束；负数表示不过滤
	Status int32
	Keyword string
	Now    int64
	// IncludeDisabled: 是否包含禁用活动（a.status=0）
	IncludeDisabled bool
}

// SeckillActivityRow 查询结果（带 SKU 快照字段）
type SeckillActivityRow struct {
	model.SeckillActivity
	SkuName      string     `gorm:"column:sku_name"`
	SkuImage     string     `gorm:"column:sku_image"`
	SkuPrice     float64    `gorm:"column:sku_price"`
	SkuStatus    int8       `gorm:"column:sku_status"`
	SkuDeletedAt *time.Time `gorm:"column:sku_deleted_at"`
}

type SeckillActivityRepository interface {
	GetByID(ctx context.Context, id uint64) (*SeckillActivityRow, error)
	GetActiveBySkuID(ctx context.Context, skuID uint64, now int64) (*model.SeckillActivity, error)
	List(ctx context.Context, req *ListSeckillActivitiesRequest) ([]*SeckillActivityRow, int64, error)
	Create(ctx context.Context, act *model.SeckillActivity) error
	Update(ctx context.Context, id uint64, updates map[string]any) error
	Delete(ctx context.Context, id uint64) error
}

type seckillActivityRepository struct {
	db *gorm.DB
}

func NewSeckillActivityRepository(db *gorm.DB) SeckillActivityRepository {
	return &seckillActivityRepository{db: db}
}

func (r *seckillActivityRepository) baseQuery(ctx context.Context) *gorm.DB {
	// 显式指定表 + 软删除条件，和其他 repo 一致
	return r.db.WithContext(ctx).
		Table("seckill_activity AS a").
		Where("a.deleted_at IS NULL").
		Joins("LEFT JOIN sku s ON s.id = a.sku_id AND s.deleted_at IS NULL").
		Select(`a.*,
			COALESCE(s.name, '') AS sku_name,
			COALESCE(s.image, '') AS sku_image,
			COALESCE(s.price, 0) AS sku_price,
			COALESCE(s.status, 0) AS sku_status,
			s.deleted_at AS sku_deleted_at`)
}

func (r *seckillActivityRepository) GetByID(ctx context.Context, id uint64) (*SeckillActivityRow, error) {
	var row SeckillActivityRow
	err := r.baseQuery(ctx).Where("a.id = ?", id).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *seckillActivityRepository) GetActiveBySkuID(ctx context.Context, skuID uint64, now int64) (*model.SeckillActivity, error) {
	// 活动启用 + 时间窗口内
	var act model.SeckillActivity
	if now <= 0 {
		now = time.Now().Unix()
	}
	err := r.db.WithContext(ctx).
		Table("seckill_activity").
		Where("deleted_at IS NULL").
		Where("sku_id = ?", skuID).
		Where("status = 1").
		Where("start_time <= ? AND end_time >= ?", now, now).
		Order("id DESC").
		First(&act).Error
	if err != nil {
		return nil, err
	}
	return &act, nil
}

func (r *seckillActivityRepository) List(ctx context.Context, req *ListSeckillActivitiesRequest) ([]*SeckillActivityRow, int64, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	now := req.Now
	if now <= 0 {
		now = time.Now().Unix()
	}

	query := r.baseQuery(ctx)
	if !req.IncludeDisabled {
		query = query.Where("a.status = 1")
	}
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		query = query.Where("a.name LIKE ? OR s.name LIKE ?", like, like)
	}

	// 按“对外状态”过滤（由时间决定）
	switch req.Status {
	case 0:
		query = query.Where("a.start_time > ?", now)
	case 1:
		query = query.Where("a.start_time <= ? AND a.end_time >= ?", now, now)
	case 2:
		query = query.Where("a.end_time < ?", now)
	default:
		// 不过滤
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []*SeckillActivityRow
	if err := query.
		Order("a.start_time DESC, a.id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

func (r *seckillActivityRepository) Create(ctx context.Context, act *model.SeckillActivity) error {
	return r.db.WithContext(ctx).Create(act).Error
}

func (r *seckillActivityRepository) Update(ctx context.Context, id uint64, updates map[string]any) error {
	if id == 0 {
		return gorm.ErrRecordNotFound
	}
	if updates == nil {
		updates = map[string]any{}
	}
	return r.db.WithContext(ctx).
		Table("seckill_activity").
		Where("deleted_at IS NULL").
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *seckillActivityRepository) Delete(ctx context.Context, id uint64) error {
	if id == 0 {
		return gorm.ErrRecordNotFound
	}
	return r.db.WithContext(ctx).Delete(&model.SeckillActivity{}, id).Error
}
