package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/go-sql-driver/mysql"
)

var databaseNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func OpenMySQLFromEnv(ctx context.Context) (*sql.DB, error) {
	host := envOr("MYSQL_HOST", "127.0.0.1")
	port := envOr("MYSQL_PORT", "3306")
	user := envOr("MYSQL_USER", "root")
	password := os.Getenv("MYSQL_PASSWORD")
	databaseName := envOr("MYSQL_DATABASE", "protocol_parser")
	if !databaseNamePattern.MatchString(databaseName) {
		return nil, fmt.Errorf("MYSQL_DATABASE只能包含字母、数字和下划线")
	}
	config := mysql.NewConfig()
	config.User = user
	config.Passwd = password
	config.Net = "tcp"
	config.Addr = host + ":" + port
	config.ParseTime = true
	config.Loc = time.UTC
	config.Params = map[string]string{"charset": "utf8mb4", "collation": "utf8mb4_unicode_ci"}

	adminDB, err := sql.Open("mysql", config.FormatDSN())
	if err != nil {
		return nil, err
	}
	defer adminDB.Close()
	if err := adminDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("连接MySQL失败(%s:%s): %w", host, port, err)
	}
	if _, err := adminDB.ExecContext(ctx, `CREATE DATABASE IF NOT EXISTS `+"`"+databaseName+"`"+` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci`); err != nil {
		return nil, fmt.Errorf("创建数据库%s失败: %w", databaseName, err)
	}

	config.DBName = databaseName
	db, err := sql.Open("mysql", config.FormatDSN())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(envIntOr("MYSQL_MAX_OPEN_CONNS", 10))
	db.SetMaxIdleConns(envIntOr("MYSQL_MAX_IDLE_CONNS", 5))
	db.SetConnMaxLifetime(30 * time.Minute)
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("连接数据库%s失败: %w", databaseName, err)
	}
	if err := migrate(ctx, db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func migrate(ctx context.Context, db *sql.DB) error {
	const schema = `CREATE TABLE IF NOT EXISTS parse_history (
		id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
		protocol VARCHAR(32) NOT NULL,
		message_id VARCHAR(32) NOT NULL,
		message_name VARCHAR(128) NOT NULL,
		packet_length INT UNSIGNED NOT NULL,
		raw_hex LONGTEXT NOT NULL,
		packet_hash CHAR(64) NOT NULL,
		result_json JSON NOT NULL,
		created_at DATETIME(3) NOT NULL,
		PRIMARY KEY (id),
		INDEX idx_parse_history_created_at (created_at),
		INDEX idx_parse_history_protocol (protocol),
		INDEX idx_parse_history_message_id (message_id),
		INDEX idx_parse_history_message_name (message_name),
		UNIQUE INDEX uk_parse_history_packet_hash (packet_hash)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("创建解析历史表失败: %w", err)
	}
	if err := migrateHistoryPacketHash(ctx, db); err != nil {
		return err
	}
	return nil
}

func migrateHistoryPacketHash(ctx context.Context, db *sql.DB) error {
	var columnCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = DATABASE() AND table_name = 'parse_history' AND column_name = 'packet_hash'`).Scan(&columnCount); err != nil {
		return fmt.Errorf("检查解析记录去重字段失败: %w", err)
	}
	if columnCount == 0 {
		if _, err := db.ExecContext(ctx, `ALTER TABLE parse_history ADD COLUMN packet_hash CHAR(64) NULL AFTER raw_hex`); err != nil {
			return fmt.Errorf("添加解析记录去重字段失败: %w", err)
		}
	}
	if _, err := db.ExecContext(ctx, `UPDATE parse_history
		SET packet_hash = SHA2(CONCAT(LOWER(TRIM(protocol)), ':', LOWER(TRIM(raw_hex))), 256)
		WHERE packet_hash IS NULL OR packet_hash = ''`); err != nil {
		return fmt.Errorf("生成历史报文指纹失败: %w", err)
	}
	if _, err := db.ExecContext(ctx, `DELETE duplicate_record FROM parse_history duplicate_record
		INNER JOIN parse_history retained_record
			ON duplicate_record.packet_hash = retained_record.packet_hash
			AND duplicate_record.id < retained_record.id`); err != nil {
		return fmt.Errorf("清理重复解析记录失败: %w", err)
	}

	var indexCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.statistics
		WHERE table_schema = DATABASE() AND table_name = 'parse_history' AND index_name = 'uk_parse_history_packet_hash'`).Scan(&indexCount); err != nil {
		return fmt.Errorf("检查解析记录唯一索引失败: %w", err)
	}
	if indexCount == 0 {
		if _, err := db.ExecContext(ctx, `ALTER TABLE parse_history
			MODIFY COLUMN packet_hash CHAR(64) NOT NULL,
			ADD UNIQUE INDEX uk_parse_history_packet_hash (packet_hash)`); err != nil {
			return fmt.Errorf("创建解析记录唯一索引失败: %w", err)
		}
	}
	return nil
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envIntOr(key string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
