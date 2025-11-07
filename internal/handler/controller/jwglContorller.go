package controller

import (
	"spider-go/internal/handler/service"

	"github.com/gin-gonic/gin"
)

func Jwclogin(c *gin.Context) {
	username := c.PostForm("username")
	jwcpassword := c.PostForm("jwcpassword")

	if username == "" || jwcpassword == "" {
		c.JSON(200, gin.H{
			"code":    200,
			"message": "username or jwcpassword is empty",
		})
		return
	}

	name := service.Jwclogin(username, jwcpassword)

	c.JSON(200, gin.H{
		"code":    200,
		"message": "success",
		"name":    name,
	})

}
