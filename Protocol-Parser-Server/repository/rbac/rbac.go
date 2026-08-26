package rbac

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var ErrNotFound = errors.New("记录不存在")

type User struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Status      int    `json:"status"`
	RoleID      *int64 `json:"roleId"`
	RoleName    string `json:"roleName"`
	CreatedAt   string `json:"createdAt"`
}
type Role struct {
	ID          int64   `json:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Status      int     `json:"status"`
	IsSystem    bool    `json:"isSystem"`
	MenuIDs     []int64 `json:"menuIds"`
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
	SaveMenu(context.Context, Menu) (int64, error)
	DeleteMenu(context.Context, int64) error
}
type MySQLStore struct{ db *sql.DB }

func NewMySQLStore(db *sql.DB) *MySQLStore { return &MySQLStore{db: db} }

func (s *MySQLStore) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT u.id,u.username,u.display_name,u.email,u.phone,u.status,u.role_id,COALESCE(r.name,u.role),DATE_FORMAT(u.created_at,'%Y-%m-%d %H:%i:%s') FROM users u LEFT JOIN roles r ON r.id=u.role_id ORDER BY u.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []User{}
	for rows.Next() {
		var v User
		var rid sql.NullInt64
		if err := rows.Scan(&v.ID, &v.Username, &v.DisplayName, &v.Email, &v.Phone, &v.Status, &rid, &v.RoleName, &v.CreatedAt); err != nil {
			return nil, err
		}
		if rid.Valid {
			v.RoleID = &rid.Int64
		}
		result = append(result, v)
	}
	return result, rows.Err()
}
func (s *MySQLStore) CreateUser(ctx context.Context, v User, password string) (int64, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	role := "user"
	if v.RoleID != nil {
		_ = s.db.QueryRowContext(ctx, `SELECT code FROM roles WHERE id=?`, *v.RoleID).Scan(&role)
	}
	r, err := s.db.ExecContext(ctx, `INSERT INTO users(username,password_hash,display_name,role,role_id,email,phone,status) VALUES(?,?,?,?,?,?,?,?)`, v.Username, string(hash), v.DisplayName, role, v.RoleID, v.Email, v.Phone, v.Status)
	if err != nil {
		return 0, err
	}
	return r.LastInsertId()
}
func (s *MySQLStore) UpdateUser(ctx context.Context, v User) error {
	res, err := s.db.ExecContext(ctx, `UPDATE users u LEFT JOIN roles r ON r.id=? SET u.display_name=?,u.email=?,u.phone=?,u.status=?,u.role_id=?,u.role=COALESCE(r.code,u.role) WHERE u.id=?`, v.RoleID, v.DisplayName, v.Email, v.Phone, v.Status, v.RoleID, v.ID)
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
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `UPDATE users SET password_hash=? WHERE id=?`, string(hash), id)
	return err
}
func (s *MySQLStore) DeleteUser(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM users WHERE id=?`, id)
	return err
}

func (s *MySQLStore) ListRoles(ctx context.Context) ([]Role, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,code,name,description,status,is_system FROM roles ORDER BY is_system DESC,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Role{}
	for rows.Next() {
		var v Role
		if err := rows.Scan(&v.ID, &v.Code, &v.Name, &v.Description, &v.Status, &v.IsSystem); err != nil {
			return nil, err
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
	if v.ID == 0 {
		r, err := s.db.ExecContext(ctx, `INSERT INTO roles(code,name,description,status,is_system) VALUES(?,?,?,?,0)`, v.Code, v.Name, v.Description, v.Status)
		if err != nil {
			return 0, err
		}
		return r.LastInsertId()
	}
	_, err := s.db.ExecContext(ctx, `UPDATE roles SET name=?,description=?,status=? WHERE id=?`, v.Name, v.Description, v.Status, v.ID)
	return v.ID, err
}
func (s *MySQLStore) DeleteRole(ctx context.Context, id int64) error {
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
	rows, err := s.db.QueryContext(ctx, `SELECT id,parent_id,name,code,menu_type,path,icon,sort_order,status FROM menus ORDER BY sort_order,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Menu{}
	for rows.Next() {
		var v Menu
		if err := rows.Scan(&v.ID, &v.ParentID, &v.Name, &v.Code, &v.MenuType, &v.Path, &v.Icon, &v.SortOrder, &v.Status); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
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
		r, err := tx.ExecContext(ctx, `INSERT INTO menus(parent_id,name,code,menu_type,path,icon,sort_order,status) VALUES(?,?,?,?,?,?,?,?)`, v.ParentID, v.Name, v.Code, v.MenuType, v.Path, v.Icon, v.SortOrder, v.Status)
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
