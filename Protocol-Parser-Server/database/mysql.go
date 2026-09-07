package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
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
	if err := migrateUsers(ctx, db); err != nil {
		return err
	}
	if err := migrateSystemSettings(ctx, db); err != nil {
		return err
	}
	if err := migrateRBAC(ctx, db); err != nil {
		return err
	}
	if err := migrateBasicInfo(ctx, db); err != nil {
		return err
	}
	return nil
}

func migrateBasicInfo(ctx context.Context, db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS organizations (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, parent_id BIGINT UNSIGNED NOT NULL DEFAULT 0, name VARCHAR(128) NOT NULL,
			code VARCHAR(64) NOT NULL, status TINYINT UNSIGNED NOT NULL DEFAULT 1, created_by VARCHAR(64) NOT NULL DEFAULT '系统管理员', created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
			PRIMARY KEY(id), UNIQUE INDEX uk_organizations_code(code), INDEX idx_organizations_parent(parent_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS vehicles (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, plate_no VARCHAR(32) NOT NULL, plate_color VARCHAR(16) NOT NULL DEFAULT '蓝色',
			vin VARCHAR(32) NOT NULL DEFAULT '', owner_name VARCHAR(64) NOT NULL DEFAULT '', owner_phone VARCHAR(32) NOT NULL DEFAULT '',
			organization_id BIGINT UNSIGNED NULL, vehicle_type VARCHAR(32) NOT NULL DEFAULT '', brand_model VARCHAR(64) NOT NULL DEFAULT '', engine_no VARCHAR(64) NOT NULL DEFAULT '',
			registration_date DATE NULL, use_nature VARCHAR(32) NOT NULL DEFAULT '', insurance_expiry DATE NULL, inspection_expiry DATE NULL, mileage DECIMAL(12,1) NOT NULL DEFAULT 0,
			driver_name VARCHAR(64) NOT NULL DEFAULT '', driver_phone VARCHAR(32) NOT NULL DEFAULT '', driver_id_no VARCHAR(32) NOT NULL DEFAULT '', driver_license_no VARCHAR(32) NOT NULL DEFAULT '',
			driver_license_class VARCHAR(16) NOT NULL DEFAULT '', driver_license_issue_date DATE NULL, driver_license_expiry DATE NULL,
			driving_license_no VARCHAR(64) NOT NULL DEFAULT '', registration_authority VARCHAR(128) NOT NULL DEFAULT '', approved_load DECIMAL(12,1) NOT NULL DEFAULT 0,
			overall_dimensions VARCHAR(64) NOT NULL DEFAULT '', fuel_type VARCHAR(32) NOT NULL DEFAULT '', emission_standard VARCHAR(32) NOT NULL DEFAULT '', driving_license_issue_date DATE NULL,
			photo_vehicle_front VARCHAR(255) NOT NULL DEFAULT '', photo_vehicle_rear VARCHAR(255) NOT NULL DEFAULT '', photo_driving_license_front VARCHAR(255) NOT NULL DEFAULT '',
			photo_driving_license_back VARCHAR(255) NOT NULL DEFAULT '', photo_driver_license_front VARCHAR(255) NOT NULL DEFAULT '', photo_driver_license_back VARCHAR(255) NOT NULL DEFAULT '',
			delegated_org_id BIGINT UNSIGNED NULL, delegation_status VARCHAR(16) NOT NULL DEFAULT 'none', delegated_at DATETIME NULL, delegation_note VARCHAR(255) NOT NULL DEFAULT '',
			status TINYINT UNSIGNED NOT NULL DEFAULT 1, remark VARCHAR(255) NOT NULL DEFAULT '', created_by VARCHAR(64) NOT NULL DEFAULT '系统管理员', created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
			updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3), PRIMARY KEY(id), UNIQUE INDEX uk_vehicles_plate(plate_no), INDEX idx_vehicles_vin(vin), INDEX idx_vehicles_org(organization_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS vehicle_records (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, vehicle_id BIGINT UNSIGNED NOT NULL, record_type VARCHAR(16) NOT NULL,
			record_date DATE NOT NULL, title VARCHAR(128) NOT NULL DEFAULT '', amount DECIMAL(12,2) NOT NULL DEFAULT 0, mileage DECIMAL(12,1) NOT NULL DEFAULT 0,
			status VARCHAR(32) NOT NULL DEFAULT '', detail VARCHAR(500) NOT NULL DEFAULT '', attachment_url VARCHAR(255) NOT NULL DEFAULT '', created_by VARCHAR(64) NOT NULL DEFAULT '系统管理员', created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
			PRIMARY KEY(id), INDEX idx_vehicle_records_vehicle(vehicle_id), INDEX idx_vehicle_records_type(record_type), INDEX idx_vehicle_records_date(record_date)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS devices (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, device_no VARCHAR(64) NOT NULL, imei VARCHAR(32) NOT NULL DEFAULT '', model VARCHAR(64) NOT NULL DEFAULT '',
			protocol VARCHAR(32) NOT NULL DEFAULT '', sim_no VARCHAR(32) NOT NULL DEFAULT '', iccid VARCHAR(32) NOT NULL DEFAULT '', organization_id BIGINT UNSIGNED NULL, vehicle_id BIGINT UNSIGNED NULL,
			inventory_status VARCHAR(32) NOT NULL DEFAULT 'pending_production', inbound_date DATE NULL, remark VARCHAR(255) NOT NULL DEFAULT '',
			created_by VARCHAR(64) NOT NULL DEFAULT '系统管理员', created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3), updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
			PRIMARY KEY(id), UNIQUE INDEX uk_devices_no(device_no), INDEX idx_devices_imei(imei), INDEX idx_devices_status(inventory_status), INDEX idx_devices_vehicle(vehicle_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS device_maintenance (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, device_id BIGINT UNSIGNED NOT NULL, maintenance_type VARCHAR(32) NOT NULL,
			issue_description VARCHAR(500) NOT NULL DEFAULT '', handling_result VARCHAR(500) NOT NULL DEFAULT '', handler VARCHAR(64) NOT NULL DEFAULT '',
			status VARCHAR(16) NOT NULL DEFAULT 'pending', handled_at DATETIME NULL, created_by VARCHAR(64) NOT NULL DEFAULT '系统管理员', created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
			updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3), PRIMARY KEY(id), INDEX idx_maintenance_device(device_id), INDEX idx_maintenance_status(status)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS system_notifications (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, organization_id BIGINT UNSIGNED NULL, device_id BIGINT UNSIGNED NULL,
			type VARCHAR(32) NOT NULL, title VARCHAR(128) NOT NULL, content VARCHAR(500) NOT NULL, notification_key VARCHAR(128) NOT NULL,
			is_read TINYINT UNSIGNED NOT NULL DEFAULT 0, created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
			PRIMARY KEY(id), UNIQUE INDEX uk_notification_key(notification_key), INDEX idx_notifications_org(organization_id), INDEX idx_notifications_read(is_read)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS finance_companies (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, name VARCHAR(128) NOT NULL, code VARCHAR(64) NOT NULL,
			organization_id BIGINT UNSIGNED NULL,
			contact_name VARCHAR(64) NOT NULL DEFAULT '', contact_phone VARCHAR(32) NOT NULL DEFAULT '', address VARCHAR(255) NOT NULL DEFAULT '',
			status TINYINT UNSIGNED NOT NULL DEFAULT 1, remark VARCHAR(500) NOT NULL DEFAULT '', created_by VARCHAR(64) NOT NULL DEFAULT '系统管理员', created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
			updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3), PRIMARY KEY(id), UNIQUE INDEX uk_finance_companies_code(code)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS finance_products (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, company_id BIGINT UNSIGNED NOT NULL, name VARCHAR(128) NOT NULL, code VARCHAR(64) NOT NULL,
			organization_id BIGINT UNSIGNED NULL,
			product_type VARCHAR(64) NOT NULL DEFAULT '', annual_rate DECIMAL(8,4) NOT NULL DEFAULT 0, term_months INT UNSIGNED NOT NULL DEFAULT 0,
			status TINYINT UNSIGNED NOT NULL DEFAULT 1, remark VARCHAR(500) NOT NULL DEFAULT '', created_by VARCHAR(64) NOT NULL DEFAULT '系统管理员', created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
			updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3), PRIMARY KEY(id), UNIQUE INDEX uk_finance_products_code(code), INDEX idx_finance_products_company(company_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS collection_companies (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, name VARCHAR(128) NOT NULL, code VARCHAR(64) NOT NULL,
			organization_id BIGINT UNSIGNED NULL,
			contact_name VARCHAR(64) NOT NULL DEFAULT '', contact_phone VARCHAR(32) NOT NULL DEFAULT '', service_area VARCHAR(255) NOT NULL DEFAULT '', address VARCHAR(255) NOT NULL DEFAULT '',
			status TINYINT UNSIGNED NOT NULL DEFAULT 1, remark VARCHAR(500) NOT NULL DEFAULT '', created_by VARCHAR(64) NOT NULL DEFAULT '系统管理员', created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
			updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3), PRIMARY KEY(id), UNIQUE INDEX uk_collection_companies_code(code)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	}
	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("创建基础信息数据表失败: %w", err)
		}
	}
	columns := []struct{ table, name, definition string }{
		{"users", "organization_id", "BIGINT UNSIGNED NULL AFTER role_id"}, {"roles", "organization_id", "BIGINT UNSIGNED NULL AFTER is_system"},
		{"finance_companies", "organization_id", "BIGINT UNSIGNED NULL AFTER code"}, {"finance_products", "organization_id", "BIGINT UNSIGNED NULL AFTER company_id"}, {"collection_companies", "organization_id", "BIGINT UNSIGNED NULL AFTER code"},
		{"vehicles", "organization_id", "BIGINT UNSIGNED NULL AFTER owner_phone"}, {"vehicles", "vehicle_type", "VARCHAR(32) NOT NULL DEFAULT '' AFTER organization_id"},
		{"vehicles", "brand_model", "VARCHAR(64) NOT NULL DEFAULT '' AFTER vehicle_type"}, {"vehicles", "engine_no", "VARCHAR(64) NOT NULL DEFAULT '' AFTER brand_model"},
		{"vehicles", "registration_date", "DATE NULL AFTER engine_no"}, {"vehicles", "use_nature", "VARCHAR(32) NOT NULL DEFAULT '' AFTER registration_date"},
		{"vehicles", "insurance_expiry", "DATE NULL AFTER use_nature"}, {"vehicles", "inspection_expiry", "DATE NULL AFTER insurance_expiry"},
		{"vehicles", "mileage", "DECIMAL(12,1) NOT NULL DEFAULT 0 AFTER inspection_expiry"}, {"devices", "organization_id", "BIGINT UNSIGNED NULL AFTER iccid"},
		{"vehicles", "driver_name", "VARCHAR(64) NOT NULL DEFAULT '' AFTER mileage"}, {"vehicles", "driver_phone", "VARCHAR(32) NOT NULL DEFAULT '' AFTER driver_name"},
		{"vehicles", "driver_id_no", "VARCHAR(32) NOT NULL DEFAULT '' AFTER driver_phone"}, {"vehicles", "driver_license_no", "VARCHAR(32) NOT NULL DEFAULT '' AFTER driver_id_no"},
		{"vehicles", "driver_license_class", "VARCHAR(16) NOT NULL DEFAULT '' AFTER driver_license_no"}, {"vehicles", "driver_license_issue_date", "DATE NULL AFTER driver_license_class"},
		{"vehicles", "driver_license_expiry", "DATE NULL AFTER driver_license_issue_date"}, {"vehicles", "driving_license_no", "VARCHAR(64) NOT NULL DEFAULT '' AFTER driver_license_expiry"},
		{"vehicles", "registration_authority", "VARCHAR(128) NOT NULL DEFAULT '' AFTER driving_license_no"}, {"vehicles", "approved_load", "DECIMAL(12,1) NOT NULL DEFAULT 0 AFTER registration_authority"},
		{"vehicles", "overall_dimensions", "VARCHAR(64) NOT NULL DEFAULT '' AFTER approved_load"}, {"vehicles", "fuel_type", "VARCHAR(32) NOT NULL DEFAULT '' AFTER overall_dimensions"},
		{"vehicles", "emission_standard", "VARCHAR(32) NOT NULL DEFAULT '' AFTER fuel_type"}, {"vehicles", "driving_license_issue_date", "DATE NULL AFTER emission_standard"},
		{"vehicles", "photo_vehicle_front", "VARCHAR(255) NOT NULL DEFAULT '' AFTER driving_license_issue_date"}, {"vehicles", "photo_vehicle_rear", "VARCHAR(255) NOT NULL DEFAULT '' AFTER photo_vehicle_front"},
		{"vehicles", "photo_driving_license_front", "VARCHAR(255) NOT NULL DEFAULT '' AFTER photo_vehicle_rear"}, {"vehicles", "photo_driving_license_back", "VARCHAR(255) NOT NULL DEFAULT '' AFTER photo_driving_license_front"},
		{"vehicles", "photo_driver_license_front", "VARCHAR(255) NOT NULL DEFAULT '' AFTER photo_driving_license_back"}, {"vehicles", "photo_driver_license_back", "VARCHAR(255) NOT NULL DEFAULT '' AFTER photo_driver_license_front"},
		{"vehicles", "delegated_org_id", "BIGINT UNSIGNED NULL AFTER photo_driver_license_back"}, {"vehicles", "delegation_status", "VARCHAR(16) NOT NULL DEFAULT 'none' AFTER delegated_org_id"},
		{"vehicles", "delegated_at", "DATETIME NULL AFTER delegation_status"}, {"vehicles", "delegation_note", "VARCHAR(255) NOT NULL DEFAULT '' AFTER delegated_at"},
		{"vehicles", "created_by", "VARCHAR(64) NOT NULL DEFAULT '系统管理员' AFTER remark"},
		{"organizations", "created_by", "VARCHAR(64) NOT NULL DEFAULT '系统管理员' AFTER status"},
		{"vehicle_records", "created_by", "VARCHAR(64) NOT NULL DEFAULT '系统管理员' AFTER attachment_url"},
		{"devices", "created_by", "VARCHAR(64) NOT NULL DEFAULT '系统管理员' AFTER remark"},
		{"device_maintenance", "created_by", "VARCHAR(64) NOT NULL DEFAULT '系统管理员' AFTER handled_at"},
		{"finance_companies", "created_by", "VARCHAR(64) NOT NULL DEFAULT '系统管理员' AFTER remark"},
		{"finance_products", "created_by", "VARCHAR(64) NOT NULL DEFAULT '系统管理员' AFTER remark"},
		{"collection_companies", "created_by", "VARCHAR(64) NOT NULL DEFAULT '系统管理员' AFTER remark"},
		{"vehicle_records", "attachment_url", "VARCHAR(255) NOT NULL DEFAULT '' AFTER detail"},
		{"devices", "vehicle_id", "BIGINT UNSIGNED NULL AFTER organization_id"},
		{"devices", "install_position", "VARCHAR(64) NOT NULL DEFAULT '' AFTER vehicle_id"},
		{"devices", "installer", "VARCHAR(64) NOT NULL DEFAULT '' AFTER install_position"},
		{"devices", "installer_phone", "VARCHAR(32) NOT NULL DEFAULT '' AFTER installer"},
		{"devices", "photo_vin", "VARCHAR(255) NOT NULL DEFAULT '' AFTER installer_phone"},
		{"devices", "photo_position", "VARCHAR(255) NOT NULL DEFAULT '' AFTER photo_vin"},
		{"devices", "photo_vehicle", "VARCHAR(255) NOT NULL DEFAULT '' AFTER photo_position"},
		{"devices", "bound_vehicle_vin", "VARCHAR(64) NOT NULL DEFAULT '' AFTER photo_vehicle"},
		{"devices", "bound_at", "DATETIME NULL AFTER bound_vehicle_vin"},
		{"devices", "service_start_time", "DATETIME NULL AFTER bound_at"},
		{"devices", "service_end_time", "DATETIME NULL AFTER service_start_time"},
		{"devices", "service_duration_months", "INT UNSIGNED NOT NULL DEFAULT 0 AFTER service_end_time"},
		{"devices", "device_key", "VARCHAR(255) NOT NULL DEFAULT '' AFTER service_duration_months"},
	}
	for _, column := range columns {
		var count int
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name=? AND column_name=?`, column.table, column.name).Scan(&count); err != nil {
			return err
		}
		if count == 0 {
			if _, err := db.ExecContext(ctx, "ALTER TABLE `"+column.table+"` ADD COLUMN `"+column.name+"` "+column.definition); err != nil {
				return err
			}
		}
	}
	// 早期版本使用 VARCHAR(16)，无法存储 pending_production、pending_confirmation 等完整状态值。
	if _, err := db.ExecContext(ctx, `ALTER TABLE devices MODIFY COLUMN inventory_status VARCHAR(32) NOT NULL DEFAULT 'pending_production'`); err != nil {
		return fmt.Errorf("扩展设备库存状态字段失败: %w", err)
	}
	// 兼容安装档案上线前已经存在的绑定关系：数据库无法还原真实绑定时刻，
	// 使用设备最后更新时间作为最接近的历史时间，同时补齐绑定车辆 VIN。
	if _, err := db.ExecContext(ctx, `UPDATE devices d JOIN vehicles v ON v.id=d.vehicle_id SET d.bound_at=COALESCE(d.bound_at,d.updated_at,d.created_at),d.bound_vehicle_vin=CASE WHEN d.bound_vehicle_vin='' THEN v.vin ELSE d.bound_vehicle_vin END WHERE d.vehicle_id IS NOT NULL AND (d.bound_at IS NULL OR d.bound_vehicle_vin='')`); err != nil {
		return fmt.Errorf("补齐历史设备绑定信息失败: %w", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO organizations(parent_id,name,code,status) VALUES(0,'张三科技有限公司','ZS',1) ON DUPLICATE KEY UPDATE name=VALUES(name)`); err != nil {
		return err
	}
	if err := backfillOrganizationOwnership(ctx, db); err != nil {
		return err
	}
	return nil
}

// backfillOrganizationOwnership makes the new data-range constraint safe for
// existing installations.  Historical records predate organization_id, so we
// attach them to the first top-level organization instead of silently making
// them disappear for every non-admin account.  Admin remains global.
func backfillOrganizationOwnership(ctx context.Context, db *sql.DB) error {
	var rootID int64
	if err := db.QueryRowContext(ctx, `SELECT id FROM organizations WHERE parent_id=0 ORDER BY id LIMIT 1`).Scan(&rootID); err != nil {
		return fmt.Errorf("获取默认机构失败: %w", err)
	}
	updates := []struct {
		query string
		args  []any
	}{
		{`UPDATE vehicles SET organization_id=? WHERE organization_id IS NULL`, []any{rootID}},
		{`UPDATE devices d LEFT JOIN vehicles v ON v.id=d.vehicle_id SET d.organization_id=COALESCE(v.organization_id,?) WHERE d.organization_id IS NULL`, []any{rootID}},
		{`UPDATE finance_companies SET organization_id=? WHERE organization_id IS NULL`, []any{rootID}},
		{`UPDATE finance_products p JOIN finance_companies c ON c.id=p.company_id SET p.organization_id=c.organization_id WHERE p.organization_id IS NULL`, nil},
		{`UPDATE collection_companies SET organization_id=? WHERE organization_id IS NULL`, []any{rootID}},
		{`UPDATE roles SET organization_id=? WHERE code<>'admin' AND organization_id IS NULL`, []any{rootID}},
		{`UPDATE users u LEFT JOIN roles r ON r.id=u.role_id SET u.organization_id=COALESCE(r.organization_id,?) WHERE u.username<>'admin' AND u.organization_id IS NULL`, []any{rootID}},
	}
	for _, update := range updates {
		if _, err := db.ExecContext(ctx, update.query, update.args...); err != nil {
			return fmt.Errorf("补齐机构归属失败: %w", err)
		}
	}
	return nil
}

func migrateRBAC(ctx context.Context, db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS roles (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, code VARCHAR(64) NOT NULL, name VARCHAR(64) NOT NULL,
			description VARCHAR(255) NOT NULL DEFAULT '', status TINYINT UNSIGNED NOT NULL DEFAULT 1,
			is_system TINYINT UNSIGNED NOT NULL DEFAULT 0, organization_id BIGINT UNSIGNED NULL, created_by VARCHAR(64) NOT NULL DEFAULT '系统管理员', created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
			updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
			PRIMARY KEY(id), UNIQUE INDEX uk_roles_code(code)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS menus (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, parent_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
			name VARCHAR(64) NOT NULL, code VARCHAR(128) NOT NULL, menu_type VARCHAR(16) NOT NULL DEFAULT 'menu',
			path VARCHAR(255) NOT NULL DEFAULT '', icon VARCHAR(64) NOT NULL DEFAULT '', sort_order INT NOT NULL DEFAULT 0,
			status TINYINT UNSIGNED NOT NULL DEFAULT 1, created_by VARCHAR(64) NOT NULL DEFAULT '系统管理员', created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
			updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
			PRIMARY KEY(id), UNIQUE INDEX uk_menus_code(code), INDEX idx_menus_parent(parent_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS role_menus (
			role_id BIGINT UNSIGNED NOT NULL, menu_id BIGINT UNSIGNED NOT NULL,
			PRIMARY KEY(role_id,menu_id), INDEX idx_role_menus_menu(menu_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	}
	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("创建RBAC数据表失败: %w", err)
		}
	}
	for _, column := range []struct{ table, name, definition string }{
		{"roles", "created_by", "VARCHAR(64) NOT NULL DEFAULT '系统管理员' AFTER is_system"},
		{"roles", "organization_id", "BIGINT UNSIGNED NULL AFTER is_system"},
		{"menus", "created_by", "VARCHAR(64) NOT NULL DEFAULT '系统管理员' AFTER status"},
	} {
		var count int
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name=? AND column_name=?`, column.table, column.name).Scan(&count); err != nil {
			return err
		}
		if count == 0 {
			if _, err := db.ExecContext(ctx, "ALTER TABLE `"+column.table+"` ADD COLUMN `"+column.name+"` "+column.definition); err != nil {
				return err
			}
		}
	}
	var roleID sql.NullInt64
	if err := db.QueryRowContext(ctx, `SELECT id FROM roles WHERE code='admin'`).Scan(&roleID); err == sql.ErrNoRows {
		result, insertErr := db.ExecContext(ctx, `INSERT INTO roles(code,name,description,status,is_system) VALUES('admin','系统管理员','拥有全部系统管理权限',1,1)`)
		if insertErr != nil {
			return fmt.Errorf("初始化管理员角色失败: %w", insertErr)
		}
		roleID.Int64, _ = result.LastInsertId()
		roleID.Valid = true
	} else if err != nil {
		return fmt.Errorf("检查管理员角色失败: %w", err)
	}
	type menuSeed struct {
		parentCode, name, code, menuType, path, icon string
		sort                                         int
	}
	menuSeeds := []menuSeed{
		{"", "工作台", "workbench", "directory", "/workbench", "DesktopOutlined", 100},
		{"workbench", "协议解析", "workbench:parser", "menu", "/parser", "CodeOutlined", 110},
		{"workbench:parser", "解析报文", "workbench:parser:analyze", "button", "", "PlayCircleOutlined", 111},
		{"workbench:parser", "复制解析结果", "workbench:parser:copy", "button", "", "CopyOutlined", 112},
		{"workbench:parser", "查看解析记录", "workbench:parser:history", "button", "", "HistoryOutlined", 113},
		{"workbench:parser", "删除解析记录", "workbench:parser:history-delete", "button", "", "DeleteOutlined", 114},
		{"", "基础信息", "basic", "directory", "/basic", "DatabaseOutlined", 500},
		{"basic", "组织机构管理", "basic:organization", "menu", "/basic/organizations", "ApartmentOutlined", 501},
		{"basic:organization", "查询", "basic:organization:query", "button", "", "SearchOutlined", 502},
		{"basic:organization", "新增组织", "basic:organization:add", "button", "", "PlusOutlined", 502},
		{"basic:organization", "编辑组织", "basic:organization:edit", "button", "", "EditOutlined", 503},
		{"basic:organization", "删除组织", "basic:organization:delete", "button", "", "DeleteOutlined", 504},
		{"basic:organization", "批量导入组织", "basic:organization:import", "button", "", "UploadOutlined", 505},
		{"basic:organization", "下载导入模板", "basic:organization:template", "button", "", "DownloadOutlined", 506},
		{"basic", "车辆综合管理", "basic:vehicle", "menu", "/basic/vehicles", "CarOutlined", 510},
		{"basic:vehicle", "查询", "basic:vehicle:query", "button", "", "SearchOutlined", 511},
		{"basic:vehicle", "新增车辆", "basic:vehicle:add", "button", "", "PlusOutlined", 511},
		{"basic:vehicle", "编辑车辆", "basic:vehicle:edit", "button", "", "EditOutlined", 512},
		{"basic:vehicle", "删除车辆", "basic:vehicle:delete", "button", "", "DeleteOutlined", 513},
		{"basic:vehicle", "绑定设备", "basic:vehicle:bind", "button", "", "LinkOutlined", 514},
		{"basic:vehicle", "解绑设备", "basic:vehicle:unbind", "button", "", "DisconnectOutlined", 515},
		{"basic:vehicle", "批量绑定设备", "basic:vehicle:batch-bind", "button", "", "LinkOutlined", 516},
		{"basic:vehicle", "批量解绑设备", "basic:vehicle:batch-unbind", "button", "", "DisconnectOutlined", 517},
		{"basic:vehicle", "变更机构", "basic:vehicle:change-org", "button", "", "SwapOutlined", 518},
		{"basic:vehicle", "批量变更机构", "basic:vehicle:batch-change-org", "button", "", "SwapOutlined", 519},
		{"basic:vehicle", "委托管理", "basic:vehicle:delegate", "button", "", "AuditOutlined", 520},
		{"basic:vehicle", "批量委托", "basic:vehicle:batch-delegate", "button", "", "AuditOutlined", 521},
		{"basic:vehicle", "批量删除", "basic:vehicle:batch-delete", "button", "", "DeleteOutlined", 522},
		{"basic:vehicle", "批量导入", "basic:vehicle:import", "button", "", "UploadOutlined", 523},
		{"basic:vehicle", "批量导出", "basic:vehicle:export", "button", "", "DownloadOutlined", 524},
		{"basic:vehicle", "一键换绑", "basic:vehicle:swap-device", "button", "", "SwapOutlined", 525},
		{"basic:vehicle", "维护业务记录", "basic:vehicle:record", "button", "", "FileTextOutlined", 526},
		{"basic", "设备管理", "basic:device", "directory", "/basic/devices", "HddOutlined", 520},
		{"basic:device", "设备库存管理", "basic:device:inventory", "menu", "/basic/devices/inventory", "InboxOutlined", 521},
		{"basic:device:inventory", "查询", "basic:device:inventory:query", "button", "", "SearchOutlined", 522},
		{"basic:device:inventory", "设备入库", "basic:device:inventory:add", "button", "", "PlusOutlined", 522},
		{"basic:device:inventory", "编辑设备", "basic:device:inventory:edit", "button", "", "EditOutlined", 523},
		{"basic:device:inventory", "删除设备", "basic:device:inventory:delete", "button", "", "DeleteOutlined", 524},
		{"basic:device:inventory", "批量导入", "basic:device:inventory:import", "button", "", "UploadOutlined", 525},
		{"basic:device:inventory", "批量导出", "basic:device:inventory:export", "button", "", "DownloadOutlined", 526},
		{"basic:device:inventory", "状态确认", "basic:device:inventory:confirm", "button", "", "CheckCircleOutlined", 527},
		{"basic:device:inventory", "设备迁移", "basic:device:inventory:migrate", "button", "", "SwapOutlined", 528},
		{"basic:device:inventory", "设备续费", "basic:device:inventory:renew", "button", "", "ClockCircleOutlined", 529},
		{"basic:device", "设备生命周期管理", "basic:device:maintenance", "menu", "/basic/devices/maintenance", "ToolOutlined", 530},
		{"basic:device:maintenance", "查询", "basic:device:maintenance:query", "button", "", "SearchOutlined", 531},
		{"basic:device:maintenance", "新增运维记录", "basic:device:maintenance:add", "button", "", "PlusOutlined", 531},
		{"basic:device:maintenance", "编辑运维记录", "basic:device:maintenance:edit", "button", "", "EditOutlined", 532},
		{"basic:device:maintenance", "删除运维记录", "basic:device:maintenance:delete", "button", "", "DeleteOutlined", 533},
		{"basic", "合作方管理", "basic:partner", "directory", "/basic/partners", "BankOutlined", 540},
		{"basic:partner", "金融公司管理", "basic:partner:finance-company", "menu", "/basic/partners/finance-companies", "BankOutlined", 541},
		{"basic:partner:finance-company", "查询", "basic:partner:finance-company:query", "button", "", "SearchOutlined", 542},
		{"basic:partner:finance-company", "新增金融公司", "basic:partner:finance-company:add", "button", "", "PlusOutlined", 542},
		{"basic:partner:finance-company", "编辑金融公司", "basic:partner:finance-company:edit", "button", "", "EditOutlined", 543},
		{"basic:partner:finance-company", "删除金融公司", "basic:partner:finance-company:delete", "button", "", "DeleteOutlined", 544},
		{"basic:partner", "金融产品管理", "basic:partner:finance-product", "menu", "/basic/partners/finance-products", "FundOutlined", 542},
		{"basic:partner:finance-product", "查询", "basic:partner:finance-product:query", "button", "", "SearchOutlined", 545},
		{"basic:partner:finance-product", "新增金融产品", "basic:partner:finance-product:add", "button", "", "PlusOutlined", 545},
		{"basic:partner:finance-product", "编辑金融产品", "basic:partner:finance-product:edit", "button", "", "EditOutlined", 546},
		{"basic:partner:finance-product", "删除金融产品", "basic:partner:finance-product:delete", "button", "", "DeleteOutlined", 547},
		{"basic:partner", "清收公司管理", "basic:partner:collection-company", "menu", "/basic/partners/collection-companies", "TeamOutlined", 543},
		{"basic:partner:collection-company", "查询", "basic:partner:collection-company:query", "button", "", "SearchOutlined", 548},
		{"basic:partner:collection-company", "新增清收公司", "basic:partner:collection-company:add", "button", "", "PlusOutlined", 548},
		{"basic:partner:collection-company", "编辑清收公司", "basic:partner:collection-company:edit", "button", "", "EditOutlined", 549},
		{"basic:partner:collection-company", "删除清收公司", "basic:partner:collection-company:delete", "button", "", "DeleteOutlined", 550},
		{"", "系统管理", "system", "directory", "/system", "SettingOutlined", 900},
		{"system", "用户管理", "system:user", "menu", "/system/users", "UserOutlined", 910},
		{"system:user", "查询", "system:user:query", "button", "", "SearchOutlined", 911},
		{"system:user", "新增用户", "system:user:add", "button", "", "PlusOutlined", 912},
		{"system:user", "编辑用户", "system:user:edit", "button", "", "EditOutlined", 913},
		{"system:user", "删除用户", "system:user:delete", "button", "", "DeleteOutlined", 914},
		{"system:user", "重置密码", "system:user:password", "button", "", "KeyOutlined", 915},
		{"system", "角色管理", "system:role", "menu", "/system/roles", "TeamOutlined", 920},
		{"system:role", "查询", "system:role:query", "button", "", "SearchOutlined", 921},
		{"system:role", "新增角色", "system:role:add", "button", "", "PlusOutlined", 922},
		{"system:role", "编辑角色", "system:role:edit", "button", "", "EditOutlined", 923},
		{"system:role", "删除角色", "system:role:delete", "button", "", "DeleteOutlined", 924},
		{"system:role", "分配权限", "system:role:permission", "button", "", "SafetyCertificateOutlined", 925},
		{"system", "菜单管理", "system:menu", "menu", "/system/menus", "MenuOutlined", 930},
		{"system:menu", "查询", "system:menu:query", "button", "", "SearchOutlined", 931},
		{"system:menu", "新增权限项", "system:menu:add", "button", "", "PlusOutlined", 932},
		{"system:menu", "编辑权限项", "system:menu:edit", "button", "", "EditOutlined", 933},
		{"system:menu", "删除权限项", "system:menu:delete", "button", "", "DeleteOutlined", 934},
		{"system", "系统设置", "system:settings", "menu", "/system/settings", "SettingOutlined", 940},
		{"system:settings", "保存设置", "system:settings:save", "button", "", "SaveOutlined", 941},
	}
	menuIDs := make(map[string]int64, len(menuSeeds))
	for _, seed := range menuSeeds {
		parentID := int64(0)
		if seed.parentCode != "" {
			parentID = menuIDs[seed.parentCode]
		}
		_, err := db.ExecContext(ctx, `INSERT INTO menus(parent_id,name,code,menu_type,path,icon,sort_order,status) VALUES(?,?,?,?,?,?,?,1) ON DUPLICATE KEY UPDATE parent_id=VALUES(parent_id),name=VALUES(name),menu_type=VALUES(menu_type),path=VALUES(path),icon=VALUES(icon),sort_order=VALUES(sort_order)`, parentID, seed.name, seed.code, seed.menuType, seed.path, seed.icon, seed.sort)
		if err != nil {
			return fmt.Errorf("初始化系统菜单失败: %w", err)
		}
		var menuID int64
		if err := db.QueryRowContext(ctx, `SELECT id FROM menus WHERE code=?`, seed.code).Scan(&menuID); err != nil {
			return err
		}
		menuIDs[seed.code] = menuID
	}
	var columnCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='users' AND column_name='role_id'`).Scan(&columnCount); err != nil {
		return err
	}
	if columnCount == 0 {
		if _, err := db.ExecContext(ctx, `ALTER TABLE users ADD COLUMN role_id BIGINT UNSIGNED NULL AFTER role, ADD INDEX idx_users_role_id(role_id)`); err != nil {
			return fmt.Errorf("升级用户角色字段失败: %w", err)
		}
	}
	if roleID.Valid {
		adminUsername := envOr("ADMIN_USERNAME", "admin")
		if _, err := db.ExecContext(ctx, `UPDATE users SET role='admin',role_id=?,status=1 WHERE username=?`, roleID.Int64, adminUsername); err != nil {
			return err
		}
		if _, err := db.ExecContext(ctx, `INSERT IGNORE INTO role_menus(role_id,menu_id) SELECT ?,id FROM menus`, roleID.Int64); err != nil {
			return err
		}
	}
	return nil
}

func migrateSystemSettings(ctx context.Context, db *sql.DB) error {
	const schema = `CREATE TABLE IF NOT EXISTS system_settings (
		setting_key VARCHAR(64) NOT NULL,
		setting_value JSON NOT NULL,
		updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
		PRIMARY KEY (setting_key)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("创建系统设置表失败: %w", err)
	}
	const preferenceSchema = `CREATE TABLE IF NOT EXISTS user_preferences (
		user_id BIGINT UNSIGNED NOT NULL, preference_key VARCHAR(128) NOT NULL, preference_value JSON NOT NULL,
		updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
		PRIMARY KEY(user_id,preference_key), INDEX idx_user_preferences_user(user_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
	if _, err := db.ExecContext(ctx, preferenceSchema); err != nil {
		return fmt.Errorf("创建用户偏好表失败: %w", err)
	}
	const defaultLogin = `{"layoutType":"split","splitImage":"/iot-login-hero-v2.png","backgroundImage":"/iot-login-fullscreen-clean.png","overlayOpacity":0,"animationEnabled":true,"navigationType":"sidebar"}`
	if _, err := db.ExecContext(ctx, `INSERT IGNORE INTO system_settings (setting_key,setting_value) VALUES ('login_page', CAST(? AS JSON))`, defaultLogin); err != nil {
		return fmt.Errorf("初始化登录页设置失败: %w", err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE system_settings
		SET setting_value = JSON_SET(setting_value,
			'$.splitImage', '/iot-login-hero-v2.png',
			'$.backgroundImage', '/iot-login-fullscreen-clean.png',
			'$.overlayOpacity', 0)
		WHERE setting_key = 'login_page'
			AND JSON_UNQUOTE(JSON_EXTRACT(setting_value, '$.backgroundImage')) IN ('/iot-login-hero-v2.png', '/iot-login-fullscreen.png')`); err != nil {
		return fmt.Errorf("升级登录页默认背景失败: %w", err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE system_settings SET setting_value=JSON_SET(setting_value,'$.navigationType','sidebar') WHERE setting_key='login_page' AND JSON_EXTRACT(setting_value,'$.navigationType') IS NULL`); err != nil {
		return fmt.Errorf("初始化后台导航设置失败: %w", err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE system_settings SET setting_value=JSON_SET(setting_value,
		'$.systemName',COALESCE(JSON_UNQUOTE(JSON_EXTRACT(setting_value,'$.systemName')),'协议解析工具'),
		'$.systemNameEn',COALESCE(JSON_UNQUOTE(JSON_EXTRACT(setting_value,'$.systemNameEn')),'Protocol Parser Tool'),
		'$.menuShortName',COALESCE(JSON_UNQUOTE(JSON_EXTRACT(setting_value,'$.menuShortName')),'协议解析工具'),
		'$.browserTitleMode',COALESCE(JSON_UNQUOTE(JSON_EXTRACT(setting_value,'$.browserTitleMode')),'system'),
		'$.browserTitle',COALESCE(JSON_UNQUOTE(JSON_EXTRACT(setting_value,'$.browserTitle')),''),
		'$.systemIcon',COALESCE(JSON_UNQUOTE(JSON_EXTRACT(setting_value,'$.systemIcon')),'/favicon.png'),
		'$.footerCopyright',COALESCE(JSON_UNQUOTE(JSON_EXTRACT(setting_value,'$.footerCopyright')),'智能风控云平台'),
		'$.footerSlogan',COALESCE(JSON_UNQUOTE(JSON_EXTRACT(setting_value,'$.footerSlogan')),'让协议解析更简单高效'),
		'$.developerName',COALESCE(JSON_UNQUOTE(JSON_EXTRACT(setting_value,'$.developerName')),'张三科技有限公司'),
		'$.developerPhone',COALESCE(JSON_UNQUOTE(JSON_EXTRACT(setting_value,'$.developerPhone')),''),
		'$.systemVersion',COALESCE(JSON_UNQUOTE(JSON_EXTRACT(setting_value,'$.systemVersion')),'V1.0.0')) WHERE setting_key='login_page'`); err != nil {
		return fmt.Errorf("初始化系统品牌设置失败: %w", err)
	}
	return nil
}

func migrateUsers(ctx context.Context, db *sql.DB) error {
	const schema = `CREATE TABLE IF NOT EXISTS users (
		id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
		username VARCHAR(64) NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		display_name VARCHAR(64) NOT NULL,
		role VARCHAR(32) NOT NULL DEFAULT 'admin',
		organization_id BIGINT UNSIGNED NULL,
		email VARCHAR(128) NOT NULL DEFAULT '',
		phone VARCHAR(32) NOT NULL DEFAULT '',
		status TINYINT UNSIGNED NOT NULL DEFAULT 1,
		last_login_at DATETIME(3) NULL,
		created_by VARCHAR(64) NOT NULL DEFAULT '系统管理员',
		created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
		updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
		PRIMARY KEY (id), UNIQUE INDEX uk_users_username (username)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("创建用户表失败: %w", err)
	}
	var createdByCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='users' AND column_name='created_by'`).Scan(&createdByCount); err != nil {
		return err
	}
	if createdByCount == 0 {
		if _, err := db.ExecContext(ctx, `ALTER TABLE users ADD COLUMN created_by VARCHAR(64) NOT NULL DEFAULT '系统管理员' AFTER last_login_at`); err != nil {
			return err
		}
	}
	username := envOr("ADMIN_USERNAME", "admin")
	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE username=?`, username).Scan(&count); err != nil {
		return fmt.Errorf("检查管理员账号失败: %w", err)
	}
	if count > 0 {
		return nil
	}
	password := strings.TrimSpace(os.Getenv("ADMIN_PASSWORD"))
	if password == "" {
		return fmt.Errorf("users表为空，首次启动必须配置ADMIN_PASSWORD")
	}
	if len(password) < 10 || strings.EqualFold(password, "admin123") {
		return fmt.Errorf("ADMIN_PASSWORD必须至少10位，且不能使用默认弱密码admin123")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("生成管理员密码失败: %w", err)
	}
	if _, err = db.ExecContext(ctx, `INSERT INTO users (username,password_hash,display_name,role) VALUES (?,?,?,'admin')`, username, string(hash), "系统管理员"); err != nil {
		return fmt.Errorf("创建管理员账号失败: %w", err)
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
