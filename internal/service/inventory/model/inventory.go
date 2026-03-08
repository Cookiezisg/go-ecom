package model

import (
	"time"
)

// Inventory 库存模型
type Inventory struct {
	ID                uint64    `gorm:"primaryKey;column:id" json:"id"`
	SkuID             uint64    `gorm:"column:sku_id;uniqueIndex;not null" json:"sku_id"`
	TotalStock        int       `gorm:"column:total_stock;default:0;not null" json:"total_stock"`
	AvailableStock    int       `gorm:"column:available_stock;default:0;not null" json:"available_stock"`
	LockedStock       int       `gorm:"column:locked_stock;default:0;not null" json:"locked_stock"`
	SoldStock         int       `gorm:"column:sold_stock;default:0;not null" json:"sold_stock"`
	LowStockThreshold int       `gorm:"column:low_stock_threshold;default:10" json:"low_stock_threshold"`
	CreatedAt         time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (Inventory) TableName() string {
	return "inventory"
}

// InventoryLog 库存流水模型
type InventoryLog struct {
	ID          uint64    `gorm:"primaryKey;column:id" json:"id"`
	SkuID       uint64    `gorm:"column:sku_id;not null;index" json:"sku_id"`
	OrderID     *uint64   `gorm:"column:order_id;index" json:"order_id"`
	Type        int8      `gorm:"column:type;not null;index" json:"type"` // 1-入库, 2-出库, 3-锁定, 4-解锁, 5-扣减, 6-回退
	Quantity    int       `gorm:"column:quantity;not null" json:"quantity"`
	BeforeStock int       `gorm:"column:before_stock;not null" json:"before_stock"`
	AfterStock  int       `gorm:"column:after_stock;not null" json:"after_stock"`
	Remark      string    `gorm:"column:remark;size:255" json:"remark"`
	CreatedAt   time.Time `gorm:"column:created_at;index" json:"created_at"`
}

// TableName 指定表名
func (InventoryLog) TableName() string {
	return "inventory_log"
}
