package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Claims struct {
	UserID   int64  `json:"sub"`
	Username string `json:"username"`
	Expires  int64  `json:"exp"`
}

type Manager struct {
	secret []byte
	ttl    time.Duration
}

func NewManagerFromEnv() *Manager {
	secret := os.Getenv("AUTH_SECRET")
	if secret == "" {
		secret = "protocol-parser-change-this-secret"
	}
	hours, err := strconv.Atoi(os.Getenv("AUTH_EXPIRE_HOURS"))
	if err != nil || hours <= 0 {
		hours = 12
	}
	return &Manager{secret: []byte(secret), ttl: time.Duration(hours) * time.Hour}
}

func (manager *Manager) Issue(userID int64, username string) (string, error) {
	header, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	payload, err := json.Marshal(Claims{UserID: userID, Username: username, Expires: time.Now().Add(manager.ttl).Unix()})
	if err != nil {
		return "", err
	}
	unsigned := encode(header) + "." + encode(payload)
	return unsigned + "." + encode(manager.sign(unsigned)), nil
}

func (manager *Manager) Verify(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("登录令牌格式错误")
	}
	unsigned := parts[0] + "." + parts[1]
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(signature, manager.sign(unsigned)) {
		return nil, errors.New("登录令牌无效")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("登录令牌内容错误")
	}
	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("解析登录令牌失败: %w", err)
	}
	if claims.UserID <= 0 || time.Now().Unix() >= claims.Expires {
		return nil, errors.New("登录已过期")
	}
	return &claims, nil
}

func (manager *Manager) sign(value string) []byte {
	mac := hmac.New(sha256.New, manager.secret)
	_, _ = mac.Write([]byte(value))
	return mac.Sum(nil)
}
func encode(value []byte) string { return base64.RawURLEncoding.EncodeToString(value) }
