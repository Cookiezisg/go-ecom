package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ecommerce-system/internal/service/user/model"
)

// AddressRepository 地址仓储接口
type AddressRepository interface {
	Create(ctx context.Context, address *model.Address) error
	GetByID(ctx context.Context, id uint64) (*model.Address, error)
	GetByUserID(ctx context.Context, userID uint64) ([]*model.Address, error)
	GetDefaultByUserID(ctx context.Context, userID uint64) (*model.Address, error)
	Update(ctx context.Context, address *model.Address) error
	Delete(ctx context.Context, id uint64) error
	SetDefault(ctx context.Context, userID uint64, addressID uint64) error
}

// addressRepository 地址仓储实现
type addressRepository struct {
	db *gorm.DB
}

// NewAddressRepository 创建地址仓储
func NewAddressRepository(db *gorm.DB) AddressRepository {
	return &addressRepository{
		db: db,
	}
}

// Create 创建地址
func (r *addressRepository) Create(ctx context.Context, address *model.Address) error {
	return r.db.WithContext(ctx).Create(address).Error
}

// GetByID 根据ID获取地址
func (r *addressRepository) GetByID(ctx context.Context, id uint64) (*model.Address, error) {
	var address model.Address
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&address).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &address, nil
}

// GetByUserID 根据用户ID获取地址列表
func (r *addressRepository) GetByUserID(ctx context.Context, userID uint64) ([]*model.Address, error) {
	var addresses []*model.Address
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("is_default DESC, created_at DESC").Find(&addresses).Error
	return addresses, err
}

// GetDefaultByUserID 获取用户默认地址
func (r *addressRepository) GetDefaultByUserID(ctx context.Context, userID uint64) (*model.Address, error) {
	var address model.Address
	err := r.db.WithContext(ctx).Where("user_id = ? AND is_default = ?", userID, 1).First(&address).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &address, nil
}

// Update 更新地址
func (r *addressRepository) Update(ctx context.Context, address *model.Address) error {
	return r.db.WithContext(ctx).Save(address).Error
}

// Delete 删除地址（软删除）
func (r *addressRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Address{}, id).Error
}

// SetDefault 设置默认地址
func (r *addressRepository) SetDefault(ctx context.Context, userID uint64, addressID uint64) error {
	// 开启事务
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先取消该用户所有地址的默认状态
		if err := tx.Model(&model.Address{}).Where("user_id = ?", userID).Update("is_default", 0).Error; err != nil {
			return err
		}
		// 设置指定地址为默认
		return tx.Model(&model.Address{}).Where("id = ? AND user_id = ?", addressID, userID).Update("is_default", 1).Error
	})
}
