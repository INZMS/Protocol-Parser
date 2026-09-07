package api

import (
	"net/http/httptest"
	"testing"

	"protocol-parser-server/repository/paging"

	"github.com/gin-gonic/gin"
)

func TestPagedContextIsAvailableToRepositoryRequestContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/?page=3&pageSize=25&keyword=vehicle", nil)

	state := pagedContext(c)
	fromRequest := paging.FromContext(c.Request.Context())

	if fromRequest != state {
		t.Fatal("the repository request context must retain the pagination state")
	}
	if fromRequest.Query.Page != 3 || fromRequest.Query.PageSize != 25 || fromRequest.Query.Keyword != "vehicle" {
		t.Fatalf("unexpected pagination query: %+v", fromRequest.Query)
	}
}
