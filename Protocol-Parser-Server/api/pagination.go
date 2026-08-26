package api

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

// respondPagedData provides one consistent paging contract for every list API.
func respondPagedData(c *gin.Context, key string, data any, err error) {
	if err != nil {
		respondData(c, key, data, err)
		return
	}
	page := positiveInt(c.Query("page"), 1)
	pageSize := positiveInt(c.Query("pageSize"), 10)
	if pageSize > 200 {
		pageSize = 200
	}
	keyword := strings.ToLower(strings.TrimSpace(c.Query("keyword")))
	status := strings.TrimSpace(c.Query("status"))
	organizationID := strings.TrimSpace(c.Query("organizationId"))
	category := strings.ToLower(strings.TrimSpace(c.Query("category")))
	vehicleType := strings.ToLower(strings.TrimSpace(c.Query("vehicleType")))
	inventoryStatus := strings.ToLower(strings.TrimSpace(c.Query("inventoryStatus")))
	protocol := strings.ToLower(strings.TrimSpace(c.Query("protocol")))
	maintenanceStatus := strings.ToLower(strings.TrimSpace(c.Query("maintenanceStatus")))
	maintenanceType := strings.ToLower(strings.TrimSpace(c.Query("maintenanceType")))
	bindingStatus := strings.TrimSpace(c.Query("bindingStatus"))
	deviceNo := strings.ToLower(strings.TrimSpace(c.Query("deviceNo")))
	v := reflect.ValueOf(data)
	filtered := reflect.MakeSlice(v.Type(), 0, v.Len())
	for i := 0; i < v.Len(); i++ {
		item := v.Index(i)
		raw, _ := json.Marshal(item.Interface())
		text := strings.ToLower(string(raw))
		if keyword != "" && !strings.Contains(text, keyword) {
			continue
		}
		if status != "" && !strings.Contains(text, "\"status\":"+status) {
			continue
		}
		if organizationID != "" && !strings.Contains(text, "\"organizationid\":"+organizationID) {
			continue
		}
		if category != "" && !strings.Contains(text, category) {
			continue
		}
		if vehicleType != "" {
			encoded, _ := json.Marshal(vehicleType)
			if !strings.Contains(text, `"vehicletype":`+string(encoded)) {
				continue
			}
		}
		if !jsonFieldMatches(text, "inventoryStatus", inventoryStatus) ||
			!jsonFieldMatches(text, "protocol", protocol) ||
			!jsonFieldMatches(text, "status", maintenanceStatus) ||
			!jsonFieldMatches(text, "maintenanceType", maintenanceType) {
			continue
		}
		if deviceNo != "" && !strings.Contains(text, deviceNo) {
			continue
		}
		if bindingStatus == "bound" && strings.Contains(text, "\"bounddevicecount\":0") {
			continue
		}
		if bindingStatus == "unbound" && !strings.Contains(text, "\"bounddevicecount\":0") {
			continue
		}
		filtered = reflect.Append(filtered, item)
	}
	total := filtered.Len()
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	items := filtered.Slice(start, end).Interface()
	c.JSON(200, gin.H{"success": true, key: items, "items": items, "total": total, "page": page, "pageSize": pageSize})
}

func jsonFieldMatches(text, field, value string) bool {
	if value == "" {
		return true
	}
	encoded, _ := json.Marshal(strings.ToLower(value))
	return strings.Contains(text, `"`+strings.ToLower(field)+`":`+string(encoded))
}
