package api

import (
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"protocol-parser-server/parser/core"
	"protocol-parser-server/repository/history"
)

type AnalyzeRequest struct {
	Hex      string `json:"hex"`
	Protocol string `json:"protocol"`
}

func RegisterParserRouter(r *gin.Engine, stores ...history.Store) {
	var store history.Store
	if len(stores) > 0 {
		store = stores[0]
	}

	r.POST("/api/parser/analyze", analyze(store))
	registerHistoryRouter(r, store)
}

func analyze(store history.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AnalyzeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}

		normalized := strings.NewReplacer(" ", "", "\n", "", "\r", "", "\t", "").Replace(req.Hex)
		data, err := hex.DecodeString(normalized)
		if err != nil || len(data) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "HEX格式错误"})
			return
		}

		var p core.Protocol
		if req.Protocol != "" {
			p, err = core.Get(req.Protocol)
		} else {
			p, err = core.Detect(data)
		}

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}

		result, err := p.Parse(data)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}

		var recordID int64
		if store != nil {
			recordID, err = store.Create(c.Request.Context(), result)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "recordId": recordID, "data": result})
	}
}
