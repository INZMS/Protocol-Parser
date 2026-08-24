package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"protocol-parser-server/protocol"
)

func TestAnalyze2929SelectedProtocol(t *testing.T) {
	gin.SetMode(gin.TestMode)
	protocol.RegisterProtocols()
	router := gin.New()
	store := &fakeHistoryStore{}
	RegisterParserRouter(router, store)
	request := httptest.NewRequest(http.MethodPost, "/api/parser/analyze", bytes.NewBufferString(`{"protocol":"2929","hex":"29 29 21 00 05 D0 84 C4 B4 0D"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"messageName":"中心确认"`) || !strings.Contains(response.Body.String(), `"原包主信令"`) || !strings.Contains(response.Body.String(), `"recordId":7`) {
		t.Fatalf("unexpected body: %s", response.Body.String())
	}
	if store.created != 1 {
		t.Fatalf("expected one saved history record, got %d", store.created)
	}
}
