package middleware

import (
	"spider-go/internal/common"
	"spider-go/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthAdminMiddleWare(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			common.Error(c, common.CodeUnauthorized, "请提供有效的令牌")
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.ParseWithClaims(tokenString, &service.Claims{}, func(token *jwt.Token) (any, error) {
			return secret, nil
		})

		if err != nil {
			common.Error(c, common.CodeInvalidToken, "令牌无效或已过期")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*service.Claims)
		if !ok || !token.Valid {
			common.Error(c, common.CodeInvalidToken, "令牌无效")
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("uid", claims.Uid)
		c.Set("username", claims.Name)
		c.Next()
	}
}
