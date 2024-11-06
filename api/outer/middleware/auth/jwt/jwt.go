package jwt

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// JWT JSON Web Token
type JWT struct {
	SigningKey []byte
}

var (
	// ErrTokenExpired 令牌已过期
	ErrTokenExpired = errors.New("token is expired")
	// ErrTokenNotValidYet 令牌未激活
	ErrTokenNotValidYet = errors.New("token not active yet")
	// ErrTokenMalformed 不是令牌
	ErrTokenMalformed = errors.New("that's not even a token")
	// ErrTokenInvalid 无法处理此令牌
	ErrTokenInvalid = errors.New("couldn't handle this token")
	// SignKey token签名的钥匙
	SignKey = "125sdas1dda231d1wdasd"
)

// CustomClaims 自定义jwt声明
type CustomClaims struct {
	UserID     string `json:"user_id" bson:"user_id"`
	Email      string `json:"email" bson:"email"`
	Domain     string `json:"domain" bson:"domain"`
	CustomerID string `json:"customer_id" bson:"customer_id"`
	jwt.StandardClaims
}

// NewJWT 创建jwt
func NewJWT() *JWT {
	return &JWT{
		[]byte(getSignKey()),
	}
}

// getSignKey 获取签名字符串
func getSignKey() string {
	return SignKey
}

// CreateToken 创建token
func (j *JWT) CreateToken(claims CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

// ParseToken 验证并转换token
func (j *JWT) ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, ErrTokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				// Token is expired
				return nil, ErrTokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, ErrTokenNotValidYet
			} else {
				return nil, ErrTokenInvalid
			}
		}
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, ErrTokenInvalid
}

// RefreshToken token过期，重新生成token
func (j *JWT) RefreshToken(tokenString string) (string, error) {
	jwt.TimeFunc = func() time.Time {
		return time.Unix(0, 0)
	}
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		jwt.TimeFunc = time.Now
		claims.StandardClaims.ExpiresAt = time.Now().Add(8 * time.Hour).Unix()
		return j.CreateToken(*claims)
	}
	return "", ErrTokenInvalid
}
