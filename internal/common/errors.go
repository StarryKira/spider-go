package common

// 错误码定义
const (
	CodeSuccess           = 0     // 成功
	CodeInvalidParams     = 40000 // 参数错误
	CodeUnauthorized      = 40100 // 未授权
	CodeInvalidToken      = 40101 // Token无效
	CodeForbidden         = 40300 // 禁止访问
	CodeUserNotFound      = 40400 // 用户不存在
	CodeNotFound          = 40404 // 资源不存在
	CodeInvalidPassword   = 40100 // 密码错误
	CodeUserAlreadyExists = 40900 // 用户已存在
	CodeCaptchaInvalid    = 40001 // 验证码错误
	CodeInternalError     = 50000 // 内部错误
	CodeJwcInvalidParams  = 40002 // 教务系统参数错误
	CodeJwcNotBound       = 40003 // 教务系统未绑定
	CodeJwcLoginFailed    = 40004 // 教务系统登录失败
	CodeJwcParseFailed    = 40005 // 教务系统解析失败
	CodeJwcRequestFailed  = 40006 // 教务系统请求失败
	CodeCacheError        = 50001 // 缓存错误
)

// AppError 应用错误
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	return e.Message
}

// NewAppError 创建应用错误
func NewAppError(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}
