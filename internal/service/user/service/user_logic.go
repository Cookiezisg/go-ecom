package service

import (
	"context"
	"time"

	"ecommerce-system/internal/pkg/cache"
	"ecommerce-system/internal/pkg/constants"
	apperrors "ecommerce-system/internal/pkg/errors"
	"ecommerce-system/internal/pkg/utils"
	"ecommerce-system/internal/service/user/model"
	"ecommerce-system/internal/service/user/repository"
)

// UserLogic 用户业务逻辑
type UserLogic struct {
	userRepo       repository.UserRepository
	credentialRepo repository.CredentialRepository
	addressRepo    repository.AddressRepository
	cache          *cache.CacheOperations
}

// NewUserLogic 创建用户业务逻辑
func NewUserLogic(
	userRepo repository.UserRepository,
	credentialRepo repository.CredentialRepository,
	addressRepo repository.AddressRepository,
	cache *cache.CacheOperations,
) *UserLogic {
	return &UserLogic{
		userRepo:       userRepo,
		credentialRepo: credentialRepo,
		addressRepo:    addressRepo,
		cache:          cache,
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username   string
	Password   string
	Phone      string
	Email      string
	VerifyCode string
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	UserID   uint64
	Username string
	User     *model.User
}

// Register 用户注册
func (l *UserLogic) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	// 1. 参数验证
	if req.Username == "" {
		return nil, apperrors.NewInvalidParamError("用户名不能为空")
	}
	if req.Password == "" || len(req.Password) < 6 {
		return nil, apperrors.NewInvalidParamError("密码长度至少6位")
	}

	// 2. 检查用户名是否已存在
	existingUser, err := l.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, apperrors.NewInternalError("查询用户失败: " + err.Error())
	}
	if existingUser != nil {
		return nil, apperrors.NewError(apperrors.CodeUserAlreadyExists, "用户名已存在")
	}

	// 3. 检查手机号是否已存在（如果提供）
	if req.Phone != "" {
		existingUser, err = l.userRepo.GetByPhone(ctx, req.Phone)
		if err != nil {
			return nil, apperrors.NewInternalError("查询用户失败: " + err.Error())
		}
		if existingUser != nil {
			return nil, apperrors.NewError(apperrors.CodeUserAlreadyExists, "手机号已被注册")
		}
	}

	// 4. 检查邮箱是否已存在（如果提供）
	if req.Email != "" {
		existingUser, err = l.userRepo.GetByEmail(ctx, req.Email)
		if err != nil {
			return nil, apperrors.NewInternalError("查询用户失败: " + err.Error())
		}
		if existingUser != nil {
			return nil, apperrors.NewError(apperrors.CodeUserAlreadyExists, "邮箱已被注册")
		}
	}

	// 5. 密码加密
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, apperrors.NewInternalError("密码加密失败: " + err.Error())
	}

	// 6. 创建用户
	now := time.Now()
	user := &model.User{
		Username:    req.Username,
		Status:      constants.UserStatusNormal,
		MemberLevel: constants.MemberLevelNormal,
		Points:      0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// 只有当 phone 和 email 不为空时才设置，避免空字符串导致唯一索引冲突
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Email != "" {
		user.Email = req.Email
	}

	if err := l.userRepo.CreateWithOmit(ctx, user); err != nil {
		return nil, apperrors.NewInternalError("创建用户失败: " + err.Error())
	}

	// 7. 创建密码凭证
	credential := &model.Credential{
		UserID:          user.ID,
		CredentialType:  1, // 1-密码
		CredentialKey:   req.Username,
		CredentialValue: hashedPassword,
		Extra:           "{}", // JSON 字段不能为空字符串，设置为空对象
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := l.credentialRepo.Create(ctx, credential); err != nil {
		// 如果凭证创建失败，删除已创建的用户（可选，根据业务需求）
		_ = l.userRepo.Delete(ctx, user.ID)
		return nil, apperrors.NewInternalError("创建凭证失败: " + err.Error())
	}

	return &RegisterResponse{
		UserID:   user.ID,
		Username: user.Username,
		User:     user,
	}, nil
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username   string // 用户名/手机号/邮箱
	Password   string
	LoginType  int    // 1-用户名, 2-手机号, 3-邮箱
	VerifyCode string // 验证码
}

// LoginResponse 登录响应
type LoginResponse struct {
	UserID     uint64
	Username   string
	User       *model.User
	Token      string
	ExpireTime int64
}

// Login 用户登录
func (l *UserLogic) Login(ctx context.Context, req *LoginRequest, jwtSecret string, jwtExpire int64) (*LoginResponse, error) {
	// 1. 参数验证
	if req.Username == "" {
		return nil, apperrors.NewInvalidParamError("用户名不能为空")
	}
	if req.Password == "" {
		return nil, apperrors.NewInvalidParamError("密码不能为空")
	}
	// 验证码验证（前端已做基础验证，这里可以进一步验证或记录）
	// 注意：验证码可以为空（如果前端没有发送），但如果有值则应该验证
	if req.VerifyCode != "" {
		// 这里可以添加后端验证码验证逻辑（如 Redis 缓存验证）
		// 目前前端已经做了基础验证，这里暂时只做非空检查
	}

	// 2. 根据登录类型查找用户
	var user *model.User
	var err error

	switch req.LoginType {
	case 1: // 用户名登录
		user, err = l.userRepo.GetByUsername(ctx, req.Username)
	case 2: // 手机号登录
		user, err = l.userRepo.GetByPhone(ctx, req.Username)
	case 3: // 邮箱登录
		user, err = l.userRepo.GetByEmail(ctx, req.Username)
	default:
		return nil, apperrors.NewInvalidParamError("登录类型错误")
	}

	if err != nil {
		return nil, apperrors.NewInternalError("查询用户失败: " + err.Error())
	}
	if user == nil {
		return nil, apperrors.NewError(apperrors.CodeUserNotFound, "用户不存在")
	}

	// 3. 检查用户状态
	if user.Status != constants.UserStatusNormal {
		return nil, apperrors.NewError(apperrors.CodeForbidden, "用户已被禁用")
	}

	// 4. 验证密码
	credential, err := l.credentialRepo.GetByUserIDAndType(ctx, user.ID, 1) // 1-密码
	if err != nil {
		return nil, apperrors.NewInternalError("查询凭证失败: " + err.Error())
	}
	if credential == nil {
		return nil, apperrors.NewError(apperrors.CodePasswordError, "密码凭证不存在")
	}

	if !utils.CheckPassword(req.Password, credential.CredentialValue) {
		return nil, apperrors.NewError(apperrors.CodePasswordError, "密码错误")
	}

	// 5. 生成JWT Token
	token, err := utils.GenerateToken(user.ID, user.Username, jwtSecret, jwtExpire)
	if err != nil {
		return nil, apperrors.NewInternalError("生成Token失败: " + err.Error())
	}

	expireTime := time.Now().Add(time.Duration(jwtExpire) * time.Second).Unix()

	// 缓存用户会话信息
	if l.cache != nil {
		sessionKey := cache.BuildKey(cache.KeyPrefixUserSession, token)
		sessionData := map[string]interface{}{
			"user_id":   user.ID,
			"username":  user.Username,
			"expire_at": expireTime,
		}
		_ = l.cache.Set(ctx, sessionKey, sessionData, time.Duration(jwtExpire)*time.Second)

		// 缓存用户信息
		userKey := cache.BuildKey(cache.KeyPrefixUserInfo, user.ID)
		_ = l.cache.Set(ctx, userKey, user, 30*time.Minute)
	}

	return &LoginResponse{
		UserID:     user.ID,
		Username:   user.Username,
		User:       user,
		Token:      token,
		ExpireTime: expireTime,
	}, nil
}

// GetUserInfoRequest 获取用户信息请求
type GetUserInfoRequest struct {
	UserID uint64
}

// GetUserInfoResponse 获取用户信息响应
type GetUserInfoResponse struct {
	User *model.User
}

// GetUserInfo 获取用户信息（带缓存）
func (l *UserLogic) GetUserInfo(ctx context.Context, req *GetUserInfoRequest) (*GetUserInfoResponse, error) {
	if req.UserID == 0 {
		return nil, apperrors.NewInvalidParamError("用户ID不能为空")
	}

	// 尝试从缓存获取
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixUserInfo, req.UserID)
		var user model.User
		if err := l.cache.GetJSON(ctx, cacheKey, &user); err == nil {
			return &GetUserInfoResponse{User: &user}, nil
		}
	}

	// 从数据库查询
	user, err := l.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, apperrors.NewInternalError("查询用户失败: " + err.Error())
	}
	if user == nil {
		return nil, apperrors.NewError(apperrors.CodeUserNotFound, "用户不存在")
	}

	// 写入缓存
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixUserInfo, req.UserID)
		_ = l.cache.Set(ctx, cacheKey, user, 30*time.Minute)
	}

	return &GetUserInfoResponse{
		User: user,
	}, nil
}

// UpdateUserInfoRequest 更新用户信息请求
type UpdateUserInfoRequest struct {
	UserID   uint64
	Nickname string
	Avatar   string
	Gender   int
	Birthday string
}

// UpdateUserInfoResponse 更新用户信息响应
type UpdateUserInfoResponse struct {
	User *model.User
}

// UpdateUserInfo 更新用户信息
func (l *UserLogic) UpdateUserInfo(ctx context.Context, req *UpdateUserInfoRequest) (*UpdateUserInfoResponse, error) {
	// 1. 参数验证
	if req.UserID == 0 {
		return nil, apperrors.NewInvalidParamError("用户ID不能为空")
	}

	// 2. 获取用户
	user, err := l.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, apperrors.NewInternalError("查询用户失败: " + err.Error())
	}
	if user == nil {
		return nil, apperrors.NewError(apperrors.CodeUserNotFound, "用户不存在")
	}

	// 3. 更新用户信息
	updated := false
	if req.Nickname != "" {
		user.Nickname = req.Nickname
		updated = true
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
		updated = true
	}
	if req.Gender >= 0 {
		user.Gender = int8(req.Gender)
		updated = true
	}
	if req.Birthday != "" {
		birthday, err := time.Parse("2006-01-02", req.Birthday)
		if err != nil {
			return nil, apperrors.NewInvalidParamError("生日格式错误，应为 YYYY-MM-DD")
		}
		user.Birthday = &birthday
		updated = true
	}

	if !updated {
		return &UpdateUserInfoResponse{User: user}, nil
	}

	// 4. 保存更新
	user.UpdatedAt = time.Now()
	if err := l.userRepo.Update(ctx, user); err != nil {
		return nil, apperrors.NewInternalError("更新用户信息失败: " + err.Error())
	}

	// 删除缓存
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixUserInfo, req.UserID)
		_ = l.cache.Delete(ctx, cacheKey)
	}

	return &UpdateUserInfoResponse{
		User: user,
	}, nil
}

// ListUsersRequest 用户列表请求
type ListUsersRequest struct {
	Page     int
	PageSize int
	Keyword  string
	Status   *int8 // nil 表示不筛选
}

// ListUsersResponse 用户列表响应
type ListUsersResponse struct {
	Users    []*model.User
	Total    int64
	Page     int
	PageSize int
}

// ListUsers 获取用户列表（管理后台）
func (l *UserLogic) ListUsers(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error) {
	if l.userRepo == nil {
		return nil, apperrors.NewInternalError("数据库连接未初始化")
	}

	// 参数验证
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100 // 限制最大每页数量
	}

	users, total, err := l.userRepo.List(ctx, req.Page, req.PageSize, req.Keyword, req.Status)
	if err != nil {
		return nil, apperrors.NewInternalError("查询用户列表失败: " + err.Error())
	}

	return &ListUsersResponse{
		Users:    users,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// DeleteUserRequest 删除用户请求
type DeleteUserRequest struct {
	UserID uint64
}

// DeleteUserResponse 删除用户响应
type DeleteUserResponse struct {
}

// DeleteUser 删除用户（管理后台）
func (l *UserLogic) DeleteUser(ctx context.Context, req *DeleteUserRequest) (*DeleteUserResponse, error) {
	if l.userRepo == nil {
		return nil, apperrors.NewInternalError("数据库连接未初始化")
	}

	if req.UserID == 0 {
		return nil, apperrors.NewInvalidParamError("用户ID不能为空")
	}

	// 检查用户是否存在
	user, err := l.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, apperrors.NewInternalError("查询用户失败: " + err.Error())
	}
	if user == nil {
		return nil, apperrors.NewError(apperrors.CodeUserNotFound, "用户不存在")
	}

	// 删除用户（软删除）
	if err := l.userRepo.Delete(ctx, req.UserID); err != nil {
		return nil, apperrors.NewInternalError("删除用户失败: " + err.Error())
	}

	return &DeleteUserResponse{}, nil
}
