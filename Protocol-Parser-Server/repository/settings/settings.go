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
	SystemName       string  `json:"systemName"`
	SystemNameEn     string  `json:"systemNameEn"`
	MenuShortName    string  `json:"menuShortName"`
	BrowserTitleMode string  `json:"browserTitleMode"`
	BrowserTitle     string  `json:"browserTitle"`
	SystemIcon       string  `json:"systemIcon"`
	FooterCopyright  string  `json:"footerCopyright"`
	FooterSlogan     string  `json:"footerSlogan"`
	DeveloperName    string  `json:"developerName"`
	DeveloperPhone   string  `json:"developerPhone"`
	SystemVersion    string  `json:"systemVersion"`
}

type Store interface {
	GetLoginPage(context.Context) (*LoginPage, error)
	SaveLoginPage(context.Context, LoginPage) error
	GetUserPreference(context.Context, int64, string) (json.RawMessage, error)
	SaveUserPreference(context.Context, int64, string, json.RawMessage) error
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
	if strings.TrimSpace(value.SystemName) == "" {
		value.SystemName = "协议解析工具"
	}
	if strings.TrimSpace(value.SystemNameEn) == "" {
		value.SystemNameEn = "Protocol Parser Tool"
	}
	if strings.TrimSpace(value.MenuShortName) == "" {
		value.MenuShortName = value.SystemName
	}
	if len([]rune(strings.TrimSpace(value.MenuShortName))) > 12 {
		return errors.New("系统简称不能超过12个字符")
	}
	if strings.TrimSpace(value.FooterCopyright) == "" {
		value.FooterCopyright = value.SystemName
	}
	if strings.TrimSpace(value.FooterSlogan) == "" {
		value.FooterSlogan = "让协议解析更简单高效"
	}
	if len([]rune(value.FooterCopyright)) > 60 || len([]rune(value.FooterSlogan)) > 100 || len([]rune(value.DeveloperName)) > 80 || len([]rune(value.DeveloperPhone)) > 40 || len([]rune(value.SystemVersion)) > 30 {
		return errors.New("底部信息内容超过允许长度")
	}
	if value.BrowserTitleMode == "" {
		value.BrowserTitleMode = "system"
	}
	if value.BrowserTitleMode != "system" && value.BrowserTitleMode != "menu" && value.BrowserTitleMode != "custom" {
		return errors.New("不支持的浏览器标题模式")
	}
	if value.BrowserTitleMode == "custom" && strings.TrimSpace(value.BrowserTitle) == "" {
		return errors.New("请输入浏览器标签页标题")
	}
	if value.SystemIcon == "" {
		value.SystemIcon = "/favicon.png"
	}
	if !validImageURL(value.SystemIcon) {
		return errors.New("系统图标地址仅支持站内路径或HTTP/HTTPS地址")
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

func (store *MySQLStore) GetUserPreference(ctx context.Context, userID int64, key string) (json.RawMessage, error) {
	var raw []byte
	err := store.db.QueryRowContext(ctx, `SELECT preference_value FROM user_preferences WHERE user_id=? AND preference_key=?`, userID, key).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return json.RawMessage(`null`), nil
	}
	return json.RawMessage(raw), err
}

func (store *MySQLStore) SaveUserPreference(ctx context.Context, userID int64, key string, value json.RawMessage) error {
	if strings.TrimSpace(key) == "" || len(key) > 128 || !json.Valid(value) {
		return errors.New("用户偏好格式错误")
	}
	_, err := store.db.ExecContext(ctx, `INSERT INTO user_preferences(user_id,preference_key,preference_value) VALUES(?,?,CAST(? AS JSON)) ON DUPLICATE KEY UPDATE preference_value=VALUES(preference_value)`, userID, key, []byte(value))
	return err
}

func validImageURL(value string) bool {
	if strings.HasPrefix(value, "/") && !strings.ContainsAny(value, "\"'()\\") {
		return true
	}
	parsed, err := url.ParseRequestURI(value)
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}
