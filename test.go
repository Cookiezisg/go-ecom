package service

import (
	"context"
	"time"

	"ecommerce-system/internal/pkg/cache"
	apperrors "ecommerce-system/internal/pkg/errors"
	"ecommerce-system/internal/service/user/model"
	"ecommerce-system/internal/service/user/repository"
)

// AddressLogic 地址业务逻辑
type AddressLogic struct {
	addressRepo repository.AddressRepository
	cache       *cache.CacheOperations
}

// NewAddressLogic 创建地址业务逻辑
func NewAddressLogic(addressRepo repository.AddressRepository, cache *cache.CacheOperations) *AddressLogic {
	return &AddressLogic{
		addressRepo: addressRepo,
		cache:       cache,
	}
}

// GetAddressListRequest 获取地址列表请求
type GetAddressListRequest struct {
	UserID uint64
}

// GetAddressListResponse 获取地址列表响应
type GetAddressListResponse struct {
	Addresses []*model.Address
}

// GetAddressList 获取用户地址列表（带缓存）
func (l *AddressLogic) GetAddressList(ctx context.Context, req *GetAddressListRequest) (*GetAddressListResponse, error) {
	if req.UserID == 0 {
		return nil, apperrors.NewInvalidParamError("用户ID不能为空")
	}

	// 尝试从缓存获取
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixUserAddress, req.UserID)
		var addresses []*model.Address
		if err := l.cache.GetJSON(ctx, cacheKey, &addresses); err == nil {
			return &GetAddressListResponse{Addresses: addresses}, nil
		}
	}

	// 从数据库查询
	addresses, err := l.addressRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return nil, apperrors.NewInternalError("查询地址列表失败: " + err.Error())
	}

	// 写入缓存
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixUserAddress, req.UserID)
		_ = l.cache.Set(ctx, cacheKey, addresses, 1*time.Hour)
	}

	return &GetAddressListResponse{
		Addresses: addresses,
	}, nil
}

// AddAddressRequest 添加地址请求
type AddAddressRequest struct {
	UserID        uint64
	ReceiverName  string
	ReceiverPhone string
	Province      string
	City          string
	District      string
	Detail        string
	PostalCode    string
	IsDefault     int8
}

// AddAddressResponse 添加地址响应
type AddAddressResponse struct {
	Address *model.Address
}

// AddAddress 添加地址
func (l *AddressLogic) AddAddress(ctx context.Context, req *AddAddressRequest) (*AddAddressResponse, error) {
	// 参数验证
	if req.UserID == 0 {
		return nil, apperrors.NewInvalidParamError("用户ID不能为空")
	}
	if req.ReceiverName == "" {
		return nil, apperrors.NewInvalidParamError("收货人姓名不能为空")
	}
	if req.ReceiverPhone == "" {
		return nil, apperrors.NewInvalidParamError("收货人电话不能为空")
	}
	if req.Province == "" || req.City == "" || req.District == "" {
		return nil, apperrors.NewInvalidParamError("地址信息不完整")
	}
	if req.Detail == "" {
		return nil, apperrors.NewInvalidParamError("详细地址不能为空")
	}

	// 如果设置为默认地址，先取消其他默认地址
	if req.IsDefault == 1 {
		if err := l.addressRepo.SetDefault(ctx, req.UserID, 0); err != nil {
			// 如果设置失败，继续创建地址，但不设置为默认
			req.IsDefault = 0
		}
	}

	address := &model.Address{
		UserID:        req.UserID,
		ReceiverName:  req.ReceiverName,
		ReceiverPhone: req.ReceiverPhone,
		Province:      req.Province,
		City:          req.City,
		District:      req.District,
		Detail:        req.Detail,
		PostalCode:    req.PostalCode,
		IsDefault:     req.IsDefault,
	}

	if err := l.addressRepo.Create(ctx, address); err != nil {
		return nil, apperrors.NewInternalError("创建地址失败: " + err.Error())
	}

	// 删除缓存
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixUserAddress, req.UserID)
		_ = l.cache.Delete(ctx, cacheKey)
	}

	return &AddAddressResponse{
		Address: address,
	}, nil
}

// UpdateAddressRequest 更新地址请求
type UpdateAddressRequest struct {
	ID            uint64
	UserID        uint64
	ReceiverName  string
	ReceiverPhone string
	Province      string
	City          string
	District      string
	Detail        string
	PostalCode    string
	IsDefault     int8
}

// UpdateAddressResponse 更新地址响应
type UpdateAddressResponse struct {
	Address *model.Address
}

// UpdateAddress 更新地址
func (l *AddressLogic) UpdateAddress(ctx context.Context, req *UpdateAddressRequest) (*UpdateAddressResponse, error) {
	// 参数验证
	if req.ID == 0 {
		return nil, apperrors.NewInvalidParamError("地址ID不能为空")
	}
	if req.UserID == 0 {
		return nil, apperrors.NewInvalidParamError("用户ID不能为空")
	}

	// 获取原地址
	address, err := l.addressRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, apperrors.NewInternalError("查询地址失败: " + err.Error())
	}
	if address == nil {
		return nil, apperrors.NewError(apperrors.CodeNotFound, "地址不存在")
	}

	// 验证地址属于该用户
	if address.UserID != req.UserID {
		return nil, apperrors.NewError(apperrors.CodeForbidden, "无权操作此地址")
	}

	// 更新地址信息
	if req.ReceiverName != "" {
		address.ReceiverName = req.ReceiverName
	}
	if req.ReceiverPhone != "" {
		address.ReceiverPhone = req.ReceiverPhone
	}
	if req.Province != "" {
		address.Province = req.Province
	}
	if req.City != "" {
		address.City = req.City
	}
	if req.District != "" {
		address.District = req.District
	}
	if req.Detail != "" {
		address.Detail = req.Detail
	}
	if req.PostalCode != "" {
		address.PostalCode = req.PostalCode
	}

	// 如果设置为默认地址
	if req.IsDefault == 1 && address.IsDefault != 1 {
		if err := l.addressRepo.SetDefault(ctx, req.UserID, req.ID); err != nil {
			return nil, apperrors.NewInternalError("设置默认地址失败: " + err.Error())
		}
		address.IsDefault = 1
	}

	if err := l.addressRepo.Update(ctx, address); err != nil {
		return nil, apperrors.NewInternalError("更新地址失败: " + err.Error())
	}

	// 删除缓存
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixUserAddress, req.UserID)
		_ = l.cache.Delete(ctx, cacheKey)
	}

	return &UpdateAddressResponse{
		Address: address,
	}, nil
}

// DeleteAddressRequest 删除地址请求
type DeleteAddressRequest struct {
	ID     uint64
	UserID uint64
}

// DeleteAddressResponse 删除地址响应
type DeleteAddressResponse struct {
}

// DeleteAddress 删除地址
func (l *AddressLogic) DeleteAddress(ctx context.Context, req *DeleteAddressRequest) (*DeleteAddressResponse, error) {
	// 参数验证
	if req.ID == 0 {
		return nil, apperrors.NewInvalidParamError("地址ID不能为空")
	}
	if req.UserID == 0 {
		return nil, apperrors.NewInvalidParamError("用户ID不能为空")
	}

	// 获取原地址
	address, err := l.addressRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, apperrors.NewInternalError("查询地址失败: " + err.Error())
	}
	if address == nil {
		return nil, apperrors.NewError(apperrors.CodeNotFound, "地址不存在")
	}

	// 验证地址属于该用户
	if address.UserID != req.UserID {
		return nil, apperrors.NewError(apperrors.CodeForbidden, "无权操作此地址")
	}

	// 删除地址
	if err := l.addressRepo.Delete(ctx, req.ID); err != nil {
		return nil, apperrors.NewInternalError("删除地址失败: " + err.Error())
	}

	// 删除缓存
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixUserAddress, req.UserID)
		_ = l.cache.Delete(ctx, cacheKey)
	}

	return &DeleteAddressResponse{}, nil
}
