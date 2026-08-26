package settings

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
)

type LoginPage struct {
	LayoutType       string  `json:"layoutType"`
	SplitImage       string  `json:"splitImage"`
	BackgroundImage  string  `json:"backgroundImage"`
	OverlayOpacity   float64 `json:"overlayOpacity"`
	AnimationEnabled bool    `json:"animationEnabled"`
	NavigationType   string  `json:"navigationType"`
}

type Store interface {
	GetLoginPage(context.Context) (*LoginPage, error)
	SaveLoginPage(context.Context, LoginPage) error
}

type MySQLStore struct{ db *sql.DB }

func NewMySQLStore(db *sql.DB) *MySQLStore { return &MySQLStore{db: db} }

func (store *MySQLStore) GetLoginPage(ctx context.Context) (*LoginPage, error) {
	var raw []byte
	if err := store.db.QueryRowContext(ctx, `SELECT setting_value FROM system_settings WHERE setting_key='login_page'`).Scan(&raw); err != nil {
		return nil, err
	}
	var result LoginPage
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (store *MySQLStore) SaveLoginPage(ctx context.Context, value LoginPage) error {
	if value.LayoutType != "split" && value.LayoutType != "background" {
		return errors.New("不支持的登录页类型")
	}
	if value.NavigationType == "" {
		value.NavigationType = "sidebar"
	}
	if value.NavigationType != "sidebar" && value.NavigationType != "top" {
		return errors.New("不支持的后台导航类型")
	}
	if value.SplitImage == "" {
		value.SplitImage = "/iot-login-hero-v2.png"
	}
	if value.BackgroundImage == "" {
		value.BackgroundImage = "/iot-login-fullscreen-clean.png"
	}
	if !validImageURL(value.SplitImage) {
		return errors.New("分栏图片地址仅支持站内路径或HTTP/HTTPS地址")
	}
	if !validImageURL(value.BackgroundImage) {
		return errors.New("背景图片地址仅支持站内路径或HTTP/HTTPS地址")
	}
	if value.OverlayOpacity < 0 || value.OverlayOpacity > 0.8 {
		return errors.New("背景遮罩强度必须在0到0.8之间")
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = store.db.ExecContext(ctx, `INSERT INTO system_settings (setting_key,setting_value) VALUES ('login_page', CAST(? AS JSON)) ON DUPLICATE KEY UPDATE setting_value=VALUES(setting_value)`, raw)
	return err
}

func validImageURL(value string) bool {
	if strings.HasPrefix(value, "/") && !strings.ContainsAny(value, "\"'()\\") {
		return true
	}
	parsed, err := url.ParseRequestURI(value)
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}
