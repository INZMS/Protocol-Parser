package user

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("用户名或密码错误")

type User struct {
	ID          int64      `json:"id"`
	Username    string     `json:"username"`
	DisplayName string     `json:"displayName"`
	Role        string     `json:"role"`
	Email       string     `json:"email"`
	Phone       string     `json:"phone"`
	Status      int        `json:"status"`
	LastLoginAt *time.Time `json:"lastLoginAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	Permissions []string   `json:"permissions"`
}

type Store interface {
	Authenticate(context.Context, string, string) (*User, error)
	GetByID(context.Context, int64) (*User, error)
	UpdateProfile(context.Context, int64, string, string, string) (*User, error)
}

type MySQLStore struct{ db *sql.DB }

func NewMySQLStore(db *sql.DB) *MySQLStore { return &MySQLStore{db: db} }

func (store *MySQLStore) Authenticate(ctx context.Context, username, password string) (*User, error) {
	var id int64
	var hash string
	var status int
	err := store.db.QueryRowContext(ctx, `SELECT id, password_hash, status FROM users WHERE username = ? LIMIT 1`, username).Scan(&id, &hash, &status)
	if errors.Is(err, sql.ErrNoRows) || err == nil && (status != 1 || bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	if _, err := store.db.ExecContext(ctx, `UPDATE users SET last_login_at = UTC_TIMESTAMP(3) WHERE id = ?`, id); err != nil {
		return nil, err
	}
	return store.GetByID(ctx, id)
}

func (store *MySQLStore) GetByID(ctx context.Context, id int64) (*User, error) {
	var result User
	var roleID int64
	err := store.db.QueryRowContext(ctx, `SELECT u.id,u.username,u.display_name,r.code,u.email,u.phone,u.status,u.last_login_at,u.created_at,u.role_id
		FROM users u JOIN roles r ON r.id=u.role_id AND r.status=1 WHERE u.id=? AND u.status=1`, id).
		Scan(&result.ID, &result.Username, &result.DisplayName, &result.Role, &result.Email, &result.Phone, &result.Status, &result.LastLoginAt, &result.CreatedAt, &roleID)
	if err != nil {
		return nil, err
	}
	type menuNode struct {
		parentID int64
		code     string
		status   int
	}
	menuRows, err := store.db.QueryContext(ctx, `SELECT id,parent_id,code,status FROM menus`)
	if err != nil {
		return nil, err
	}
	nodes := map[int64]menuNode{}
	for menuRows.Next() {
		var menuID int64
		var node menuNode
		if err = menuRows.Scan(&menuID, &node.parentID, &node.code, &node.status); err != nil {
			menuRows.Close()
			return nil, err
		}
		nodes[menuID] = node
	}
	if err = menuRows.Close(); err != nil {
		return nil, err
	}
	rows, err := store.db.QueryContext(ctx, `SELECT menu_id FROM role_menus WHERE role_id=?`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result.Permissions = []string{}
	for rows.Next() {
		var menuID int64
		if err = rows.Scan(&menuID); err != nil {
			return nil, err
		}
		node, exists := nodes[menuID]
		if !exists || node.status != 1 {
			continue
		}
		active := true
		visited := map[int64]bool{menuID: true}
		for parentID := node.parentID; parentID != 0; {
			if visited[parentID] {
				active = false
				break
			}
			visited[parentID] = true
			parent, ok := nodes[parentID]
			if !ok || parent.status != 1 {
				active = false
				break
			}
			parentID = parent.parentID
		}
		if active {
			result.Permissions = append(result.Permissions, node.code)
		}
	}
	return &result, rows.Err()
}

func (store *MySQLStore) UpdateProfile(ctx context.Context, id int64, displayName, email, phone string) (*User, error) {
	if _, err := store.db.ExecContext(ctx, `UPDATE users SET display_name = ?, email = ?, phone = ?, updated_at = UTC_TIMESTAMP(3) WHERE id = ? AND status = 1`, displayName, email, phone, id); err != nil {
		return nil, err
	}
	return store.GetByID(ctx, id)
}
