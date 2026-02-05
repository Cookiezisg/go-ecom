package user

import (
	"context"
	"time"

	v1 "ecommerce-system/api/user/v1"
)

type Service struct {
	v1.UnimplementedUserServiceServer
}

func NewUserService() *Service {
	return &Service{}
}

func (s *Service) Register(ctx context.Context, req *v1.RegisterRequest) (*v1.RegisterResponse, error) {
	now := time.Now().Format(time.RFC3339)
	return &v1.RegisterResponse{
		Code:    0,
		Message: "注册成功(演示版)",
		Data: &v1.User{
			Id:        1,
			Username:  req.Username,
			Phone:     req.Phone,
			Email:     req.Email,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}, nil
}

func (s *Service) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	now := time.Now().Unix()
	return &v1.LoginResponse{
		Code:    0,
		Message: "登录成功(演示版)",
		Data: &v1.LoginData{
			User: &v1.User{
				Id:       1,
				Username: req.Username,
			},
			Token:      "demo-token",
			ExpireTime: now + 7200,
		},
	}, nil
}

func (s *Service) GetUserInfo(ctx context.Context, req *v1.GetUserInfoRequest) (*v1.GetUserInfoResponse, error) {
	now := time.Now().Format(time.RFC3339)
	return &v1.GetUserInfoResponse{
		Code:    0,
		Message: "成功(演示版)",
		Data: &v1.User{
			Id:        req.UserId,
			Username:  "demo",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}, nil
}

func (s *Service) UpdateUserInfo(ctx context.Context, req *v1.UpdateUserInfoRequest) (*v1.UpdateUserInfoResponse, error) {
	now := time.Now().Format(time.RFC3339)
	return &v1.UpdateUserInfoResponse{
		Code:    0,
		Message: "更新成功(演示版)",
		Data: &v1.User{
			Id:        req.UserId,
			Nickname:  req.Nickname,
			Avatar:    req.Avatar,
			Gender:    req.Gender,
			Birthday:  req.Birthday,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}, nil
}

func (s *Service) ListUsers(ctx context.Context, req *v1.ListUsersRequest) (*v1.ListUsersResponse, error) {
	return &v1.ListUsersResponse{
		Code:    0,
		Message: "成功(演示版)",
		Data: &v1.ListUsersData{
			Users:    []*v1.User{},
			Total:    0,
			Page:     req.Page,
			PageSize: req.PageSize,
		},
	}, nil
}

func (s *Service) DeleteUser(ctx context.Context, req *v1.DeleteUserRequest) (*v1.DeleteUserResponse, error) {
	return &v1.DeleteUserResponse{Code: 0, Message: "删除成功(演示版)"}, nil
}

func (s *Service) GetAddressList(ctx context.Context, req *v1.GetAddressListRequest) (*v1.GetAddressListResponse, error) {
	return &v1.GetAddressListResponse{Code: 0, Message: "成功(演示版)", Data: []*v1.Address{}}, nil
}

func (s *Service) AddAddress(ctx context.Context, req *v1.AddAddressRequest) (*v1.AddAddressResponse, error) {
	now := time.Now().Format(time.RFC3339)
	return &v1.AddAddressResponse{
		Code:    0,
		Message: "添加成功(演示版)",
		Data: &v1.Address{
			Id:            1,
			UserId:        req.UserId,
			ReceiverName:  req.ReceiverName,
			ReceiverPhone: req.ReceiverPhone,
			Province:      req.Province,
			City:          req.City,
			District:      req.District,
			Detail:        req.Detail,
			PostalCode:    req.PostalCode,
			IsDefault:     req.IsDefault,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
	}, nil
}

func (s *Service) UpdateAddress(ctx context.Context, req *v1.UpdateAddressRequest) (*v1.UpdateAddressResponse, error) {
	now := time.Now().Format(time.RFC3339)
	return &v1.UpdateAddressResponse{
		Code:    0,
		Message: "更新成功(演示版)",
		Data: &v1.Address{
			Id:            req.Id,
			UserId:        req.UserId,
			ReceiverName:  req.ReceiverName,
			ReceiverPhone: req.ReceiverPhone,
			Province:      req.Province,
			City:          req.City,
			District:      req.District,
			Detail:        req.Detail,
			PostalCode:    req.PostalCode,
			IsDefault:     req.IsDefault,
			UpdatedAt:     now,
		},
	}, nil
}

func (s *Service) DeleteAddress(ctx context.Context, req *v1.DeleteAddressRequest) (*v1.DeleteAddressResponse, error) {
	return &v1.DeleteAddressResponse{Code: 0, Message: "删除成功(演示版)"}, nil
}
