package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html"
	"strings"
	"sync"
	"time"
)

const captchaAlphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"

type captchaEntry struct {
	code    string
	expires time.Time
}
type Captcha struct {
	ID    string `json:"id"`
	Image string `json:"image"`
}
type CaptchaManager struct {
	mu      sync.Mutex
	entries map[string]captchaEntry
}

func NewCaptchaManager() *CaptchaManager {
	return &CaptchaManager{entries: make(map[string]captchaEntry)}
}

func (manager *CaptchaManager) Create() (*Captcha, error) {
	idBytes := make([]byte, 18)
	if _, err := rand.Read(idBytes); err != nil {
		return nil, err
	}
	codeBytes := make([]byte, 4)
	random := make([]byte, 4)
	if _, err := rand.Read(random); err != nil {
		return nil, err
	}
	for index, value := range random {
		codeBytes[index] = captchaAlphabet[int(value)%len(captchaAlphabet)]
	}
	id := base64.RawURLEncoding.EncodeToString(idBytes)
	code := string(codeBytes)
	manager.mu.Lock()
	now := time.Now()
	for key, item := range manager.entries {
		if now.After(item.expires) {
			delete(manager.entries, key)
		}
	}
	manager.entries[id] = captchaEntry{code: code, expires: now.Add(3 * time.Minute)}
	manager.mu.Unlock()
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="132" height="42" viewBox="0 0 132 42"><rect width="132" height="42" rx="6" fill="#f0f6ff"/><path d="M5 31L127 8M8 9L124 34M20 39L105 3" stroke="#aacbfa" stroke-width="1"/><text x="66" y="29" text-anchor="middle" font-family="Arial,sans-serif" font-size="24" font-weight="700" letter-spacing="6" fill="#175eb8">%s</text></svg>`, html.EscapeString(code))
	return &Captcha{ID: id, Image: "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))}, nil
}

func (manager *CaptchaManager) Verify(id, code string) bool {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	entry, exists := manager.entries[id]
	delete(manager.entries, id)
	return exists && time.Now().Before(entry.expires) && strings.EqualFold(entry.code, strings.TrimSpace(code))
}
