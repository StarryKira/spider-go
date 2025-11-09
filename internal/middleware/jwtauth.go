package middleware

import (
	"spider-go/internal/dto"
	"spider-go/internal/handler/service"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// 用于认证JWT的中间件
func AuthMiddleWare(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			dto.Unauthorized(c, 401, "invalid JWT token")
			c.Abort()
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.ParseWithClaims(tokenString, &service.Claims{}, func(token *jwt.Token) (any, error) {
			//返回密钥验证签名
			return secret, nil
		})
		//检查是否出错
		if err != nil {
			dto.Unauthorized(c, 401, "invalid JWT token")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*service.Claims)

		if ok {
			c.Set("uid", claims.Uid)
			c.Set("username", claims.Name)
		}
		c.Next()
	}
}
