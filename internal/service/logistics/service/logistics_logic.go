package service

import (
	"context"
	"time"

	apperrors "ecommerce-system/internal/pkg/errors"
	"ecommerce-system/internal/pkg/idgen"
	"ecommerce-system/internal/service/logistics/model"
	"ecommerce-system/internal/service/logistics/repository"
)

// LogisticsLogic 物流业务逻辑
type LogisticsLogic struct {
	logisticsRepo repository.LogisticsRepository
	idGen         *idgen.Generator
}

// NewLogisticsLogic 创建物流业务逻辑
func NewLogisticsLogic(logisticsRepo repository.LogisticsRepository, idGen *idgen.Generator) *LogisticsLogic {
	return &LogisticsLogic{
		logisticsRepo: logisticsRepo,
		idGen:         idGen,
	}
}

// CreateLogisticsRequest 创建物流单请求
type CreateLogisticsRequest struct {
	OrderID         uint64
	OrderNo         string
	CompanyCode     string
	CompanyName     string
	ReceiverName    string
	ReceiverPhone   string
	ReceiverAddress string
}

// CreateLogisticsResponse 创建物流单响应
type CreateLogisticsResponse struct {
	Logistics *model.Logistics
}

// CreateLogistics 创建物流单（物流单号由 idgen 生成，格式 LGS+yyyyMMdd+8位序号）
func (l *LogisticsLogic) CreateLogistics(ctx context.Context, req *CreateLogisticsRequest) (*CreateLogisticsResponse, error) {
	var logisticsNo string
	if l.idGen != nil {
		logisticsNo = l.idGen.LogisticsNo(ctx)
	} else {
		logisticsNo = "LGS" + time.Now().Format("20060102150405")
	}

	logistics := &model.Logistics{
		OrderID:          req.OrderID,
		OrderNo:          req.OrderNo,
		LogisticsCompany: req.CompanyName,
		LogisticsNo:      logisticsNo,
		ReceiverName:     req.ReceiverName,
		ReceiverPhone:    req.ReceiverPhone,
		ReceiverAddress:  req.ReceiverAddress,
		Status:           0, // 待发货
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	err := l.logisticsRepo.Create(ctx, logistics)
	if err != nil {
		return nil, apperrors.NewInternalError("创建物流单失败")
	}

	return &CreateLogisticsResponse{Logistics: logistics}, nil
}

// GetLogisticsRequest 获取物流信息请求
type GetLogisticsRequest struct {
	OrderID uint64
}

// GetLogisticsResponse 获取物流信息响应
type GetLogisticsResponse struct {
	Logistics *model.Logistics
}

// GetLogistics 获取物流信息
func (l *LogisticsLogic) GetLogistics(ctx context.Context, req *GetLogisticsRequest) (*GetLogisticsResponse, error) {
	logistics, err := l.logisticsRepo.GetByOrderID(ctx, req.OrderID)
	if err != nil {
		return nil, apperrors.NewNotFoundError("物流信息不存在")
	}

	return &GetLogisticsResponse{Logistics: logistics}, nil
}

// UpdateLogisticsStatusRequest 更新物流状态请求
type UpdateLogisticsStatusRequest struct {
	LogisticsNo string
	Status      int8
	Remark      string
}

// UpdateLogisticsStatus 更新物流状态
func (l *LogisticsLogic) UpdateLogisticsStatus(ctx context.Context, req *UpdateLogisticsStatusRequest) error {
	logistics, err := l.logisticsRepo.GetByLogisticsNo(ctx, req.LogisticsNo)
	if err != nil {
		return apperrors.NewNotFoundError("物流信息不存在")
	}

	logistics.Status = req.Status
	now := time.Now()
	switch req.Status {
	case 1: // 已发货
		logistics.ShippedAt = &now
	case 3: // 已签收
		logistics.DeliveredAt = &now
	}

	err = l.logisticsRepo.Update(ctx, logistics)
	if err != nil {
		return apperrors.NewInternalError("更新物流状态失败")
	}

	return nil
}

// QueryTrackingRequest 查询物流轨迹请求
type QueryTrackingRequest struct {
	LogisticsNo string
}

// TrackingNode 物流轨迹节点
type TrackingNode struct {
	Time     string
	Status   string
	Location string
	Remark   string
}

// QueryTrackingResponse 查询物流轨迹响应
type QueryTrackingResponse struct {
	Nodes []*TrackingNode
}

// QueryTracking 查询物流轨迹（从本地状态构建基础轨迹；接入三方 API 时替换此处逻辑）
func (l *LogisticsLogic) QueryTracking(ctx context.Context, req *QueryTrackingRequest) (*QueryTrackingResponse, error) {
	logistics, err := l.logisticsRepo.GetByLogisticsNo(ctx, req.LogisticsNo)
	if err != nil {
		return nil, apperrors.NewNotFoundError("物流信息不存在")
	}

	nodes := []*TrackingNode{
		{
			Time:     logistics.CreatedAt.Format(time.RFC3339),
			Status:   "已创建",
			Location: "",
			Remark:   "物流单已创建，单号：" + logistics.LogisticsNo,
		},
	}

	if logistics.ShippedAt != nil {
		nodes = append(nodes, &TrackingNode{
			Time:     logistics.ShippedAt.Format(time.RFC3339),
			Status:   "已发货",
			Location: "",
			Remark:   "商品已由" + logistics.LogisticsCompany + "揽收",
		})
	}

	if logistics.DeliveredAt != nil {
		nodes = append(nodes, &TrackingNode{
			Time:     logistics.DeliveredAt.Format(time.RFC3339),
			Status:   "已签收",
			Location: "",
			Remark:   "商品已签收",
		})
	}

	return &QueryTrackingResponse{Nodes: nodes}, nil
}

// CalculateFreightRequest 计算运费请求
type CalculateFreightRequest struct {
	Province string
	City     string
	District string
	Weight   float64
	Volume   float64
}

// CalculateFreightResponse 计算运费响应
type CalculateFreightResponse struct {
	Freight float64
}

// CalculateFreight 计算运费（简化规则：基础费 10 元，重量每千克 2 元，体积每立方厘米 1.5 元）
func (l *LogisticsLogic) CalculateFreight(ctx context.Context, req *CalculateFreightRequest) (*CalculateFreightResponse, error) {
	baseFreight := 10.0
	weightFreight := req.Weight * 2.0
	volumeFreight := req.Volume * 1.5

	freight := baseFreight + weightFreight + volumeFreight
	if freight < 10 {
		freight = 10
	}

	return &CalculateFreightResponse{Freight: freight}, nil
}
