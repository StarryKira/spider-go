package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	MFAPurposeBind  = "bind"
	MFAPurposeLogin = "login"
)

// MFAChallenge 教务 CAS 安全手机验证进行中的状态
type MFAChallenge struct {
	UID         int            `json:"uid"`
	Purpose     string         `json:"purpose"`
	SID         string         `json:"sid"`
	Password    string         `json:"password"`
	Phone       string         `json:"phone"`
	State       string         `json:"state"`
	GID         string         `json:"gid"`
	AttestURL   string         `json:"attest_url"`
	VisitorID   string         `json:"visitor_id"`
	Execution   string         `json:"execution"`
	LoginURL    string         `json:"login_url"`
	RedirectURL string         `json:"redirect_url"`
	Cookies     []*http.Cookie `json:"cookies"`
}

// MFAChallengeCache 手机验证码挑战缓存
type MFAChallengeCache interface {
	Set(ctx context.Context, uid int, challenge *MFAChallenge, expiration time.Duration) error
	Get(ctx context.Context, uid int) (*MFAChallenge, error)
	Delete(ctx context.Context, uid int) error
}

type redisMFAChallengeCache struct {
	client *redis.Client
}

func NewRedisMFAChallengeCache(client *redis.Client) MFAChallengeCache {
	return &redisMFAChallengeCache{client: client}
}

func (c *redisMFAChallengeCache) Set(ctx context.Context, uid int, challenge *MFAChallenge, expiration time.Duration) error {
	data, err := json.Marshal(challenge)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.key(uid), data, expiration).Err()
}

func (c *redisMFAChallengeCache) Get(ctx context.Context, uid int) (*MFAChallenge, error) {
	data, err := c.client.Get(ctx, c.key(uid)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var challenge MFAChallenge
	if err := json.Unmarshal(data, &challenge); err != nil {
		return nil, err
	}
	return &challenge, nil
}

func (c *redisMFAChallengeCache) Delete(ctx context.Context, uid int) error {
	return c.client.Del(ctx, c.key(uid)).Err()
}

func (c *redisMFAChallengeCache) key(uid int) string {
	return fmt.Sprintf("mfa:phone:%d", uid)
}
