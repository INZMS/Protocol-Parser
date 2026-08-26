package auth

import "testing"

func TestCaptchaIsOneTime(t *testing.T) {
	manager := NewCaptchaManager()
	result, err := manager.Create()
	if err != nil {
		t.Fatal(err)
	}
	manager.mu.Lock()
	code := manager.entries[result.ID].code
	manager.mu.Unlock()
	if !manager.Verify(result.ID, code) {
		t.Fatal("expected captcha to verify")
	}
	if manager.Verify(result.ID, code) {
		t.Fatal("captcha must be consumed after first verification")
	}
}
