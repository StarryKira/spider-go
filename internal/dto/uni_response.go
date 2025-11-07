package dto

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: http.StatusOK,
		Msg:  "success",
		Data: data,
	})
}

func Error(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Response{
		Code: code,
		Msg:  message,
		Data: nil,
	})
}

// BadRequest 400
func BadRequest(c *gin.Context, code int, message string) {
	Error(c, http.StatusBadRequest, code, message)
}

// Unauthorized 401 错误
func Unauthorized(c *gin.Context, code int, message string) {
	Error(c, http.StatusUnauthorized, code, message)
}

// Forbidden 403 错误
func Forbidden(c *gin.Context, code int, message string) {
	Error(c, http.StatusForbidden, code, message)
}

// NotFound 404 错误
func NotFound(c *gin.Context, code int, message string) {
	Error(c, http.StatusNotFound, code, message)
}

// InternalServerError 500 错误
func InternalServerError(c *gin.Context, code int, message string) {
	Error(c, http.StatusInternalServerError, code, message)
}
