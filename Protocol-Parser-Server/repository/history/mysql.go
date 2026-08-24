package history

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"protocol-parser-server/parser/core"
)

var beijingLocation = time.FixedZone("Asia/Shanghai", 8*60*60)

type MySQLStore struct {
	db *sql.DB
}

func NewMySQLStore(db *sql.DB) *MySQLStore {
	return &MySQLStore{db: db}
}

func (s *MySQLStore) Create(ctx context.Context, result *core.ParseResult) (int64, error) {
	payload, err := json.Marshal(result)
	if err != nil {
		return 0, fmt.Errorf("序列化解析结果失败: %w", err)
	}
	createdAt := time.Now().UTC()
	packetHash := hashPacket(result.Protocol, result.Raw)
	query := `INSERT INTO parse_history
		(protocol, message_id, message_name, packet_length, raw_hex, packet_hash, result_json, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			id = LAST_INSERT_ID(id),
			message_id = VALUES(message_id),
			message_name = VALUES(message_name),
			packet_length = VALUES(packet_length),
			result_json = VALUES(result_json),
			created_at = VALUES(created_at)`
	dbResult, err := s.db.ExecContext(ctx, query, result.Protocol, result.MessageID, result.MessageName, result.Length, result.Raw, packetHash, payload, createdAt)
	if err != nil {
		return 0, fmt.Errorf("保存解析记录失败: %w", err)
	}
	return dbResult.LastInsertId()
}

func hashPacket(protocol, raw string) string {
	canonical := strings.ToLower(strings.TrimSpace(protocol)) + ":" + strings.ToLower(strings.TrimSpace(raw))
	return fmt.Sprintf("%x", sha256.Sum256([]byte(canonical)))
}

func (s *MySQLStore) List(ctx context.Context, query Query) (*Page, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 10
	}
	keyword := strings.TrimSpace(query.Keyword)
	where, args := "", []interface{}{}
	if keyword != "" {
		where = ` WHERE protocol LIKE ? OR message_id LIKE ? OR message_name LIKE ?`
		pattern := "%" + keyword + "%"
		args = append(args, pattern, pattern, pattern)
	}
	var total int64
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM parse_history`+where, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("查询记录总数失败: %w", err)
	}
	listArgs := append(append([]interface{}{}, args...), query.PageSize, (query.Page-1)*query.PageSize)
	rows, err := s.db.QueryContext(ctx, `SELECT id, protocol, message_id, message_name, packet_length, created_at
		FROM parse_history`+where+` ORDER BY id DESC LIMIT ? OFFSET ?`, listArgs...)
	if err != nil {
		return nil, fmt.Errorf("查询解析记录失败: %w", err)
	}
	defer rows.Close()
	items := make([]Summary, 0)
	for rows.Next() {
		var item Summary
		if err := rows.Scan(&item.ID, &item.Protocol, &item.MessageID, &item.MessageName, &item.Length, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("读取解析记录失败: %w", err)
		}
		item.Time = formatTime(item.CreatedAt)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &Page{Items: items, Total: total, Page: query.Page, PageSize: query.PageSize}, nil
}

func (s *MySQLStore) Get(ctx context.Context, id int64) (*Record, error) {
	var record Record
	var payload []byte
	err := s.db.QueryRowContext(ctx, `SELECT id, protocol, message_id, message_name, packet_length, created_at, result_json
		FROM parse_history WHERE id = ?`, id).Scan(
		&record.ID, &record.Protocol, &record.MessageID, &record.MessageName, &record.Length, &record.CreatedAt, &payload,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("查询解析记录详情失败: %w", err)
	}
	var parsed core.ParseResult
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return nil, fmt.Errorf("读取解析结果JSON失败: %w", err)
	}
	record.Result = &parsed
	record.Time = formatTime(record.CreatedAt)
	return &record, nil
}

func (s *MySQLStore) Delete(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM parse_history WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("删除解析记录失败: %w", err)
	}
	count, err := result.RowsAffected()
	if err == nil && count == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *MySQLStore) Clear(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM parse_history`); err != nil {
		return fmt.Errorf("清空解析记录失败: %w", err)
	}
	return nil
}

func formatTime(value time.Time) string {
	return value.In(beijingLocation).Format("2006-01-02 15:04:05")
}
