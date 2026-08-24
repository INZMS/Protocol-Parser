package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"protocol-parser-server/parser/core"
	"protocol-parser-server/repository/history"
)

type fakeHistoryStore struct {
	created int
	deleted int64
	cleared bool
}

func (s *fakeHistoryStore) Create(_ context.Context, _ *core.ParseResult) (int64, error) {
	s.created++
	return 7, nil
}
func (s *fakeHistoryStore) List(_ context.Context, query history.Query) (*history.Page, error) {
	return &history.Page{Items: []history.Summary{{ID: 7, Protocol: "2929", MessageID: "80", MessageName: "一般位置数据", Length: 32, Time: "2026-08-24 12:00:00"}}, Total: 1, Page: query.Page, PageSize: query.PageSize}, nil
}
func (s *fakeHistoryStore) Get(_ context.Context, id int64) (*history.Record, error) {
	if id != 7 {
		return nil, history.ErrNotFound
	}
	return &history.Record{Summary: history.Summary{ID: 7, Time: "2026-08-24 12:00:00"}, Result: &core.ParseResult{Protocol: "2929", MessageID: "80", MessageName: "一般位置数据", Length: 32, Raw: "2929"}}, nil
}
func (s *fakeHistoryStore) Delete(_ context.Context, id int64) error {
	s.deleted = id
	return nil
}
func (s *fakeHistoryStore) Clear(context.Context) error {
	s.cleared = true
	return nil
}

func TestHistoryRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeHistoryStore{}
	router := gin.New()
	RegisterParserRouter(router, store)

	tests := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/api/parser/history?page=1&pageSize=5", `"total":1`},
		{http.MethodGet, "/api/parser/history/7", `"raw":"2929"`},
		{http.MethodDelete, "/api/parser/history/7", `"success":true`},
		{http.MethodDelete, "/api/parser/history", `"success":true`},
	}
	for _, test := range tests {
		request := httptest.NewRequest(test.method, test.path, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusOK || !contains(response.Body.String(), test.body) {
			t.Fatalf("%s %s status=%d body=%s", test.method, test.path, response.Code, response.Body.String())
		}
	}
	if store.deleted != 7 || !store.cleared {
		t.Fatalf("history mutations not called: deleted=%d cleared=%v", store.deleted, store.cleared)
	}
}

func contains(value, fragment string) bool {
	for i := 0; i+len(fragment) <= len(value); i++ {
		if value[i:i+len(fragment)] == fragment {
			return true
		}
	}
	return false
}
