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

func TestAnalyzeJT808SelectedProtocol(t *testing.T) {
	gin.SetMode(gin.TestMode)
	protocol.RegisterProtocols()
	router := gin.New()
	store := &fakeHistoryStore{}
	RegisterParserRouter(router, store)
	request := httptest.NewRequest(http.MethodPost, "/api/parser/analyze", bytes.NewBufferString(`{"protocol":"JT-808","hex":"7E020000220138001380000001000000000000000301DAE80007383280000C0258005A24082412345630011F310108547E"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, expected := range []string{`"protocol":"JT-808"`, `"messageId":"0200"`, `"messageName":"位置信息汇报"`, `"latitude":31.123456`, `"locationTime":"2024-08-24 12:34:56（北京时间）"`} {
		if !strings.Contains(body, expected) {
			t.Fatalf("response missing %s: %s", expected, body)
		}
	}
	if store.created != 1 {
		t.Fatalf("expected one saved history record, got %d", store.created)
	}
}

func TestAnalyzeVDFSelectedProtocol(t *testing.T) {
	gin.SetMode(gin.TestMode)
	protocol.RegisterProtocols()
	router := gin.New()
	store := &fakeHistoryStore{}
	RegisterParserRouter(router, store)
	request := httptest.NewRequest(http.MethodPost, "/api/parser/analyze", bytes.NewBufferString(`{"protocol":"VDF","hex":"2A48513230313836313830333137383939393030322C42412641313131313138323233333339363831313335363532313136303031363039303731382642303130303030303030302646303030302652313830372657303030303030323926493534363030303234393530453131353132343935304531323534323439353133423334323234393531323233333532363233313044373335264B34303130302654393523"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, expected := range []string{`"protocol":"VDF"`, `"messageId":"BA"`, `"messageName":"定时定位上报"`, `"latitude":22.556613333333335`, `"longitude":113.94201833333334`} {
		if !strings.Contains(body, expected) {
			t.Fatalf("response missing %s: %s", expected, body)
		}
	}
	if store.created != 1 {
		t.Fatalf("expected one saved history record, got %d", store.created)
	}
}
