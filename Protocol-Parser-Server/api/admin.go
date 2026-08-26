package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"protocol-parser-server/auth"
	"protocol-parser-server/repository/rbac"
	"protocol-parser-server/repository/user"
)

func RegisterAdminRouter(r *gin.Engine, store rbac.Store, users user.Store, tokens *auth.Manager) {
	group := r.Group("/api/admin")
	group.Use(AuthMiddleware(tokens), requirePermissionPrefix(users, "system:"))
	group.GET("/users", requirePermission(users, "system:user:query"), func(c *gin.Context) { items, err := store.ListUsers(c); respondPagedData(c, "users", items, err) })
	group.POST("/users", requirePermission(users, "system:user:add"), func(c *gin.Context) {
		var req struct {
			rbac.User
			Password string `json:"password"`
		}
		if c.ShouldBindJSON(&req) != nil || strings.TrimSpace(req.Username) == "" || len(req.Password) < 6 {
			bad(c, "用户名不能为空，密码至少6位")
			return
		}
		req.Username = strings.TrimSpace(req.Username)
		if req.DisplayName == "" {
			req.DisplayName = req.Username
		}
		if req.Status == 0 {
			req.Status = 1
		}
		id, err := store.CreateUser(c, req.User, req.Password)
		respondID(c, id, err)
	})
	group.PUT("/users/:id", requirePermission(users, "system:user:edit"), func(c *gin.Context) {
		var req rbac.User
		if c.ShouldBindJSON(&req) != nil {
			bad(c, "用户数据格式错误")
			return
		}
		req.ID = parseID(c)
		if req.ID == currentUserID(c) && req.Status != 1 {
			bad(c, "不能停用当前登录账号")
			return
		}
		respondOK(c, store.UpdateUser(c, req))
	})
	group.PUT("/users/:id/password", requirePermission(users, "system:user:password"), func(c *gin.Context) {
		var req struct {
			Password string `json:"password"`
		}
		if c.ShouldBindJSON(&req) != nil || len(req.Password) < 6 {
			bad(c, "新密码至少6位")
			return
		}
		respondOK(c, store.ResetPassword(c, parseID(c), req.Password))
	})
	group.DELETE("/users/:id", requirePermission(users, "system:user:delete"), func(c *gin.Context) {
		id := parseID(c)
		if id == currentUserID(c) {
			bad(c, "不能删除当前登录账号")
			return
		}
		respondOK(c, store.DeleteUser(c, id))
	})
	group.GET("/roles", requirePermission(users, "system:role:query"), func(c *gin.Context) { items, err := store.ListRoles(c); respondPagedData(c, "roles", items, err) })
	group.POST("/roles", requirePermission(users, "system:role:add"), func(c *gin.Context) {
		var req rbac.Role
		if c.ShouldBindJSON(&req) != nil || req.Code == "" || req.Name == "" {
			bad(c, "角色编码和名称不能为空")
			return
		}
		if req.Status == 0 {
			req.Status = 1
		}
		id, err := store.SaveRole(c, req)
		respondID(c, id, err)
	})
	group.PUT("/roles/:id", requirePermission(users, "system:role:edit"), func(c *gin.Context) {
		var req rbac.Role
		if c.ShouldBindJSON(&req) != nil {
			bad(c, "角色数据格式错误")
			return
		}
		req.ID = parseID(c)
		id, err := store.SaveRole(c, req)
		respondID(c, id, err)
	})
	group.PUT("/roles/:id/menus", requirePermission(users, "system:role:permission"), func(c *gin.Context) {
		var req struct {
			MenuIDs []int64 `json:"menuIds"`
		}
		if c.ShouldBindJSON(&req) != nil {
			bad(c, "菜单权限格式错误")
			return
		}
		respondOK(c, store.SetRoleMenus(c, parseID(c), req.MenuIDs))
	})
	group.DELETE("/roles/:id", requirePermission(users, "system:role:delete"), func(c *gin.Context) { respondOK(c, store.DeleteRole(c, parseID(c))) })
	group.GET("/menus", requirePermission(users, "system:menu:query"), func(c *gin.Context) { items, err := store.ListMenus(c); respondPagedData(c, "menus", items, err) })
	group.POST("/menus", requirePermission(users, "system:menu:add"), func(c *gin.Context) {
		var req rbac.Menu
		if c.ShouldBindJSON(&req) != nil || req.Code == "" || req.Name == "" {
			bad(c, "菜单编码和名称不能为空")
			return
		}
		if req.MenuType == "" {
			req.MenuType = "menu"
		}
		if req.Status == 0 {
			req.Status = 1
		}
		id, err := store.SaveMenu(c, req)
		respondID(c, id, err)
	})
	group.PUT("/menus/:id", requirePermission(users, "system:menu:edit"), func(c *gin.Context) {
		var req rbac.Menu
		if c.ShouldBindJSON(&req) != nil {
			bad(c, "菜单数据格式错误")
			return
		}
		req.ID = parseID(c)
		id, err := store.SaveMenu(c, req)
		respondID(c, id, err)
	})
	group.DELETE("/menus/:id", requirePermission(users, "system:menu:delete"), func(c *gin.Context) { respondOK(c, store.DeleteMenu(c, parseID(c))) })
}

func requirePermissionPrefix(users user.Store, prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		current, err := users.GetByID(c, currUserID(c))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "error": "无权访问该功能"})
			return
		}
		for _, code := range current.Permissions {
			if code == strings.TrimSuffix(prefix, ":") || strings.HasPrefix(code, prefix) {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "error": "当前角色未分配该功能权限"})
	}
}

func requirePermission(users user.Store, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		current, err := users.GetByID(c, currUserID(c))
		if err == nil {
			for _, code := range current.Permissions {
				if code == permission {
					c.Next()
					return
				}
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "error": "缺少按钮权限：" + permission})
	}
}

func requireAnyPermission(users user.Store, permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		current, err := users.GetByID(c, currUserID(c))
		if err == nil {
			for _, owned := range current.Permissions {
				for _, required := range permissions {
					if owned == required {
						c.Next()
						return
					}
				}
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "error": "缺少对应的按钮操作权限"})
	}
}
func currUserID(c *gin.Context) int64 { return currentUserID(c) }
func parseID(c *gin.Context) int64    { id, _ := strconv.ParseInt(c.Param("id"), 10, 64); return id }
func bad(c *gin.Context, text string) {
	c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": text})
}
func respondOK(c *gin.Context, err error) {
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
func respondID(c *gin.Context, id int64, err error) {
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "id": id})
}
func respondData(c *gin.Context, key string, value interface{}, err error) {
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, key: value})
}
