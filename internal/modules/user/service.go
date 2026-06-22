package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"spider-go/internal/common"
	"spider-go/internal/service"
	"spider-go/internal/shared"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = common.NewAppError(common.CodeInvalidPassword, "用户不存在或密码错误")
	ErrEmailAlreadyExists = common.NewAppError(common.CodeUserAlreadyExists, "邮箱已被注册")
	ErrInvalidCaptcha     = common.NewAppError(common.CodeCaptchaInvalid, "验证码错误")
	ErrEmptyParams        = common.NewAppError(common.CodeInvalidParams, "参数不能为空")
	ErrWeChatAlreadyBind  = common.NewAppError(common.CodeWeChatAlreadyBind, "该微信已绑定其他账号")
)

// Service 用户服务接口
type Service interface {
	// Register 注册
	Register(ctx context.Context, email, password, name, captcha string) (token string, err error)
	// Login 用户登录
	Login(ctx context.Context, email, password string) (token string, user *User, err error)
	// ResetPassword 重置密码
	ResetPassword(ctx context.Context, email, newPassword, captcha string) error
	// WeChatLogin 微信注册登录相关
	WeChatLogin(ctx context.Context, code string) (token string, user *User, err error)
	// WeChatBind 老用户绑定微信
	WeChatBind(ctx context.Context, uid int, code string) (err error)
	// GetUserInfo 用户信息
	GetUserInfo(ctx context.Context, uid int) (*User, error)

	// BindJwc 教务系统绑定相关
	BindJwc(ctx context.Context, uid int, sid, spwd, ipAddress, userAgent string) error
	// CheckIsBind 检查是否绑定教务处
	CheckIsBind(ctx context.Context, uid int) (bool, error)
	// GetBindStatus 获取绑定状态（包含绑定次数信息）
	GetBindStatus(ctx context.Context, uid int) (*BindStatusResponse, error)

	// UpdateName 更新用户名
	UpdateName(ctx context.Context, uid int, name string) error
	// UpdateEmail 更新邮箱（需要验证码）
	UpdateEmail(ctx context.Context, uid int, email, captcha string) error
}

// userService 用户服务实现
type userService struct {
	repo           Repository
	sessionService service.SessionService
	captchaService CaptchaService
	dauService     service.DAUService
	jwtSecret      []byte
	jwtIssuer      string
	jwtExpire      time.Duration
	appid          string
	appsecret      string
}

// NewService 创建用户服务
func NewService(
	repo Repository,
	sessionService service.SessionService,
	captchaService CaptchaService,
	dauService service.DAUService,
	jwtSecret string,
	jwtIssuer string,
	appid string,
	appsecret string,
) Service {
	return &userService{
		repo:           repo,
		sessionService: sessionService,
		captchaService: captchaService,
		dauService:     dauService,
		jwtSecret:      []byte(jwtSecret),
		jwtIssuer:      jwtIssuer,
		jwtExpire:      168 * time.Hour, // 7天
		appid:          appid,
		appsecret:      appsecret,
	}
}

// Register 用户注册
func (s *userService) Register(ctx context.Context, email, password, name, captcha string) (string, error) {
	// 检查用户是否已存在
	existing, err := s.repo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		return "", err
	}
	if existing != nil {
		return "", ErrEmailAlreadyExists
	}

	// 验证验证码
	if err := s.captchaService.VerifyEmailCaptcha(ctx, email, captcha); err != nil {
		return "", ErrInvalidCaptcha
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	// 创建用户
	user := &User{
		Name:      name,
		Email:     email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return "", err
	}

	// 生成JWT token
	claims := shared.UserClaims{
		Uid:  user.Uid,
		Name: user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtExpire)),
			Issuer:    s.jwtIssuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Login 用户登录
func (s *userService) Login(ctx context.Context, email, password string) (string, *User, error) {
	// 查找用户
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", nil, ErrInvalidCredentials
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	// 记录DAU
	_ = s.dauService.RecordUserActivity(ctx, user.Uid)

	// 生成JWT token
	claims := shared.UserClaims{
		Uid:  user.Uid,
		Name: user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtExpire)),
			Issuer:    s.jwtIssuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", nil, err
	}

	return tokenString, user, nil
}

// ResetPassword 重置密码
func (s *userService) ResetPassword(ctx context.Context, email, newPassword, captcha string) error {
	// 查找用户
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return ErrUserNotFound
	}

	// 验证验证码
	if err := s.captchaService.VerifyEmailCaptcha(ctx, email, captcha); err != nil {
		return ErrInvalidCaptcha
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 更新密码
	return s.repo.UpdatePassword(ctx, user.Uid, string(hashedPassword))
}

// GetUserInfo 获取用户信息
func (s *userService) GetUserInfo(ctx context.Context, uid int) (*User, error) {
	user, err := s.repo.FindByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// BindJwc 绑定教务系统（学号一旦绑定不可更换）
func (s *userService) BindJwc(ctx context.Context, uid int, sid, spwd, ipAddress, userAgent string) error {
	if sid == "" || spwd == "" {
		return ErrEmptyParams
	}

	// 判断教务系统密码含有大小写字符，数字
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(spwd)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(spwd)
	hasDigit := regexp.MustCompile(`\d`).MatchString(spwd)
	if !(hasUpper && hasLower && hasDigit) {
		return common.NewAppError(common.CodeInvalidParams, "请绑定i中南林APP账号")
	}

	// 1. 查询用户当前绑定状态
	user, err := s.repo.FindByID(ctx, uid)
	if err != nil {
		return err
	}

	// 2. 检查是否已绑定学号（学号一旦绑定不可更换）
	if user.Sid != "" && user.Sid != sid {
		_ = s.logBindAttempt(ctx, uid, user.Sid, sid, BindStatusFailedLimit, "学号已绑定，不允许更换", ipAddress, userAgent)
		return common.NewAppError(common.CodeBindLimitExceeded, "学号已绑定，不允许更换。如需更换请联系管理员")
	}

	// 3. 判断是否为相同学号（只修改密码）
	isSameSid := (user.Sid != "" && user.Sid == sid)

	// 4. 验证教务系统账号
	if err := s.sessionService.LoginCheck(ctx, sid, spwd); err != nil {
		message := err.Error()
		if appErr, ok := err.(*common.AppError); ok {
			message = appErr.Message
			_ = s.logBindAttempt(ctx, uid, user.Sid, sid, BindStatusFailedAuth, message, ipAddress, userAgent)
			return appErr
		}

		_ = s.logBindAttempt(ctx, uid, user.Sid, sid, BindStatusFailedAuth, message, ipAddress, userAgent)
		return common.NewAppError(common.CodeJwcLoginFailed, message)
	}

	// 5. 开启事务：更新绑定信息
	tx := s.repo.(*repository).db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 5.1 更新用户表
	oldSid := user.Sid
	now := time.Now()

	// 基础更新字段
	updates := map[string]interface{}{
		"sid":          sid,
		"spwd":         spwd,
		"last_bind_at": now,
	}

	// 首次绑定时记录绑定次数
	if !isSameSid {
		updates["total_bind_count"] = user.TotalBindCount + 1
	}

	if err := tx.WithContext(ctx).Model(&User{}).Where("uid = ?", uid).Updates(updates).Error; err != nil {
		tx.Rollback()
		_ = s.logBindAttempt(ctx, uid, oldSid, sid, BindStatusFailedOther, fmt.Sprintf("更新数据库失败: %v", err), ipAddress, userAgent)
		return common.NewAppError(common.CodeDatabaseError, "绑定失败，请稍后重试")
	}

	// 5.2 记录绑定日志
	log := &JwcBindLog{
		Uid:        uid,
		OldSid:     oldSid,
		NewSid:     sid,
		BindStatus: BindStatusSuccess,
		IpAddress:  ipAddress,
		UserAgent:  userAgent,
		CreatedAt:  now,
	}
	if err := tx.WithContext(ctx).Create(log).Error; err != nil {
		tx.Rollback()
		return common.NewAppError(common.CodeDatabaseError, "记录日志失败")
	}

	// 5.3 提交事务
	if err := tx.Commit().Error; err != nil {
		return common.NewAppError(common.CodeDatabaseError, "提交事务失败")
	}

	// 6. 清除旧的教务系统会话缓存
	_ = s.sessionService.InvalidateSession(ctx, uid)

	return nil
}

// logBindAttempt 记录绑定尝试日志（辅助方法）
func (s *userService) logBindAttempt(ctx context.Context, uid int, oldSid, newSid string, status int, errMsg, ipAddress, userAgent string) error {
	log := &JwcBindLog{
		Uid:        uid,
		OldSid:     oldSid,
		NewSid:     newSid,
		BindStatus: status,
		ErrorMsg:   errMsg,
		IpAddress:  ipAddress,
		UserAgent:  userAgent,
		CreatedAt:  time.Now(),
	}
	return s.repo.(*repository).db.WithContext(ctx).Create(log).Error
}

// GetBindStatus 获取绑定状态
func (s *userService) GetBindStatus(ctx context.Context, uid int) (*BindStatusResponse, error) {
	user, err := s.repo.FindByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	return &BindStatusResponse{
		IsBound:        user.Sid != "" && user.Spwd != "",
		CurrentSid:     user.Sid,
		TotalBindCount: user.TotalBindCount,
		LastBindAt:     user.LastBindAt,
		CanChangeSid:   user.Sid == "", // 只有未绑定时才能更换学号
	}, nil
}

// CheckIsBind 检查是否绑定教务系统
func (s *userService) CheckIsBind(ctx context.Context, uid int) (bool, error) {
	user, err := s.repo.FindByID(ctx, uid)
	if err != nil {
		return false, err
	}

	return user.Sid != "" && user.Spwd != "", nil
}

// WeChatLogin 微信登录/注册
func (s *userService) WeChatLogin(ctx context.Context, code string) (string, *User, error) {
	// 1. 使用code换取openid和unionid
	wxInfo, err := s.code2Session(ctx, code)
	if err != nil {
		return "", nil, err
	}

	// 2. 查找是否存在该openid的绑定
	user, err := s.repo.FindByWeChatOpenID(ctx, s.appid, wxInfo.OpenID)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		return "", nil, err
	}

	// 3. 如果不存在，创建新用户并绑定
	if user == nil {
		user, err = s.createUserFromWeChat(ctx, wxInfo)
		if err != nil {
			return "", nil, err
		}
	} else {
		// 更新最后登录时间和unionid（如果有）
		if err := s.updateWeChatLoginInfo(ctx, user.Uid, wxInfo); err != nil {
			return "", nil, err
		}
	}

	// 记录DAU
	_ = s.dauService.RecordUserActivity(ctx, user.Uid)

	// 生成JWT token
	claims := shared.UserClaims{
		Uid:  user.Uid,
		Name: user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtExpire)),
			Issuer:    s.jwtIssuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", nil, err
	}

	return tokenString, user, nil
}

// WeChatBind 老用户绑定微信
func (s *userService) WeChatBind(ctx context.Context, uid int, code string) error {
	// 1. 使用code换取openid和unionid
	wxInfo, err := s.code2Session(ctx, code)
	if err != nil {
		return err
	}

	// 2. 检查openid是否已被其他用户绑定
	existingUser, err := s.repo.FindByWeChatOpenID(ctx, s.appid, wxInfo.OpenID)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		return err
	}
	if existingUser != nil && existingUser.Uid != uid {
		return ErrWeChatAlreadyBind
	}

	// 3. 检查当前用户是否已绑定微信
	existingBind, err := s.repo.FindWeChatBindByUID(ctx, uid, s.appid)
	if err == nil && existingBind != nil {
		// 已存在绑定，更新信息
		existingBind.OpenID = wxInfo.OpenID
		existingBind.UnionID = wxInfo.UnionID
		existingBind.LastLogin = time.Now()
		existingBind.UpdatedAt = time.Now()
		return s.repo.UpdateWeChatBind(ctx, existingBind)
	}

	// 4. 创建新的绑定关系
	bind := &UserWeChatMiniProgram{
		Uid:       uid,
		AppID:     s.appid,
		OpenID:    wxInfo.OpenID,
		UnionID:   wxInfo.UnionID,
		LastLogin: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return s.repo.CreateWeChatBind(ctx, bind)
}

// CheckIsWeChatBind 检查用户是否绑定微信
func (s *userService) CheckIsWeChatBind(ctx context.Context, uid int) (bool, error) {
	isExistUser, err := s.repo.FindWeChatBindByUID(ctx, uid, s.appid)
	if err != nil {
		return false, err
	}
	return isExistUser != nil, nil
}

// WeChatSessionResponse 微信登录响应
type WeChatSessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// code2Session 使用code换取openid和unionid
func (s *userService) code2Session(ctx context.Context, code string) (*WeChatSessionResponse, error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		s.appid, s.appsecret, code,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, common.NewAppError(common.CodeHttpRequestFailed, "创建请求失败")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, common.NewAppError(common.CodeWeChatLoginFailed, "请求微信接口失败")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, common.NewAppError(common.CodeWeChatLoginFailed, "读取响应失败")
	}

	var wxResp WeChatSessionResponse
	if err := json.Unmarshal(body, &wxResp); err != nil {
		return nil, common.NewAppError(common.CodeInvalidResponse, "解析微信响应失败")
	}

	if wxResp.ErrCode != 0 {
		return nil, common.NewAppError(common.CodeWeChatLoginFailed, fmt.Sprintf("微信接口返回错误: %d - %s", wxResp.ErrCode, wxResp.ErrMsg))
	}

	if wxResp.OpenID == "" {
		return nil, common.NewAppError(common.CodeWeChatLoginFailed, "未获取到OpenID")
	}

	return &wxResp, nil
}

// createUserFromWeChat 从微信信息创建用户
func (s *userService) createUserFromWeChat(ctx context.Context, wxInfo *WeChatSessionResponse) (*User, error) {
	// 生成默认用户名
	defaultName := fmt.Sprintf("微信用户_%s", wxInfo.OpenID[len(wxInfo.OpenID)-8:])
	defaultEmail := fmt.Sprintf("wx_%s@wechat.local", wxInfo.OpenID)

	user := &User{
		Name:      defaultName,
		Email:     defaultEmail,
		Password:  "",
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, common.NewAppError(common.CodeDatabaseError, "创建用户失败")
	}

	bind := &UserWeChatMiniProgram{
		Uid:       user.Uid,
		AppID:     s.appid,
		OpenID:    wxInfo.OpenID,
		UnionID:   wxInfo.UnionID,
		LastLogin: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.CreateWeChatBind(ctx, bind); err != nil {
		return nil, common.NewAppError(common.CodeWeChatBindFailed, "创建微信绑定失败")
	}

	return user, nil
}

// updateWeChatLoginInfo 更新微信登录信息
func (s *userService) updateWeChatLoginInfo(ctx context.Context, uid int, wxInfo *WeChatSessionResponse) error {
	bind, err := s.repo.FindWeChatBindByUID(ctx, uid, s.appid)
	if err != nil {
		return err
	}

	bind.UnionID = wxInfo.UnionID
	bind.LastLogin = time.Now()
	bind.UpdatedAt = time.Now()

	return s.repo.UpdateWeChatBind(ctx, bind)
}

// UpdateName 更新用户名
func (s *userService) UpdateName(ctx context.Context, uid int, name string) error {
	if name == "" {
		return ErrEmptyParams
	}

	// 获取用户
	user, err := s.repo.FindByID(ctx, uid)
	if err != nil {
		return err
	}

	// 更新用户名
	user.Name = name
	return s.repo.Update(ctx, user)
}

// UpdateEmail 更新邮箱（需要验证码）
func (s *userService) UpdateEmail(ctx context.Context, uid int, email, captcha string) error {
	if email == "" {
		return ErrEmptyParams
	}

	// 验证验证码
	if err := s.captchaService.VerifyEmailCaptcha(ctx, email, captcha); err != nil {
		return ErrInvalidCaptcha
	}

	// 检查新邮箱是否已被使用
	existingUser, err := s.repo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		return err
	}
	if existingUser != nil && existingUser.Uid != uid {
		return ErrEmailAlreadyExists
	}

	// 获取当前用户
	user, err := s.repo.FindByID(ctx, uid)
	if err != nil {
		return err
	}

	// 更新邮箱
	user.Email = email
	return s.repo.Update(ctx, user)
}
