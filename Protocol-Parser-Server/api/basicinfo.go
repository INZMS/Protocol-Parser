package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"protocol-parser-server/auth"
	"protocol-parser-server/repository/basicinfo"
	"protocol-parser-server/repository/datascope"
	"protocol-parser-server/repository/user"
)

func RegisterBasicInfoRouter(r *gin.Engine, store basicinfo.Store, users user.Store, tokens *auth.Manager) {
	g := r.Group("/api/basic-info")
	g.Use(AuthMiddleware(tokens))
	// Data range is resolved once per request and carried through to the
	// repository.  UI filtering is intentionally not trusted for isolation.
	g.Use(func(c *gin.Context) {
		scope, err := users.DataScope(c.Request.Context(), currentUserID(c))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "error": "无法确认当前用户的数据范围"})
			return
		}
		c.Request = c.Request.WithContext(datascope.With(c.Request.Context(), scope))
		c.Next()
	})
	g.GET("/notifications", func(c *gin.Context) { v, e := store.ListNotifications(c); respondData(c, "notifications", v, e) })
	g.PUT("/notifications/:id/read", func(c *gin.Context) { respondOK(c, store.ReadNotification(c, parseID(c))) })
	g.POST("/upload", requireAnyPermission(users,
		"basic:vehicle:add", "basic:vehicle:edit", "basic:vehicle:record",
		"basic:device:inventory:add", "basic:device:inventory:edit",
		"basic:device:maintenance:add", "basic:device:maintenance:edit",
	), func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 21*1024*1024)
		file, e := c.FormFile("file")
		if e != nil {
			bad(c, "请选择上传文件")
			return
		}
		if file.Size > 20*1024*1024 {
			bad(c, "文件不能超过20MB")
			return
		}
		ext := strings.ToLower(filepath.Ext(file.Filename))
		allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true}
		if !allowed[ext] {
			bad(c, "仅支持图片、PDF、Word 和 Excel 文件")
			return
		}
		source, e := file.Open()
		if e != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "无法读取上传文件"})
			return
		}
		var signature [512]byte
		read, readErr := io.ReadFull(source, signature[:])
		_ = source.Close()
		if readErr != nil && readErr != io.ErrUnexpectedEOF {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "上传文件内容无效"})
			return
		}
		contentType := http.DetectContentType(signature[:read])
		if !uploadContentTypeAllowed(ext, contentType) {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "文件扩展名与实际内容不匹配"})
			return
		}
		dir := filepath.Join("uploads", "vehicles")
		if e = os.MkdirAll(dir, 0755); e != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "创建上传目录失败"})
			return
		}
		name := secureUploadName(ext)
		target := filepath.Join(dir, name)
		if e = c.SaveUploadedFile(file, target); e != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "保存上传文件失败"})
			return
		}
		c.JSON(200, gin.H{"success": true, "url": "/uploads/vehicles/" + name})
	})
	g.GET("/organizations", requireAnyPermission(users,
		"basic:organization:query", "basic:vehicle:query",
		"basic:device:inventory:query", "basic:device:maintenance:query",
		"basic:partner:finance-company:query", "basic:partner:finance-product:query",
		"basic:partner:collection-company:query",
	), func(c *gin.Context) {
		state := pagedContext(c)
		v, e := store.ListOrganizations(c.Request.Context())
		respondPagedData(c, "organizations", v, state, e)
	})
	g.POST("/organizations", requirePermission(users, "basic:organization:add"), func(c *gin.Context) {
		var v basicinfo.Organization
		if c.ShouldBindJSON(&v) != nil || v.Name == "" || v.Code == "" {
			bad(c, "组织名称和编码不能为空")
			return
		}
		if v.Status == 0 {
			v.Status = 1
		}
		v.CreatedBy = currentCreator(c, users)
		id, e := store.SaveOrganization(c, v)
		respondID(c, id, e)
	})
	g.POST("/organizations/batch", requirePermission(users, "basic:organization:import"), func(c *gin.Context) {
		var req struct {
			Organizations []basicinfo.Organization `json:"organizations"`
		}
		if c.ShouldBindJSON(&req) != nil || len(req.Organizations) == 0 {
			bad(c, "导入数据不能为空")
			return
		}
		creator := currentCreator(c, users)
		for index := range req.Organizations {
			req.Organizations[index].CreatedBy = creator
		}
		respondOK(c, store.BatchSaveOrganizations(c, req.Organizations))
	})
	g.PUT("/organizations/:id", requirePermission(users, "basic:organization:edit"), func(c *gin.Context) {
		var v basicinfo.Organization
		if c.ShouldBindJSON(&v) != nil || strings.TrimSpace(v.Name) == "" || strings.TrimSpace(v.Code) == "" {
			bad(c, "组织名称和编码不能为空")
			return
		}
		v.ID = parseID(c)
		id, e := store.SaveOrganization(c, v)
		respondID(c, id, e)
	})
	g.DELETE("/organizations/:id", requirePermission(users, "basic:organization:delete"), func(c *gin.Context) { respondOK(c, store.DeleteOrganization(c, parseID(c))) })
	g.GET("/vehicles", requirePermission(users, "basic:vehicle:query"), func(c *gin.Context) {
		state := pagedContext(c)
		v, e := store.ListVehicles(c.Request.Context())
		respondPagedData(c, "vehicles", v, state, e)
	})
	g.POST("/vehicles", requirePermission(users, "basic:vehicle:add"), func(c *gin.Context) {
		var v basicinfo.Vehicle
		if c.ShouldBindJSON(&v) != nil || strings.TrimSpace(v.PlateNo) == "" {
			bad(c, "车牌号码不能为空")
			return
		}
		if v.Status == 0 {
			v.Status = 1
		}
		v.CreatedBy = currentCreator(c, users)
		id, e := store.SaveVehicle(c, v)
		respondID(c, id, e)
	})
	g.POST("/vehicles/batch", requirePermission(users, "basic:vehicle:import"), func(c *gin.Context) {
		var req struct {
			Vehicles []basicinfo.Vehicle `json:"vehicles"`
		}
		if c.ShouldBindJSON(&req) != nil || len(req.Vehicles) == 0 {
			bad(c, "导入数据不能为空")
			return
		}
		creator := currentCreator(c, users)
		for index, v := range req.Vehicles {
			if strings.TrimSpace(v.PlateNo) == "" || v.OrganizationID == nil || *v.OrganizationID == 0 {
				bad(c, fmt.Sprintf("第%d条车辆的所属组织和车牌号码不能为空", index+1))
				return
			}
			v.CreatedBy = creator
			if _, err := store.SaveVehicle(c, v); err != nil {
				bad(c, fmt.Sprintf("第%d条车辆导入失败：%v", index+1, err))
				return
			}
		}
		c.JSON(200, gin.H{"success": true, "count": len(req.Vehicles)})
	})
	g.PUT("/vehicles/:id", requirePermission(users, "basic:vehicle:edit"), func(c *gin.Context) {
		var v basicinfo.Vehicle
		if c.ShouldBindJSON(&v) != nil {
			bad(c, "车辆数据格式错误")
			return
		}
		v.ID = parseID(c)
		id, e := store.SaveVehicle(c, v)
		respondID(c, id, e)
	})
	g.DELETE("/vehicles/:id", requirePermission(users, "basic:vehicle:delete"), func(c *gin.Context) { respondOK(c, store.DeleteVehicle(c, parseID(c))) })
	g.POST("/vehicles/bind-devices", requireAnyPermission(users, "basic:vehicle:bind", "basic:vehicle:batch-bind"), func(c *gin.Context) {
		var req struct {
			VehicleID int64                     `json:"vehicleId"`
			Bindings  []basicinfo.DeviceBinding `json:"bindings"`
			Force     bool                      `json:"force"`
		}
		if c.ShouldBindJSON(&req) != nil || req.VehicleID == 0 || len(req.Bindings) == 0 {
			bad(c, "车辆和设备不能为空")
			return
		}
		for _, binding := range req.Bindings {
			if binding.DeviceID == 0 || strings.TrimSpace(binding.InstallPosition) == "" || binding.ServiceDurationMonths <= 0 {
				bad(c, "每台设备必须选择安装位置和服务时长")
				return
			}
		}
		respondOK(c, store.BindDevices(c, req.VehicleID, req.Bindings, req.Force))
	})
	g.POST("/vehicles/unbind-devices", requireAnyPermission(users, "basic:vehicle:unbind", "basic:vehicle:batch-unbind"), func(c *gin.Context) {
		var req struct {
			DeviceIDs []int64 `json:"deviceIds"`
		}
		if c.ShouldBindJSON(&req) != nil {
			bad(c, "设备数据格式错误")
			return
		}
		respondOK(c, store.UnbindDevices(c, req.DeviceIDs))
	})
	g.POST("/vehicles/swap-device", requirePermission(users, "basic:vehicle:swap-device"), func(c *gin.Context) {
		var req struct {
			OldDeviceID int64 `json:"oldDeviceId"`
			NewDeviceID int64 `json:"newDeviceId"`
		}
		if c.ShouldBindJSON(&req) != nil || req.OldDeviceID == 0 || req.NewDeviceID == 0 {
			bad(c, "请选择需要换绑的新旧设备")
			return
		}
		respondOK(c, store.SwapDevice(c, req.OldDeviceID, req.NewDeviceID))
	})
	g.PUT("/devices/:id/installation", requirePermission(users, "basic:device:inventory:edit"), func(c *gin.Context) {
		var binding basicinfo.DeviceBinding
		if c.ShouldBindJSON(&binding) != nil || strings.TrimSpace(binding.InstallPosition) == "" {
			bad(c, "安装位置不能为空")
			return
		}
		respondOK(c, store.UpdateDeviceInstallation(c, parseID(c), binding))
	})
	g.POST("/vehicles/change-organization", requireAnyPermission(users, "basic:vehicle:change-org", "basic:vehicle:batch-change-org"), func(c *gin.Context) {
		var req struct {
			VehicleIDs     []int64 `json:"vehicleIds"`
			OrganizationID int64   `json:"organizationId"`
			IncludeDevices bool    `json:"includeDevices"`
		}
		if c.ShouldBindJSON(&req) != nil || len(req.VehicleIDs) == 0 || req.OrganizationID == 0 {
			bad(c, "车辆和目标机构不能为空")
			return
		}
		respondOK(c, store.ChangeVehicleOrganization(c, req.VehicleIDs, req.OrganizationID, req.IncludeDevices))
	})
	g.POST("/vehicles/delegate", requireAnyPermission(users, "basic:vehicle:delegate", "basic:vehicle:batch-delegate"), func(c *gin.Context) {
		var req struct {
			VehicleIDs          []int64 `json:"vehicleIds"`
			CollectionCompanyID *int64  `json:"collectionCompanyId"`
			Note                string  `json:"note"`
		}
		if c.ShouldBindJSON(&req) != nil || len(req.VehicleIDs) == 0 {
			bad(c, "请选择车辆")
			return
		}
		respondOK(c, store.DelegateVehicles(c, req.VehicleIDs, req.CollectionCompanyID, req.Note))
	})
	g.POST("/vehicles/batch-delete", requirePermission(users, "basic:vehicle:batch-delete"), func(c *gin.Context) {
		var req struct {
			VehicleIDs []int64 `json:"vehicleIds"`
			Force      bool    `json:"force"`
		}
		if c.ShouldBindJSON(&req) != nil || len(req.VehicleIDs) == 0 {
			bad(c, "请选择车辆")
			return
		}
		respondOK(c, store.BatchDeleteVehicles(c, req.VehicleIDs, req.Force))
	})
	g.GET("/vehicle-records", requirePermission(users, "basic:vehicle:query"), func(c *gin.Context) {
		state := pagedContext(c)
		v, e := store.ListVehicleRecords(c.Request.Context())
		respondPagedData(c, "records", v, state, e)
	})
	g.POST("/vehicle-records", requirePermission(users, "basic:vehicle:record"), func(c *gin.Context) {
		var v basicinfo.VehicleRecord
		if c.ShouldBindJSON(&v) != nil || v.VehicleID == 0 || v.RecordType == "" || v.RecordDate == "" {
			bad(c, "车辆、记录类型和日期不能为空")
			return
		}
		v.CreatedBy = currentCreator(c, users)
		id, e := store.SaveVehicleRecord(c, v)
		respondID(c, id, e)
	})
	g.PUT("/vehicle-records/:id", requirePermission(users, "basic:vehicle:record"), func(c *gin.Context) {
		var v basicinfo.VehicleRecord
		if c.ShouldBindJSON(&v) != nil {
			bad(c, "车辆记录格式错误")
			return
		}
		v.ID = parseID(c)
		id, e := store.SaveVehicleRecord(c, v)
		respondID(c, id, e)
	})
	g.DELETE("/vehicle-records/:id", requirePermission(users, "basic:vehicle:record"), func(c *gin.Context) { respondOK(c, store.DeleteVehicleRecord(c, parseID(c))) })

	g.GET("/devices", requireAnyPermission(users,
		"basic:device:inventory:query", "basic:device:maintenance:query",
		"basic:vehicle:query", "basic:vehicle:bind", "basic:vehicle:batch-bind",
		"basic:vehicle:swap-device",
	), func(c *gin.Context) {
		state := pagedContext(c)
		v, e := store.ListDevices(c.Request.Context())
		respondPagedData(c, "devices", v, state, e)
	})
	g.POST("/devices", requirePermission(users, "basic:device:inventory:add"), func(c *gin.Context) {
		var v basicinfo.Device
		if c.ShouldBindJSON(&v) != nil || strings.TrimSpace(v.DeviceNo) == "" {
			bad(c, "设备编号不能为空")
			return
		}
		if v.InventoryStatus == "" {
			v.InventoryStatus = "pending_production"
		}
		if deviceRequiresKey(v.Model) && strings.TrimSpace(v.DeviceKey) == "" {
			bad(c, "该设备型号必须填写设备密钥")
			return
		}
		v.CreatedBy = currentCreator(c, users)
		id, e := store.SaveDevice(c, v)
		respondID(c, id, e)
	})
	g.PUT("/devices/:id", requirePermission(users, "basic:device:inventory:edit"), func(c *gin.Context) {
		var v basicinfo.Device
		if c.ShouldBindJSON(&v) != nil {
			bad(c, "设备数据格式错误")
			return
		}
		v.ID = parseID(c)
		if deviceRequiresKey(v.Model) && strings.TrimSpace(v.DeviceKey) == "" {
			bad(c, "该设备型号必须填写设备密钥")
			return
		}
		id, e := store.SaveDevice(c, v)
		respondID(c, id, e)
	})
	g.DELETE("/devices/:id", requirePermission(users, "basic:device:inventory:delete"), func(c *gin.Context) { respondOK(c, store.DeleteDevice(c, parseID(c))) })
	g.PUT("/devices/:id/status", requirePermission(users, "basic:device:inventory:confirm"), func(c *gin.Context) {
		var req struct {
			Status string `json:"status"`
		}
		if c.ShouldBindJSON(&req) != nil || strings.TrimSpace(req.Status) == "" {
			bad(c, "目标状态不能为空")
			return
		}
		respondOK(c, store.TransitionDeviceStatus(c, parseID(c), req.Status))
	})
	g.PUT("/devices/batch-status", requirePermission(users, "basic:device:inventory:confirm"), func(c *gin.Context) {
		var req struct {
			IDs    []int64 `json:"ids"`
			Status string  `json:"status"`
		}
		if c.ShouldBindJSON(&req) != nil || len(req.IDs) == 0 || strings.TrimSpace(req.Status) == "" {
			bad(c, "请选择设备和目标状态")
			return
		}
		respondOK(c, store.BatchTransitionDeviceStatus(c, req.IDs, req.Status))
	})
	g.POST("/devices/migrate", requirePermission(users, "basic:device:inventory:migrate"), func(c *gin.Context) {
		var req struct {
			IDs            []int64 `json:"ids"`
			OrganizationID int64   `json:"organizationId"`
		}
		if c.ShouldBindJSON(&req) != nil || len(req.IDs) == 0 || req.OrganizationID == 0 {
			bad(c, "请选择设备和目标机构")
			return
		}
		respondOK(c, store.MigrateDevices(c, req.IDs, req.OrganizationID))
	})
	g.POST("/devices/renew", requirePermission(users, "basic:device:inventory:renew"), func(c *gin.Context) {
		var req struct {
			IDs            []int64 `json:"ids"`
			DurationMonths int     `json:"durationMonths"`
		}
		if c.ShouldBindJSON(&req) != nil || len(req.IDs) == 0 || req.DurationMonths <= 0 {
			bad(c, "请选择设备和续费时长")
			return
		}
		respondOK(c, store.RenewDevices(c, req.IDs, req.DurationMonths))
	})
	g.POST("/devices/batch", requirePermission(users, "basic:device:inventory:import"), func(c *gin.Context) {
		var req struct {
			Devices []basicinfo.Device `json:"devices"`
		}
		if c.ShouldBindJSON(&req) != nil || len(req.Devices) == 0 {
			bad(c, "导入数据不能为空")
			return
		}
		creator := currentCreator(c, users)
		for _, v := range req.Devices {
			if strings.TrimSpace(v.DeviceNo) == "" {
				continue
			}
			if v.InventoryStatus == "" {
				v.InventoryStatus = "pending_production"
			}
			if deviceRequiresKey(v.Model) && strings.TrimSpace(v.DeviceKey) == "" {
				bad(c, "设备 "+v.DeviceNo+" 的型号必须填写设备密钥")
				return
			}
			v.CreatedBy = creator
			if _, e := store.SaveDevice(c, v); e != nil {
				respondOK(c, e)
				return
			}
		}
		respondOK(c, nil)
	})
	g.POST("/devices/batch-delete", requirePermission(users, "basic:device:inventory:delete"), func(c *gin.Context) {
		var req struct {
			IDs []int64 `json:"ids"`
		}
		if c.ShouldBindJSON(&req) != nil || len(req.IDs) == 0 {
			bad(c, "请选择要删除的设备")
			return
		}
		for _, id := range req.IDs {
			if e := store.DeleteDevice(c, id); e != nil {
				respondOK(c, e)
				return
			}
		}
		respondOK(c, nil)
	})

	g.GET("/maintenance", requirePermission(users, "basic:device:maintenance:query"), func(c *gin.Context) {
		state := pagedContext(c)
		v, e := store.ListMaintenance(c.Request.Context())
		respondPagedData(c, "maintenance", v, state, e)
	})
	g.POST("/maintenance", requirePermission(users, "basic:device:maintenance:add"), func(c *gin.Context) {
		var v basicinfo.Maintenance
		if c.ShouldBindJSON(&v) != nil || v.DeviceID == 0 || v.MaintenanceType == "" {
			bad(c, "设备和运维类型不能为空")
			return
		}
		if v.Status == "" {
			v.Status = "pending"
		}
		v.CreatedBy = currentCreator(c, users)
		id, e := store.SaveMaintenance(c, v)
		respondID(c, id, e)
	})
	g.PUT("/maintenance/:id", requirePermission(users, "basic:device:maintenance:edit"), func(c *gin.Context) {
		var v basicinfo.Maintenance
		if c.ShouldBindJSON(&v) != nil {
			bad(c, "运维数据格式错误")
			return
		}
		v.ID = parseID(c)
		id, e := store.SaveMaintenance(c, v)
		respondID(c, id, e)
	})
	g.DELETE("/maintenance/:id", requirePermission(users, "basic:device:maintenance:delete"), func(c *gin.Context) { respondOK(c, store.DeleteMaintenance(c, parseID(c))) })
	registerPartnerCRUD(g, users, "basic:partner:finance-company", "finance-companies", "financeCompanies", func(c *gin.Context) (any, error) { return store.ListFinanceCompanies(c.Request.Context()) }, func(c *gin.Context, id int64) (int64, error) {
		var v basicinfo.FinanceCompany
		if c.ShouldBindJSON(&v) != nil {
			return 0, fmt.Errorf("金融公司数据格式错误")
		}
		v.ID = id
		if v.ID == 0 {
			v.CreatedBy = currentCreator(c, users)
		}
		return store.SaveFinanceCompany(c, v)
	}, func(c *gin.Context, id int64) error { return store.DeleteFinanceCompany(c, id) })
	registerPartnerCRUD(g, users, "basic:partner:finance-product", "finance-products", "financeProducts", func(c *gin.Context) (any, error) { return store.ListFinanceProducts(c.Request.Context()) }, func(c *gin.Context, id int64) (int64, error) {
		var v basicinfo.FinanceProduct
		if c.ShouldBindJSON(&v) != nil {
			return 0, fmt.Errorf("金融产品数据格式错误")
		}
		v.ID = id
		if v.ID == 0 {
			v.CreatedBy = currentCreator(c, users)
		}
		return store.SaveFinanceProduct(c, v)
	}, func(c *gin.Context, id int64) error { return store.DeleteFinanceProduct(c, id) })
	registerPartnerCRUD(g, users, "basic:partner:collection-company", "collection-companies", "collectionCompanies", func(c *gin.Context) (any, error) { return store.ListCollectionCompanies(c.Request.Context()) }, func(c *gin.Context, id int64) (int64, error) {
		var v basicinfo.CollectionCompany
		if c.ShouldBindJSON(&v) != nil {
			return 0, fmt.Errorf("清收公司数据格式错误")
		}
		v.ID = id
		if v.ID == 0 {
			v.CreatedBy = currentCreator(c, users)
		}
		return store.SaveCollectionCompany(c, v)
	}, func(c *gin.Context, id int64) error { return store.DeleteCollectionCompany(c, id) })
}

func secureUploadName(ext string) string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err == nil {
		return hex.EncodeToString(bytes[:]) + ext
	}
	return fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
}

func uploadContentTypeAllowed(ext, contentType string) bool {
	contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	switch ext {
	case ".jpg", ".jpeg":
		return contentType == "image/jpeg"
	case ".png":
		return contentType == "image/png"
	case ".webp":
		return contentType == "image/webp"
	case ".pdf":
		return contentType == "application/pdf"
	case ".doc", ".xls":
		return contentType == "application/x-ole-storage" || contentType == "application/octet-stream"
	case ".docx", ".xlsx":
		return contentType == "application/zip" || contentType == "application/octet-stream"
	default:
		return false
	}
}

func deviceRequiresKey(model string) bool {
	normalized := strings.ToUpper(strings.TrimSpace(model))
	for _, item := range []string{"T96", "J5", "NT08E", "NT06", "LT06", "LT12", "MT07", "MT15", "LT15", "LT07"} {
		if normalized == item {
			return true
		}
	}
	return false
}

func registerPartnerCRUD(g *gin.RouterGroup, users user.Store, permissionBase, path, key string, list func(*gin.Context) (any, error), save func(*gin.Context, int64) (int64, error), remove func(*gin.Context, int64) error) {
	g.GET("/"+path, requirePermission(users, permissionBase+":query"), func(c *gin.Context) { state := pagedContext(c); v, e := list(c); respondPagedData(c, key, v, state, e) })
	g.POST("/"+path, requirePermission(users, permissionBase+":add"), func(c *gin.Context) { id, e := save(c, 0); respondID(c, id, e) })
	g.PUT("/"+path+"/:id", requirePermission(users, permissionBase+":edit"), func(c *gin.Context) { id, e := save(c, parseID(c)); respondID(c, id, e) })
	g.DELETE("/"+path+"/:id", requirePermission(users, permissionBase+":delete"), func(c *gin.Context) { respondOK(c, remove(c, parseID(c))) })
}
