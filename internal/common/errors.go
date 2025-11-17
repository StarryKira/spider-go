package common

// 错误码定义
const (
	// 通用错误码
	CodeSuccess       = 0
	CodeInternalError = 10001
	CodeInvalidParams = 10002
	CodeUnauthorized  = 10003
	CodeForbidden     = 10004
	CodeNotFound      = 10005

	// 用户相关错误码
	CodeUserNotFound      = 20001
	CodeUserAlreadyExists = 20002
	CodeInvalidPassword   = 20003
	CodeInvalidToken      = 20004

	// 教务系统相关错误码
	CodeJwcNotBound      = 30001
	CodeJwcLoginFailed   = 30002
	CodeJwcInvalidParams = 30003
	CodeJwcRequestFailed = 30004
	CodeJwcParseFailed   = 30005

	// 缓存相关错误码
	CodeCacheError = 40001
)

// 错误信息映射
var errorMessages = map[int]string{
	CodeSuccess:           "成功",
	CodeInternalError:     "内部服务错误",
	CodeInvalidParams:     "参数错误",
	CodeUnauthorized:      "未授权",
	CodeForbidden:         "禁止访问",
	CodeNotFound:          "资源不存在",
	CodeUserNotFound:      "用户不存在",
	CodeUserAlreadyExists: "用户已存在",
	CodeInvalidPassword:   "密码错误",
	CodeInvalidToken:      "令牌无效",
	CodeJwcNotBound:       "请先绑定教务系统账号",
	CodeJwcLoginFailed:    "教务系统登录失败",
	CodeJwcInvalidParams:  "教务系统参数错误",
	CodeJwcRequestFailed:  "教务系统请求失败",
	CodeJwcParseFailed:    "教务系统数据解析失败",
	CodeCacheError:        "缓存错误",
}

// GetErrorMessage 获取错误信息
func GetErrorMessage(code int) string {
	if msg, ok := errorMessages[code]; ok {
		return msg
	}
	return "未知错误"
}

// AppError 应用错误
type AppError struct {
	Code    int
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

// NewAppError 创建应用错误
func NewAppError(code int, message string) *AppError {
	if message == "" {
		message = GetErrorMessage(code)
	}
	return &AppError{
		Code:    code,
		Message: message,
	}
}
