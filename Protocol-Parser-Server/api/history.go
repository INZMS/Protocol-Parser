package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"protocol-parser-server/repository/history"
)

func registerHistoryRouter(r *gin.Engine, store history.Store) {
	if store == nil {
		return
	}
	r.GET("/api/parser/history", listHistory(store))
	r.GET("/api/parser/history/:id", getHistory(store))
	r.DELETE("/api/parser/history/:id", deleteHistory(store))
	r.DELETE("/api/parser/history", clearHistory(store))
}

func listHistory(store history.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		page := positiveInt(c.Query("page"), 1)
		pageSize := positiveInt(c.Query("pageSize"), 5)
		result, err := store.List(c.Request.Context(), history.Query{
			Keyword: c.Query("keyword"), Page: page, PageSize: pageSize,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
	}
}

func getHistory(store history.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := historyID(c)
		if !ok {
			return
		}
		record, err := store.Get(c.Request.Context(), id)
		if err != nil {
			historyError(c, err)
			return
		}
		result := record.Result
		c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
			"id": record.ID, "time": record.Time,
			"protocol": result.Protocol, "messageId": result.MessageID,
			"messageName": result.MessageName, "length": result.Length,
			"raw": result.Raw, "fields": result.Fields, "data": result.Data,
		}})
	}
}

func deleteHistory(store history.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := historyID(c)
		if !ok {
			return
		}
		if err := store.Delete(c.Request.Context(), id); err != nil {
			historyError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

func clearHistory(store history.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := store.Clear(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

func historyID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "无效的记录ID"})
		return 0, false
	}
	return id, true
}

func historyError(c *gin.Context, err error) {
	if errors.Is(err, history.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": history.ErrNotFound.Error()})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
}

func positiveInt(value string, fallback int) int {
	number, err := strconv.Atoi(value)
	if err != nil || number <= 0 {
		return fallback
	}
	return number
}
