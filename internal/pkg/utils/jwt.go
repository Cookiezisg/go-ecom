package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired = errors.New("token已过期")
	ErrTokenInvalid = errors.New("token无效")
)

// JWTClaims JWT声明
type JWTClaims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT Token
func GenerateToken(userID uint64, username, secret string, expire int64) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expire) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken 解析JWT Token
func ParseToken(tokenString, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil //此处是一个回调函数，接受token来返回密钥，如果token的签名方法不符合预期，则返回错误，这里主要是一是防止攻击，而且这样做可以支持多种签名方法
	})

	if err != nil {
		return nil, err
	} //第一层检查，如果token解析失败，直接返回错误

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	} //类型断言，如果断言成功且token有效，则返回claims，否则返回错误
	// 还验证一下过期时间，虽然jwt库会自动验证，但我们可以额外检查一下，以防止时钟偏移等问题

	return nil, ErrTokenInvalid
	// 直接滚了
}
