package api

import (
	"strings"

	"protocol-parser-server/repository/paging"

	"github.com/gin-gonic/gin"
)

func pageQuery(c *gin.Context) paging.Query {
	return paging.Query{
		Page:              positiveInt(c.Query("page"), 1),
		PageSize:          positiveInt(c.Query("pageSize"), 10),
		Keyword:           strings.TrimSpace(c.Query("keyword")),
		Status:            strings.TrimSpace(c.Query("status")),
		OrganizationID:    strings.TrimSpace(c.Query("organizationId")),
		Category:          strings.TrimSpace(c.Query("category")),
		VehicleType:       strings.TrimSpace(c.Query("vehicleType")),
		InventoryStatus:   strings.TrimSpace(c.Query("inventoryStatus")),
		Protocol:          strings.TrimSpace(c.Query("protocol")),
		MaintenanceStatus: strings.TrimSpace(c.Query("maintenanceStatus")),
		MaintenanceType:   strings.TrimSpace(c.Query("maintenanceType")),
		BindingStatus:     strings.TrimSpace(c.Query("bindingStatus")),
		DeviceNo:          strings.TrimSpace(c.Query("deviceNo")),
		VehicleID:         strings.TrimSpace(c.Query("vehicleId")),
		RecordType:        strings.TrimSpace(c.Query("recordType")),
		RoleID:            strings.TrimSpace(c.Query("roleId")),
		MenuType:          strings.TrimSpace(c.Query("menuType")),
		CompanyID:         strings.TrimSpace(c.Query("companyId")),
		SortField:         strings.TrimSpace(c.Query("sortField")),
		SortOrder:         strings.TrimSpace(c.Query("sortOrder")),
	}
}

func pagedContext(c *gin.Context) *paging.State {
	ctx, state := paging.WithContext(c.Request.Context(), pageQuery(c))
	c.Request = c.Request.WithContext(ctx)
	return state
}

// respondPagedData preserves the frontend contract while repositories perform
// filtering, sorting and pagination directly in the database.
func respondPagedData(c *gin.Context, key string, data any, state *paging.State, err error) {
	if err != nil {
		respondData(c, key, data, err)
		return
	}
	query := state.Query
	pageSize, _ := query.LimitOffset()
	c.JSON(200, gin.H{"success": true, key: data, "items": data, "total": state.Total, "page": query.Page, "pageSize": pageSize})
}
