package rbac

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"protocol-parser-server/repository/paging"
	"protocol-parser-server/repository/datascope"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var ErrNotFound = errors.New("记录不存在")

type User struct {
	ID               int64  `json:"id"`
	Username         string `json:"username"`
	DisplayName      string `json:"displayName"`
	Email            string `json:"email"`
	Phone            string `json:"phone"`
	Status           int    `json:"status"`
	RoleID           *int64 `json:"roleId"`
	RoleName         string `json:"roleName"`
	OrganizationID   *int64 `json:"organizationId"`
	OrganizationName string `json:"organizationName"`
	CreatedBy        string `json:"createdBy"`
	CreatedAt        string `json:"createdAt"`
}
type Role struct {
	ID               int64   `json:"id"`
	Code             string  `json:"code"`
	Name             string  `json:"name"`
	Description      string  `json:"description"`
	Status           int     `json:"status"`
	IsSystem         bool    `json:"isSystem"`
	OrganizationID   *int64  `json:"organizationId"`
	OrganizationName string  `json:"organizationName"`
	MenuIDs          []int64 `json:"menuIds"`
	CreatedBy        string  `json:"createdBy"`
	CreatedAt        string  `json:"createdAt"`
}
type Menu struct {
	ID        int64  `json:"id"`
	ParentID  int64  `json:"parentId"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	MenuType  string `json:"menuType"`
	Path      string `json:"path"`
	Icon      string `json:"icon"`
	SortOrder int    `json:"sortOrder"`
	Status    int    `json:"status"`
	CreatedBy string `json:"createdBy"`
	CreatedAt string `json:"createdAt"`
}

type Store interface {
	ListUsers(context.Context) ([]User, error)
	CreateUser(context.Context, User, string) (int64, error)
	UpdateUser(context.Context, User) error
	ResetPassword(context.Context, int64, string) error
	DeleteUser(context.Context, int64) error
	ListRoles(context.Context) ([]Role, error)
	SaveRole(context.Context, Role) (int64, error)
	DeleteRole(context.Context, int64) error
	SetRoleMenus(context.Context, int64, []int64) error
	ListMenus(context.Context) ([]Menu, error)
	ListMenuTree(context.Context, string, string, string) ([]Menu, error)
	// ListNavigation returns only active menu records available to the current
	// user, plus the directory ancestors required to render their hierarchy.
	// Button records are included so the client can apply their enabled state,
	// but it is the caller's responsibility not to render them as navigation.
	ListNavigation(context.Context, []string) ([]Menu, error)
	SaveMenu(context.Context, Menu) (int64, error)
	DeleteMenu(context.Context, int64) error
}
type MySQLStore struct{ db *sql.DB }

func NewMySQLStore(db *sql.DB) *MySQLStore { return &MySQLStore{db: db} }

func appendScope(ctx context.Context, clauses *[]string, args *[]any, column string) {
	if clause, values := datascope.Clause(ctx, column); clause != "" {
		*clauses = append(*clauses, clause)
		*args = append(*args, values...)
	}
}

func requireOrganizationScope(ctx context.Context, organizationID *int64) error {
	if datascope.From(ctx).All { return nil }
	if organizationID == nil || !datascope.Allows(ctx, *organizationID) { return errors.New("无权操作该机构的数据") }
	return nil
}

func (s *MySQLStore) requireUserScope(ctx context.Context, id int64) error {
	var organizationID sql.NullInt64
	if err := s.db.QueryRowContext(ctx, `SELECT organization_id FROM users WHERE id=?`, id).Scan(&organizationID); err != nil { return err }
	if !organizationID.Valid { return requireOrganizationScope(ctx, nil) }
	return requireOrganizationScope(ctx, &organizationID.Int64)
}

func (s *MySQLStore) requireRoleScope(ctx context.Context, id int64) error {
	var organizationID sql.NullInt64
	if err := s.db.QueryRowContext(ctx, `SELECT organization_id FROM roles WHERE id=?`, id).Scan(&organizationID); err != nil { return err }
	if !organizationID.Valid { return requireOrganizationScope(ctx, nil) }
	return requireOrganizationScope(ctx, &organizationID.Int64)
}

func (s *MySQLStore) ListUsers(ctx context.Context) ([]User, error) {
	q := paging.FromContext(ctx).Query
	clauses, args := []string{}, []any{}
	if q.Keyword != "" {
		clauses = append(clauses, `(u.username LIKE ? OR u.display_name LIKE ? OR u.email LIKE ? OR u.phone LIKE ?)`)
		term := "%" + q.Keyword + "%"
		args = append(args, term, term, term, term)
	}
	if q.Status != "" {
		clauses = append(clauses, `u.status=?`)
		args = append(args, q.Status)
	}
	if roleID, ok := paging.Int64(q.RoleID); ok {
		clauses = append(clauses, `u.role_id=?`)
		args = append(args, roleID)
	}
	appendScope(ctx, &clauses, &args, "u.organization_id")
	rows, err := paging.QueryRows(ctx, s.db,
		`SELECT u.id,u.username,u.display_name,u.email,u.phone,u.status,u.role_id,COALESCE(r.name,u.role),u.organization_id,COALESCE(o.name,''),u.created_by,DATE_FORMAT(u.created_at,'%Y-%m-%d %H:%i:%s') FROM users u LEFT JOIN roles r ON r.id=u.role_id LEFT JOIN organizations o ON o.id=u.organization_id`,
		`SELECT COUNT(*) FROM users u LEFT JOIN roles r ON r.id=u.role_id LEFT JOIN organizations o ON o.id=u.organization_id`, clauses, args,
		q.OrderBy(map[string]string{"createdAt": "u.created_at", "username": "u.username", "status": "u.status"}, "u.id DESC"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []User{}
	for rows.Next() {
		var v User
		var rid, oid sql.NullInt64
		if err := rows.Scan(&v.ID, &v.Username, &v.DisplayName, &v.Email, &v.Phone, &v.Status, &rid, &v.RoleName, &oid, &v.OrganizationName, &v.CreatedBy, &v.CreatedAt); err != nil {
			return nil, err
		}
		if rid.Valid {
			v.RoleID = &rid.Int64
		}
		if oid.Valid {
			v.OrganizationID = &oid.Int64
		}
		result = append(result, v)
	}
	return result, rows.Err()
}
func (s *MySQLStore) CreateUser(ctx context.Context, v User, password string) (int64, error) {
	if err := requireOrganizationScope(ctx, v.OrganizationID); err != nil { return 0, err }
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	role := "user"
	if v.RoleID != nil {
		_ = s.db.QueryRowContext(ctx, `SELECT code FROM roles WHERE id=?`, *v.RoleID).Scan(&role)
	}
	r, err := s.db.ExecContext(ctx, `INSERT INTO users(username,password_hash,display_name,role,role_id,organization_id,email,phone,status,created_by) VALUES(?,?,?,?,?,?,?,?,?,?)`, v.Username, string(hash), v.DisplayName, role, v.RoleID, v.OrganizationID, v.Email, v.Phone, v.Status, v.CreatedBy)
	if err != nil {
		return 0, err
	}
	return r.LastInsertId()
}
func (s *MySQLStore) UpdateUser(ctx context.Context, v User) error {
	if err := s.requireUserScope(ctx, v.ID); err != nil { return err }
	if err := requireOrganizationScope(ctx, v.OrganizationID); err != nil { return err }
	res, err := s.db.ExecContext(ctx, `UPDATE users u LEFT JOIN roles r ON r.id=? SET u.display_name=?,u.email=?,u.phone=?,u.status=?,u.role_id=?,u.organization_id=?,u.role=COALESCE(r.code,u.role) WHERE u.id=?`, v.RoleID, v.DisplayName, v.Email, v.Phone, v.Status, v.RoleID, v.OrganizationID, v.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
func (s *MySQLStore) ResetPassword(ctx context.Context, id int64, password string) error {
	if err := s.requireUserScope(ctx, id); err != nil { return err }
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `UPDATE users SET password_hash=? WHERE id=?`, string(hash), id)
	return err
}
func (s *MySQLStore) DeleteUser(ctx context.Context, id int64) error {
	if err := s.requireUserScope(ctx, id); err != nil { return err }
	_, err := s.db.ExecContext(ctx, `DELETE FROM users WHERE id=?`, id)
	return err
}

func (s *MySQLStore) ListRoles(ctx context.Context) ([]Role, error) {
	q := paging.FromContext(ctx).Query
	clauses, args := []string{}, []any{}
	if q.Keyword != "" {
		clauses = append(clauses, `(code LIKE ? OR name LIKE ? OR description LIKE ?)`)
		term := "%" + q.Keyword + "%"
		args = append(args, term, term, term)
	}
	if q.Status != "" {
		clauses = append(clauses, `status=?`)
		args = append(args, q.Status)
	}
	appendScope(ctx, &clauses, &args, "r.organization_id")
	rows, err := paging.QueryRows(ctx, s.db,
		`SELECT r.id,r.code,r.name,r.description,r.status,r.is_system,r.organization_id,COALESCE(o.name,''),r.created_by,DATE_FORMAT(r.created_at,'%Y-%m-%d %H:%i:%s') FROM roles r LEFT JOIN organizations o ON o.id=r.organization_id`, `SELECT COUNT(*) FROM roles r LEFT JOIN organizations o ON o.id=r.organization_id`, clauses, args,
		q.OrderBy(map[string]string{"name": "name", "status": "status", "createdAt": "created_at"}, "is_system DESC,id"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Role{}
	for rows.Next() {
		var v Role
		var oid sql.NullInt64
		if err := rows.Scan(&v.ID, &v.Code, &v.Name, &v.Description, &v.Status, &v.IsSystem, &oid, &v.OrganizationName, &v.CreatedBy, &v.CreatedAt); err != nil {
			return nil, err
		}
		if oid.Valid {
			v.OrganizationID = &oid.Int64
		}
		mrows, err := s.db.QueryContext(ctx, `SELECT menu_id FROM role_menus WHERE role_id=?`, v.ID)
		if err != nil {
			return nil, err
		}
		for mrows.Next() {
			var id int64
			_ = mrows.Scan(&id)
			v.MenuIDs = append(v.MenuIDs, id)
		}
		mrows.Close()
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *MySQLStore) SaveRole(ctx context.Context, v Role) (int64, error) {
	if err := requireOrganizationScope(ctx, v.OrganizationID); err != nil { return 0, err }
	if v.ID == 0 {
		r, err := s.db.ExecContext(ctx, `INSERT INTO roles(code,name,description,status,is_system,organization_id,created_by) VALUES(?,?,?,?,0,?,?)`, v.Code, v.Name, v.Description, v.Status, v.OrganizationID, v.CreatedBy)
		if err != nil {
			return 0, err
		}
		return r.LastInsertId()
	}
	if err := s.requireRoleScope(ctx, v.ID); err != nil { return 0, err }
	_, err := s.db.ExecContext(ctx, `UPDATE roles SET name=?,description=?,status=?,organization_id=? WHERE id=?`, v.Name, v.Description, v.Status, v.OrganizationID, v.ID)
	return v.ID, err
}
func (s *MySQLStore) DeleteRole(ctx context.Context, id int64) error {
	if err := s.requireRoleScope(ctx, id); err != nil { return err }
	var system bool
	if err := s.db.QueryRowContext(ctx, `SELECT is_system FROM roles WHERE id=?`, id).Scan(&system); err != nil {
		return err
	}
	if system {
		return errors.New("系统内置角色不能删除")
	}
	var count int
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE role_id=?`, id).Scan(&count)
	if count > 0 {
		return errors.New("该角色仍有关联用户")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `DELETE FROM role_menus WHERE role_id=?`, id); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM roles WHERE id=?`, id); err != nil {
		return err
	}
	return tx.Commit()
}
func (s *MySQLStore) SetRoleMenus(ctx context.Context, id int64, menuIDs []int64) error {
	if err := s.requireRoleScope(ctx, id); err != nil { return err }
	var isSystem bool
	var roleCode string
	if err := s.db.QueryRowContext(ctx, `SELECT is_system,code FROM roles WHERE id=?`, id).Scan(&isSystem, &roleCode); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `DELETE FROM role_menus WHERE role_id=?`, id); err != nil {
		return err
	}
	if isSystem && roleCode == "admin" {
		if _, err = tx.ExecContext(ctx, `INSERT INTO role_menus(role_id,menu_id) SELECT ?,id FROM menus`, id); err != nil {
			return err
		}
		return tx.Commit()
	}
	for _, mid := range menuIDs {
		if _, err = tx.ExecContext(ctx, `INSERT INTO role_menus(role_id,menu_id) VALUES(?,?)`, id, mid); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *MySQLStore) ListMenus(ctx context.Context) ([]Menu, error) {
	q := paging.FromContext(ctx).Query
	clauses, args := []string{}, []any{}
	if q.Keyword != "" {
		clauses = append(clauses, `(name LIKE ? OR code LIKE ? OR path LIKE ?)`)
		term := "%" + q.Keyword + "%"
		args = append(args, term, term, term)
	}
	if q.Status != "" {
		clauses = append(clauses, `status=?`)
		args = append(args, q.Status)
	}
	if q.MenuType != "" {
		clauses = append(clauses, `menu_type=?`)
		args = append(args, q.MenuType)
	}
	rows, err := paging.QueryRows(ctx, s.db,
		`SELECT id,parent_id,name,code,menu_type,path,icon,sort_order,status,created_by,DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s') FROM menus`, `SELECT COUNT(*) FROM menus`, clauses, args,
		q.OrderBy(map[string]string{"sortOrder": "sort_order", "name": "name", "status": "status", "createdAt": "created_at"}, "sort_order,id"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Menu{}
	for rows.Next() {
		var v Menu
		if err := rows.Scan(&v.ID, &v.ParentID, &v.Name, &v.Code, &v.MenuType, &v.Path, &v.Icon, &v.SortOrder, &v.Status, &v.CreatedBy, &v.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// ListMenuTree deliberately does not use the normal list-page limit. A menu
// hierarchy must be read as a complete set; paginating it can split a parent
// from its children and silently drop permissions from the authorization tree.
func (s *MySQLStore) ListMenuTree(ctx context.Context, keyword, status, menuType string) ([]Menu, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,parent_id,name,code,menu_type,path,icon,sort_order,status FROM menus ORDER BY sort_order,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	all := []Menu{}
	byID := map[int64]Menu{}
	for rows.Next() {
		var item Menu
		if err := rows.Scan(&item.ID, &item.ParentID, &item.Name, &item.Code, &item.MenuType, &item.Path, &item.Icon, &item.SortOrder, &item.Status); err != nil {
			return nil, err
		}
		all = append(all, item)
		byID[item.ID] = item
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	needle := strings.ToLower(strings.TrimSpace(keyword))
	matched := map[int64]bool{}
	for _, item := range all {
		matchesKeyword := needle == "" || strings.Contains(strings.ToLower(item.Name), needle) || strings.Contains(strings.ToLower(item.Code), needle) || strings.Contains(strings.ToLower(item.Path), needle)
		matchesStatus := status == "" || status == fmt.Sprint(item.Status)
		matchesType := menuType == "" || menuType == item.MenuType
		if matchesKeyword && matchesStatus && matchesType {
			matched[item.ID] = true
		}
	}
	// Keep ancestors of a matching node. Without this, a query for a button
	// would return an orphan and Ant Table could not render the tree context.
	for id := range matched {
		visited := map[int64]bool{id: true}
		for parentID := byID[id].ParentID; parentID != 0; {
			if visited[parentID] {
				break
			}
			visited[parentID] = true
			parent, exists := byID[parentID]
			if !exists {
				break
			}
			matched[parentID] = true
			parentID = parent.ParentID
		}
	}
	items := make([]Menu, 0, len(matched))
	for _, item := range all {
		if matched[item.ID] {
			items = append(items, item)
		}
	}
	return items, nil
}

func (s *MySQLStore) ListNavigation(ctx context.Context, permissions []string) ([]Menu, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,parent_id,name,code,menu_type,path,icon,sort_order,status FROM menus WHERE status=1 ORDER BY sort_order,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	all := make([]Menu, 0)
	byID := make(map[int64]Menu)
	for rows.Next() {
		var item Menu
		if err := rows.Scan(&item.ID, &item.ParentID, &item.Name, &item.Code, &item.MenuType, &item.Path, &item.Icon, &item.SortOrder, &item.Status); err != nil {
			return nil, err
		}
		all = append(all, item)
		byID[item.ID] = item
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	allowed := make(map[int64]bool)
	for _, item := range all {
		for _, permission := range permissions {
			if permission == item.Code || (item.Code != "" && len(permission) > len(item.Code) && permission[:len(item.Code)+1] == item.Code+":") {
				allowed[item.ID] = true
				break
			}
		}
	}
	for id := range allowed {
		visited := map[int64]bool{id: true}
		for parentID := byID[id].ParentID; parentID != 0; {
			if visited[parentID] {
				break
			}
			visited[parentID] = true
			parent, exists := byID[parentID]
			if !exists {
				break
			}
			allowed[parentID] = true
			parentID = parent.ParentID
		}
	}

	result := make([]Menu, 0, len(allowed))
	for _, item := range all {
		if allowed[item.ID] {
			result = append(result, item)
		}
	}
	return result, nil
}
func (s *MySQLStore) SaveMenu(ctx context.Context, v Menu) (int64, error) {
	if v.ID == v.ParentID && v.ID != 0 {
		return 0, errors.New("上级菜单不能选择自身")
	}
	if v.ID == 0 {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return 0, err
		}
		defer tx.Rollback()
		r, err := tx.ExecContext(ctx, `INSERT INTO menus(parent_id,name,code,menu_type,path,icon,sort_order,status,created_by) VALUES(?,?,?,?,?,?,?,?,?)`, v.ParentID, v.Name, v.Code, v.MenuType, v.Path, v.Icon, v.SortOrder, v.Status, v.CreatedBy)
		if err != nil {
			return 0, err
		}
		menuID, err := r.LastInsertId()
		if err != nil {
			return 0, err
		}
		if _, err = tx.ExecContext(ctx, `INSERT IGNORE INTO role_menus(role_id,menu_id) SELECT id,? FROM roles WHERE code='admin' AND is_system=1`, menuID); err != nil {
			return 0, err
		}
		if err = tx.Commit(); err != nil {
			return 0, err
		}
		return menuID, nil
	}
	_, err := s.db.ExecContext(ctx, `UPDATE menus SET parent_id=?,name=?,menu_type=?,path=?,icon=?,sort_order=?,status=? WHERE id=?`, v.ParentID, v.Name, v.MenuType, v.Path, v.Icon, v.SortOrder, v.Status, v.ID)
	return v.ID, err
}
func (s *MySQLStore) DeleteMenu(ctx context.Context, id int64) error {
	var count int
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM menus WHERE parent_id=?`, id).Scan(&count)
	if count > 0 {
		return errors.New("请先删除下级菜单")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, _ = tx.ExecContext(ctx, `DELETE FROM role_menus WHERE menu_id=?`, id)
	if _, err = tx.ExecContext(ctx, `DELETE FROM menus WHERE id=?`, id); err != nil {
		return fmt.Errorf("删除菜单失败: %w", err)
	}
	return tx.Commit()
}
