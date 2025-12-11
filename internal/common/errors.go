package common

import pkgerrors "spider-go/pkg/errors"

// 重新导出 pkg/errors 的错误码（保持向后兼容）
const (
	CodeSuccess           = pkgerrors.CodeSuccess
	CodeInvalidParams     = pkgerrors.CodeInvalidParams
	CodeUnauthorized      = pkgerrors.CodeUnauthorized
	CodeInvalidToken      = pkgerrors.CodeInvalidToken
	CodeForbidden         = pkgerrors.CodeForbidden
	CodeUserNotFound      = pkgerrors.CodeUserNotFound
	CodeNotFound          = pkgerrors.CodeNotFound
	CodeInvalidPassword   = pkgerrors.CodeInvalidPassword
	CodeUserAlreadyExists = pkgerrors.CodeUserAlreadyExists
	CodeCaptchaInvalid    = pkgerrors.CodeCaptchaInvalid
	CodeInternalError     = pkgerrors.CodeInternalError
	CodeJwcInvalidParams  = pkgerrors.CodeJwcInvalidParams
	CodeJwcNotBound       = pkgerrors.CodeJwcNotBound
	CodeJwcLoginFailed    = pkgerrors.CodeJwcLoginFailed
	CodeJwcParseFailed    = pkgerrors.CodeJwcParseFailed
	CodeJwcRequestFailed  = pkgerrors.CodeJwcRequestFailed
	CodeCacheError        = pkgerrors.CodeCacheError
)

// AppError 重新导出 pkg/errors 的类型（保持向后兼容）
type AppError = pkgerrors.AppError

// NewAppError 重新导出（保持向后兼容）
var NewAppError = pkgerrors.NewAppError
