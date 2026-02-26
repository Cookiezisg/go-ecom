package model

import (
	"time"

	"gorm.io/gorm"
)

type Cart struct {
	ID         uint64         `gorm:"primaryKey;column:id" json:"id"`
	UserID     uint64         `gorm:"column:user_id;not null;index" json:"user_id"`
	SkuID      uint64         `gorm:"column:sku_id;not null" json:"sku_id"`
	Quantity   int            `gorm:"column:quantity;default:1;not null" json:"quantity"`
	IsSelected int8           `gorm:"column:is_selected;default:1" json:"is_selected"`
	CreatedAt  time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (Cart) TableName() string {
	return "carts"
}
