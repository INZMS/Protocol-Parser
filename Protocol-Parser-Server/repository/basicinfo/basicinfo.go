package basicinfo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type Organization struct {
	ID         int64  `json:"id"`
	ParentID   int64  `json:"parentId"`
	ParentCode string `json:"parentCode,omitempty"`
	Name       string `json:"name"`
	Code       string `json:"code"`
	Status     int    `json:"status"`
}
type Vehicle struct {
	ID                       int64   `json:"id"`
	PlateNo                  string  `json:"plateNo"`
	PlateColor               string  `json:"plateColor"`
	VIN                      string  `json:"vin"`
	OwnerName                string  `json:"ownerName"`
	OwnerPhone               string  `json:"ownerPhone"`
	OrganizationID           *int64  `json:"organizationId"`
	OrganizationName         string  `json:"organizationName"`
	VehicleType              string  `json:"vehicleType"`
	BrandModel               string  `json:"brandModel"`
	EngineNo                 string  `json:"engineNo"`
	RegistrationDate         string  `json:"registrationDate"`
	UseNature                string  `json:"useNature"`
	InsuranceExpiry          string  `json:"insuranceExpiry"`
	InspectionExpiry         string  `json:"inspectionExpiry"`
	Mileage                  float64 `json:"mileage"`
	DriverName               string  `json:"driverName"`
	DriverPhone              string  `json:"driverPhone"`
	DriverIDNo               string  `json:"driverIdNo"`
	DriverLicenseNo          string  `json:"driverLicenseNo"`
	DriverLicenseClass       string  `json:"driverLicenseClass"`
	DriverLicenseIssueDate   string  `json:"driverLicenseIssueDate"`
	DriverLicenseExpiry      string  `json:"driverLicenseExpiry"`
	DrivingLicenseNo         string  `json:"drivingLicenseNo"`
	RegistrationAuthority    string  `json:"registrationAuthority"`
	ApprovedLoad             float64 `json:"approvedLoad"`
	OverallDimensions        string  `json:"overallDimensions"`
	FuelType                 string  `json:"fuelType"`
	EmissionStandard         string  `json:"emissionStandard"`
	DrivingLicenseIssueDate  string  `json:"drivingLicenseIssueDate"`
	PhotoVehicleFront        string  `json:"photoVehicleFront"`
	PhotoVehicleRear         string  `json:"photoVehicleRear"`
	PhotoDrivingLicenseFront string  `json:"photoDrivingLicenseFront"`
	PhotoDrivingLicenseBack  string  `json:"photoDrivingLicenseBack"`
	PhotoDriverLicenseFront  string  `json:"photoDriverLicenseFront"`
	PhotoDriverLicenseBack   string  `json:"photoDriverLicenseBack"`
	DelegatedOrgID           *int64  `json:"delegatedOrgId"`
	DelegatedOrgName         string  `json:"delegatedOrgName"`
	DelegationStatus         string  `json:"delegationStatus"`
	DelegatedAt              string  `json:"delegatedAt"`
	DelegationNote           string  `json:"delegationNote"`
	BoundDeviceCount         int     `json:"boundDeviceCount"`
	BoundDeviceNos           string  `json:"boundDeviceNos"`
	Status                   int     `json:"status"`
	Remark                   string  `json:"remark"`
	CreatedBy                string  `json:"createdBy"`
	CreatedAt                string  `json:"createdAt"`
}
type Device struct {
	ID                    int64  `json:"id"`
	DeviceNo              string `json:"deviceNo"`
	IMEI                  string `json:"imei"`
	Model                 string `json:"model"`
	Protocol              string `json:"protocol"`
	SIMNo                 string `json:"simNo"`
	ICCID                 string `json:"iccid"`
	OrganizationID        *int64 `json:"organizationId"`
	OrganizationName      string `json:"organizationName"`
	VehicleID             *int64 `json:"vehicleId"`
	PlateNo               string `json:"plateNo"`
	InstallPosition       string `json:"installPosition"`
	Installer             string `json:"installer"`
	InstallerPhone        string `json:"installerPhone"`
	PhotoVIN              string `json:"photoVin"`
	PhotoPosition         string `json:"photoPosition"`
	PhotoVehicle          string `json:"photoVehicle"`
	BoundVehicleVIN       string `json:"boundVehicleVin"`
	BoundAt               string `json:"boundAt"`
	ServiceStartTime      string `json:"serviceStartTime"`
	ServiceEndTime        string `json:"serviceEndTime"`
	ServiceDurationMonths int    `json:"serviceDurationMonths"`
	DeviceKey             string `json:"deviceKey"`
	InventoryStatus       string `json:"inventoryStatus"`
	InboundDate           string `json:"inboundDate"`
	Remark                string `json:"remark"`
	CreatedAt             string `json:"createdAt"`
}
type DeviceBinding struct {
	DeviceID              int64  `json:"deviceId"`
	InstallPosition       string `json:"installPosition"`
	Installer             string `json:"installer"`
	InstallerPhone        string `json:"installerPhone"`
	PhotoVIN              string `json:"photoVin"`
	PhotoPosition         string `json:"photoPosition"`
	PhotoVehicle          string `json:"photoVehicle"`
	ServiceDurationMonths int    `json:"serviceDurationMonths"`
}
type VehicleRecord struct {
	ID            int64   `json:"id"`
	VehicleID     int64   `json:"vehicleId"`
	PlateNo       string  `json:"plateNo"`
	RecordType    string  `json:"recordType"`
	RecordDate    string  `json:"recordDate"`
	Title         string  `json:"title"`
	Amount        float64 `json:"amount"`
	Mileage       float64 `json:"mileage"`
	Status        string  `json:"status"`
	Detail        string  `json:"detail"`
	AttachmentURL string  `json:"attachmentUrl"`
	CreatedAt     string  `json:"createdAt"`
}
type Maintenance struct {
	ID               int64  `json:"id"`
	DeviceID         int64  `json:"deviceId"`
	DeviceNo         string `json:"deviceNo"`
	MaintenanceType  string `json:"maintenanceType"`
	IssueDescription string `json:"issueDescription"`
	HandlingResult   string `json:"handlingResult"`
	Handler          string `json:"handler"`
	Status           string `json:"status"`
	HandledAt        string `json:"handledAt"`
	CreatedAt        string `json:"createdAt"`
}
type FinanceCompany struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Code         string `json:"code"`
	ContactName  string `json:"contactName"`
	ContactPhone string `json:"contactPhone"`
	Address      string `json:"address"`
	Status       int    `json:"status"`
	Remark       string `json:"remark"`
	CreatedAt    string `json:"createdAt"`
}
type FinanceProduct struct {
	ID          int64   `json:"id"`
	CompanyID   int64   `json:"companyId"`
	CompanyName string  `json:"companyName"`
	Name        string  `json:"name"`
	Code        string  `json:"code"`
	ProductType string  `json:"productType"`
	AnnualRate  float64 `json:"annualRate"`
	TermMonths  int     `json:"termMonths"`
	Status      int     `json:"status"`
	Remark      string  `json:"remark"`
	CreatedAt   string  `json:"createdAt"`
}
type CollectionCompany struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Code         string `json:"code"`
	ContactName  string `json:"contactName"`
	ContactPhone string `json:"contactPhone"`
	ServiceArea  string `json:"serviceArea"`
	Address      string `json:"address"`
	Status       int    `json:"status"`
	Remark       string `json:"remark"`
	CreatedAt    string `json:"createdAt"`
}
type Notification struct {
	ID               int64  `json:"id"`
	OrganizationID   *int64 `json:"organizationId"`
	OrganizationName string `json:"organizationName"`
	DeviceID         *int64 `json:"deviceId"`
	Type             string `json:"type"`
	Title            string `json:"title"`
	Content          string `json:"content"`
	IsRead           bool   `json:"isRead"`
	CreatedAt        string `json:"createdAt"`
}

type Store interface {
	ListOrganizations(context.Context) ([]Organization, error)
	SaveOrganization(context.Context, Organization) (int64, error)
	BatchSaveOrganizations(context.Context, []Organization) error
	DeleteOrganization(context.Context, int64) error
	ListVehicles(context.Context) ([]Vehicle, error)
	SaveVehicle(context.Context, Vehicle) (int64, error)
	DeleteVehicle(context.Context, int64) error
	BindDevices(context.Context, int64, []DeviceBinding, bool) error
	SwapDevice(context.Context, int64, int64) error
	UpdateDeviceInstallation(context.Context, int64, DeviceBinding) error
	UnbindDevices(context.Context, []int64) error
	ChangeVehicleOrganization(context.Context, []int64, int64, bool) error
	DelegateVehicles(context.Context, []int64, *int64, string) error
	BatchDeleteVehicles(context.Context, []int64, bool) error
	ListVehicleRecords(context.Context) ([]VehicleRecord, error)
	SaveVehicleRecord(context.Context, VehicleRecord) (int64, error)
	DeleteVehicleRecord(context.Context, int64) error
	ListDevices(context.Context) ([]Device, error)
	SaveDevice(context.Context, Device) (int64, error)
	DeleteDevice(context.Context, int64) error
	TransitionDeviceStatus(context.Context, int64, string) error
	BatchTransitionDeviceStatus(context.Context, []int64, string) error
	MigrateDevices(context.Context, []int64, int64) error
	RenewDevices(context.Context, []int64, int) error
	ListMaintenance(context.Context) ([]Maintenance, error)
	SaveMaintenance(context.Context, Maintenance) (int64, error)
	DeleteMaintenance(context.Context, int64) error
	ListFinanceCompanies(context.Context) ([]FinanceCompany, error)
	SaveFinanceCompany(context.Context, FinanceCompany) (int64, error)
	DeleteFinanceCompany(context.Context, int64) error
	ListFinanceProducts(context.Context) ([]FinanceProduct, error)
	SaveFinanceProduct(context.Context, FinanceProduct) (int64, error)
	DeleteFinanceProduct(context.Context, int64) error
	ListCollectionCompanies(context.Context) ([]CollectionCompany, error)
	SaveCollectionCompany(context.Context, CollectionCompany) (int64, error)
	DeleteCollectionCompany(context.Context, int64) error
	ListNotifications(context.Context) ([]Notification, error)
	ReadNotification(context.Context, int64) error
}
type MySQLStore struct{ db *sql.DB }

func NewMySQLStore(db *sql.DB) *MySQLStore { return &MySQLStore{db: db} }

func (s *MySQLStore) ListNotifications(ctx context.Context) ([]Notification, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT n.id,n.organization_id,COALESCE(o.name,''),n.device_id,n.type,n.title,n.content,n.is_read,DATE_FORMAT(n.created_at,'%Y-%m-%d %H:%i:%s') FROM system_notifications n LEFT JOIN organizations o ON o.id=n.organization_id ORDER BY n.is_read,n.id DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Notification{}
	for rows.Next() {
		var item Notification
		var orgID, deviceID sql.NullInt64
		if err = rows.Scan(&item.ID, &orgID, &item.OrganizationName, &deviceID, &item.Type, &item.Title, &item.Content, &item.IsRead, &item.CreatedAt); err != nil {
			return nil, err
		}
		if orgID.Valid {
			item.OrganizationID = &orgID.Int64
		}
		if deviceID.Valid {
			item.DeviceID = &deviceID.Int64
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *MySQLStore) ReadNotification(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `UPDATE system_notifications SET is_read=1 WHERE id=?`, id)
	return err
}

func (s *MySQLStore) ListOrganizations(ctx context.Context) ([]Organization, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,parent_id,name,code,status FROM organizations ORDER BY parent_id,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Organization{}
	for rows.Next() {
		var v Organization
		if err = rows.Scan(&v.ID, &v.ParentID, &v.Name, &v.Code, &v.Status); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *MySQLStore) SaveOrganization(ctx context.Context, v Organization) (int64, error) {
	v.Name = strings.TrimSpace(v.Name)
	v.Code = strings.TrimSpace(v.Code)
	if v.Name == "" || v.Code == "" {
		return 0, errors.New("组织名称和编码不能为空")
	}
	var duplicate int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM organizations WHERE code=? AND id<>?`, v.Code, v.ID).Scan(&duplicate); err != nil {
		return 0, err
	}
	if duplicate > 0 {
		return 0, errors.New("组织编码已存在，请更换后重试")
	}
	if v.ID == 0 {
		r, e := s.db.ExecContext(ctx, `INSERT INTO organizations(parent_id,name,code,status) VALUES(?,?,?,?)`, v.ParentID, v.Name, v.Code, v.Status)
		if e != nil {
			return 0, e
		}
		return r.LastInsertId()
	}
	_, e := s.db.ExecContext(ctx, `UPDATE organizations SET parent_id=?,name=?,code=?,status=? WHERE id=?`, v.ParentID, v.Name, v.Code, v.Status, v.ID)
	return v.ID, e
}
func (s *MySQLStore) BatchSaveOrganizations(ctx context.Context, items []Organization) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	codeIDs := map[string]int64{}
	rows, err := tx.QueryContext(ctx, `SELECT code,id FROM organizations`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var code string
		var id int64
		if err = rows.Scan(&code, &id); err != nil {
			rows.Close()
			return err
		}
		codeIDs[code] = id
	}
	if err = rows.Close(); err != nil {
		return err
	}
	for index, item := range items {
		item.Name = strings.TrimSpace(item.Name)
		item.Code = strings.TrimSpace(item.Code)
		item.ParentCode = strings.TrimSpace(item.ParentCode)
		if item.Name == "" || item.Code == "" {
			return fmt.Errorf("第%d行组织名称和编码不能为空", index+2)
		}
		if _, exists := codeIDs[item.Code]; exists {
			return fmt.Errorf("第%d行组织编码 %s 已存在", index+2, item.Code)
		}
		parentID := int64(0)
		if item.ParentCode != "" {
			var exists bool
			parentID, exists = codeIDs[item.ParentCode]
			if !exists {
				return fmt.Errorf("第%d行上级组织编码 %s 不存在或尚未导入", index+2, item.ParentCode)
			}
		}
		if item.Status != 0 {
			item.Status = 1
		}
		result, execErr := tx.ExecContext(ctx, `INSERT INTO organizations(parent_id,name,code,status) VALUES(?,?,?,?)`, parentID, item.Name, item.Code, item.Status)
		if execErr != nil {
			return execErr
		}
		id, execErr := result.LastInsertId()
		if execErr != nil {
			return execErr
		}
		codeIDs[item.Code] = id
	}
	return tx.Commit()
}
func (s *MySQLStore) DeleteOrganization(ctx context.Context, id int64) error {
	var count int
	_ = s.db.QueryRowContext(ctx, `SELECT (SELECT COUNT(*) FROM organizations WHERE parent_id=?)+(SELECT COUNT(*) FROM vehicles WHERE organization_id=?)+(SELECT COUNT(*) FROM devices WHERE organization_id=?)`, id, id, id).Scan(&count)
	if count > 0 {
		return errors.New("该组织存在下级组织、车辆或设备，不能删除")
	}
	_, e := s.db.ExecContext(ctx, `DELETE FROM organizations WHERE id=?`, id)
	return e
}

func (s *MySQLStore) ListVehicles(ctx context.Context) ([]Vehicle, error) {
	rows, e := s.db.QueryContext(ctx, `SELECT v.id,v.plate_no,v.plate_color,v.vin,v.owner_name,v.owner_phone,v.organization_id,COALESCE(o.name,''),v.vehicle_type,v.brand_model,v.engine_no,COALESCE(DATE_FORMAT(v.registration_date,'%Y-%m-%d'),''),v.use_nature,COALESCE(DATE_FORMAT(v.insurance_expiry,'%Y-%m-%d'),''),COALESCE(DATE_FORMAT(v.inspection_expiry,'%Y-%m-%d'),''),v.mileage,v.driver_name,v.driver_phone,v.driver_id_no,v.driver_license_no,v.driver_license_class,COALESCE(DATE_FORMAT(v.driver_license_issue_date,'%Y-%m-%d'),''),COALESCE(DATE_FORMAT(v.driver_license_expiry,'%Y-%m-%d'),''),v.driving_license_no,v.registration_authority,v.approved_load,v.overall_dimensions,v.fuel_type,v.emission_standard,COALESCE(DATE_FORMAT(v.driving_license_issue_date,'%Y-%m-%d'),''),v.photo_vehicle_front,v.photo_vehicle_rear,v.photo_driving_license_front,v.photo_driving_license_back,v.photo_driver_license_front,v.photo_driver_license_back,v.delegated_org_id,COALESCE(cc.name,''),v.delegation_status,COALESCE(DATE_FORMAT(v.delegated_at,'%Y-%m-%d %H:%i:%s'),''),v.delegation_note,(SELECT COUNT(*) FROM devices d WHERE d.vehicle_id=v.id),COALESCE((SELECT GROUP_CONCAT(d.device_no ORDER BY d.id SEPARATOR ',') FROM devices d WHERE d.vehicle_id=v.id),''),v.status,v.remark,v.created_by,DATE_FORMAT(v.created_at,'%Y-%m-%d %H:%i:%s') FROM vehicles v LEFT JOIN organizations o ON o.id=v.organization_id LEFT JOIN collection_companies cc ON cc.id=v.delegated_org_id ORDER BY v.id DESC`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Vehicle{}
	for rows.Next() {
		var v Vehicle
		var org sql.NullInt64
		var delegated sql.NullInt64
		if e = rows.Scan(&v.ID, &v.PlateNo, &v.PlateColor, &v.VIN, &v.OwnerName, &v.OwnerPhone, &org, &v.OrganizationName, &v.VehicleType, &v.BrandModel, &v.EngineNo, &v.RegistrationDate, &v.UseNature, &v.InsuranceExpiry, &v.InspectionExpiry, &v.Mileage, &v.DriverName, &v.DriverPhone, &v.DriverIDNo, &v.DriverLicenseNo, &v.DriverLicenseClass, &v.DriverLicenseIssueDate, &v.DriverLicenseExpiry, &v.DrivingLicenseNo, &v.RegistrationAuthority, &v.ApprovedLoad, &v.OverallDimensions, &v.FuelType, &v.EmissionStandard, &v.DrivingLicenseIssueDate, &v.PhotoVehicleFront, &v.PhotoVehicleRear, &v.PhotoDrivingLicenseFront, &v.PhotoDrivingLicenseBack, &v.PhotoDriverLicenseFront, &v.PhotoDriverLicenseBack, &delegated, &v.DelegatedOrgName, &v.DelegationStatus, &v.DelegatedAt, &v.DelegationNote, &v.BoundDeviceCount, &v.BoundDeviceNos, &v.Status, &v.Remark, &v.CreatedBy, &v.CreatedAt); e != nil {
			return nil, e
		}
		if org.Valid {
			v.OrganizationID = &org.Int64
		}
		if delegated.Valid {
			v.DelegatedOrgID = &delegated.Int64
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *MySQLStore) SaveVehicle(ctx context.Context, v Vehicle) (int64, error) {
	if v.ID == 0 {
		r, e := s.db.ExecContext(ctx, `INSERT INTO vehicles(plate_no,plate_color,vin,owner_name,owner_phone,organization_id,vehicle_type,brand_model,engine_no,registration_date,use_nature,insurance_expiry,inspection_expiry,mileage,driver_name,driver_phone,driver_id_no,driver_license_no,driver_license_class,driver_license_issue_date,driver_license_expiry,driving_license_no,registration_authority,approved_load,overall_dimensions,fuel_type,emission_standard,driving_license_issue_date,photo_vehicle_front,photo_vehicle_rear,photo_driving_license_front,photo_driving_license_back,photo_driver_license_front,photo_driver_license_back,status,remark,created_by) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, v.PlateNo, v.PlateColor, v.VIN, v.OwnerName, v.OwnerPhone, v.OrganizationID, v.VehicleType, v.BrandModel, v.EngineNo, dateOrNil(v.RegistrationDate), v.UseNature, dateOrNil(v.InsuranceExpiry), dateOrNil(v.InspectionExpiry), v.Mileage, v.DriverName, v.DriverPhone, v.DriverIDNo, v.DriverLicenseNo, v.DriverLicenseClass, dateOrNil(v.DriverLicenseIssueDate), dateOrNil(v.DriverLicenseExpiry), v.DrivingLicenseNo, v.RegistrationAuthority, v.ApprovedLoad, v.OverallDimensions, v.FuelType, v.EmissionStandard, dateOrNil(v.DrivingLicenseIssueDate), v.PhotoVehicleFront, v.PhotoVehicleRear, v.PhotoDrivingLicenseFront, v.PhotoDrivingLicenseBack, v.PhotoDriverLicenseFront, v.PhotoDriverLicenseBack, v.Status, v.Remark, v.CreatedBy)
		if e != nil {
			return 0, e
		}
		return r.LastInsertId()
	}
	_, e := s.db.ExecContext(ctx, `UPDATE vehicles SET plate_no=?,plate_color=?,vin=?,owner_name=?,owner_phone=?,organization_id=?,vehicle_type=?,brand_model=?,engine_no=?,registration_date=?,use_nature=?,insurance_expiry=?,inspection_expiry=?,mileage=?,driver_name=?,driver_phone=?,driver_id_no=?,driver_license_no=?,driver_license_class=?,driver_license_issue_date=?,driver_license_expiry=?,driving_license_no=?,registration_authority=?,approved_load=?,overall_dimensions=?,fuel_type=?,emission_standard=?,driving_license_issue_date=?,photo_vehicle_front=?,photo_vehicle_rear=?,photo_driving_license_front=?,photo_driving_license_back=?,photo_driver_license_front=?,photo_driver_license_back=?,status=?,remark=? WHERE id=?`, v.PlateNo, v.PlateColor, v.VIN, v.OwnerName, v.OwnerPhone, v.OrganizationID, v.VehicleType, v.BrandModel, v.EngineNo, dateOrNil(v.RegistrationDate), v.UseNature, dateOrNil(v.InsuranceExpiry), dateOrNil(v.InspectionExpiry), v.Mileage, v.DriverName, v.DriverPhone, v.DriverIDNo, v.DriverLicenseNo, v.DriverLicenseClass, dateOrNil(v.DriverLicenseIssueDate), dateOrNil(v.DriverLicenseExpiry), v.DrivingLicenseNo, v.RegistrationAuthority, v.ApprovedLoad, v.OverallDimensions, v.FuelType, v.EmissionStandard, dateOrNil(v.DrivingLicenseIssueDate), v.PhotoVehicleFront, v.PhotoVehicleRear, v.PhotoDrivingLicenseFront, v.PhotoDrivingLicenseBack, v.PhotoDriverLicenseFront, v.PhotoDriverLicenseBack, v.Status, v.Remark, v.ID)
	return v.ID, e
}
func (s *MySQLStore) DeleteVehicle(ctx context.Context, id int64) error {
	var count int
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM devices WHERE vehicle_id=?`, id).Scan(&count)
	if count > 0 {
		return errors.New("车辆已绑定设备，请先解除绑定关系后再删除")
	}
	_, _ = s.db.ExecContext(ctx, `DELETE FROM vehicle_records WHERE vehicle_id=?`, id)
	_, e := s.db.ExecContext(ctx, `DELETE FROM vehicles WHERE id=?`, id)
	return e
}

func (s *MySQLStore) BindDevices(ctx context.Context, vehicleID int64, bindings []DeviceBinding, force bool) error {
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	for _, binding := range bindings {
		id := binding.DeviceID
		var currentStatus string
		if e = tx.QueryRowContext(ctx, `SELECT inventory_status FROM devices WHERE id=? FOR UPDATE`, id).Scan(&currentStatus); e != nil {
			return e
		}
		if currentStatus != "pending_use" {
			return errors.New("只有待使用状态的设备可以绑定车辆")
		}
		if !force {
			var existing sql.NullInt64
			if e = tx.QueryRowContext(ctx, `SELECT vehicle_id FROM devices WHERE id=?`, id).Scan(&existing); e != nil {
				return e
			}
			if existing.Valid && existing.Int64 != vehicleID {
				return errors.New("所选设备已绑定其他车辆，请启用覆盖绑定")
			}
		}
		if binding.ServiceDurationMonths <= 0 {
			return errors.New("每台设备必须选择服务时长")
		}
		if _, e = tx.ExecContext(ctx, `UPDATE devices d JOIN vehicles v ON v.id=? SET d.vehicle_id=v.id,d.organization_id=v.organization_id,d.inventory_status='in_use',d.install_position=?,d.installer=?,d.installer_phone=?,d.photo_vin=?,d.photo_position=?,d.photo_vehicle=?,d.bound_vehicle_vin=v.vin,d.bound_at=NOW(),d.service_start_time=NOW(),d.service_end_time=DATE_ADD(NOW(),INTERVAL (d.service_duration_months+?) MONTH),d.service_duration_months=d.service_duration_months+? WHERE d.id=?`, vehicleID, binding.InstallPosition, binding.Installer, binding.InstallerPhone, binding.PhotoVIN, binding.PhotoPosition, binding.PhotoVehicle, binding.ServiceDurationMonths, binding.ServiceDurationMonths, id); e != nil {
			return e
		}
	}
	return tx.Commit()
}

// SwapDevice 将车辆、安装资料和服务期限从旧设备原子迁移到新设备。
// 后续数字钥匙等用户关系表应在此事务中追加迁移，避免设备和用户关系不同步。
func (s *MySQLStore) SwapDevice(ctx context.Context, oldDeviceID, newDeviceID int64) error {
	if oldDeviceID == 0 || newDeviceID == 0 || oldDeviceID == newDeviceID {
		return errors.New("新旧设备选择不正确")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var vehicleID sql.NullInt64
	var oldStatus string
	if err = tx.QueryRowContext(ctx, `SELECT vehicle_id,inventory_status FROM devices WHERE id=? FOR UPDATE`, oldDeviceID).Scan(&vehicleID, &oldStatus); err != nil {
		return err
	}
	if !vehicleID.Valid || oldStatus != "in_use" {
		return errors.New("旧设备未处于车辆使用中状态")
	}
	var newStatus string
	var newVehicleID sql.NullInt64
	if err = tx.QueryRowContext(ctx, `SELECT inventory_status,vehicle_id FROM devices WHERE id=? FOR UPDATE`, newDeviceID).Scan(&newStatus, &newVehicleID); err != nil {
		return err
	}
	if newStatus != "pending_use" || newVehicleID.Valid {
		return errors.New("新设备必须为待使用且未绑定状态")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE devices n JOIN devices o ON o.id=? SET n.vehicle_id=o.vehicle_id,n.organization_id=o.organization_id,n.inventory_status='in_use',n.install_position=o.install_position,n.installer=o.installer,n.installer_phone=o.installer_phone,n.photo_vin=o.photo_vin,n.photo_position=o.photo_position,n.photo_vehicle=o.photo_vehicle,n.bound_vehicle_vin=o.bound_vehicle_vin,n.bound_at=NOW(),n.service_start_time=o.service_start_time,n.service_end_time=o.service_end_time,n.service_duration_months=o.service_duration_months WHERE n.id=?`, oldDeviceID, newDeviceID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE devices SET vehicle_id=NULL,inventory_status='pending_confirmation',install_position='',installer='',installer_phone='',photo_vin='',photo_position='',photo_vehicle='',bound_vehicle_vin='',bound_at=NULL,service_start_time=NULL,service_end_time=NULL,service_duration_months=0 WHERE id=?`, oldDeviceID); err != nil {
		return err
	}
	return tx.Commit()
}
func (s *MySQLStore) UpdateDeviceInstallation(ctx context.Context, deviceID int64, binding DeviceBinding) error {
	result, e := s.db.ExecContext(ctx, `UPDATE devices SET install_position=?,installer=?,installer_phone=?,photo_vin=?,photo_position=?,photo_vehicle=? WHERE id=? AND vehicle_id IS NOT NULL`, binding.InstallPosition, binding.Installer, binding.InstallerPhone, binding.PhotoVIN, binding.PhotoPosition, binding.PhotoVehicle, deviceID)
	if e != nil {
		return e
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return errors.New("设备未绑定车辆或不存在")
	}
	return nil
}
func (s *MySQLStore) UnbindDevices(ctx context.Context, deviceIDs []int64) error {
	if len(deviceIDs) == 0 {
		return errors.New("请选择设备")
	}
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	for _, id := range deviceIDs {
		if _, e = tx.ExecContext(ctx, `UPDATE devices SET vehicle_id=NULL,inventory_status='pending_confirmation',install_position='',installer='',installer_phone='',photo_vin='',photo_position='',photo_vehicle='',bound_vehicle_vin='',bound_at=NULL,service_start_time=NULL,service_end_time=NULL,service_duration_months=0 WHERE id=?`, id); e != nil {
			return e
		}
	}
	return tx.Commit()
}
func (s *MySQLStore) ChangeVehicleOrganization(ctx context.Context, vehicleIDs []int64, orgID int64, _ bool) error {
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	for _, id := range vehicleIDs {
		if _, e = tx.ExecContext(ctx, `UPDATE vehicles SET organization_id=? WHERE id=?`, orgID, id); e != nil {
			return e
		}
		if _, e = tx.ExecContext(ctx, `UPDATE devices SET organization_id=? WHERE vehicle_id=?`, orgID, id); e != nil {
			return e
		}
	}
	return tx.Commit()
}
func (s *MySQLStore) DelegateVehicles(ctx context.Context, vehicleIDs []int64, orgID *int64, note string) error {
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	if orgID != nil {
		var status int
		if e = tx.QueryRowContext(ctx, `SELECT status FROM collection_companies WHERE id=? FOR UPDATE`, *orgID).Scan(&status); e != nil {
			if errors.Is(e, sql.ErrNoRows) {
				return errors.New("所选清收公司不存在")
			}
			return e
		}
		if status != 1 {
			return errors.New("所选清收公司未启用")
		}
	}
	for _, id := range vehicleIDs {
		if orgID == nil {
			_, e = tx.ExecContext(ctx, `UPDATE vehicles SET delegated_org_id=NULL,delegation_status='none',delegated_at=NULL,delegation_note='' WHERE id=?`, id)
		} else {
			_, e = tx.ExecContext(ctx, `UPDATE vehicles SET delegated_org_id=?,delegation_status='delegated',delegated_at=NOW(),delegation_note=? WHERE id=?`, orgID, note, id)
		}
		if e != nil {
			return e
		}
	}
	return tx.Commit()
}
func (s *MySQLStore) BatchDeleteVehicles(ctx context.Context, ids []int64, force bool) error {
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	for _, id := range ids {
		var count int
		_ = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM devices WHERE vehicle_id=?`, id).Scan(&count)
		if count > 0 {
			return errors.New("所选车辆存在绑定设备，请先解除绑定关系后再删除")
		}
		if _, e = tx.ExecContext(ctx, `DELETE FROM vehicle_records WHERE vehicle_id=?`, id); e != nil {
			return e
		}
		if _, e = tx.ExecContext(ctx, `DELETE FROM vehicles WHERE id=?`, id); e != nil {
			return e
		}
	}
	return tx.Commit()
}

func dateOrNil(v string) any {
	if v == "" {
		return nil
	}
	return v
}
func (s *MySQLStore) ListVehicleRecords(ctx context.Context) ([]VehicleRecord, error) {
	rows, e := s.db.QueryContext(ctx, `SELECT r.id,r.vehicle_id,v.plate_no,r.record_type,DATE_FORMAT(r.record_date,'%Y-%m-%d'),r.title,r.amount,r.mileage,r.status,r.detail,r.attachment_url,DATE_FORMAT(r.created_at,'%Y-%m-%d %H:%i:%s') FROM vehicle_records r JOIN vehicles v ON v.id=r.vehicle_id ORDER BY r.record_date DESC,r.id DESC`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []VehicleRecord{}
	for rows.Next() {
		var v VehicleRecord
		if e = rows.Scan(&v.ID, &v.VehicleID, &v.PlateNo, &v.RecordType, &v.RecordDate, &v.Title, &v.Amount, &v.Mileage, &v.Status, &v.Detail, &v.AttachmentURL, &v.CreatedAt); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *MySQLStore) SaveVehicleRecord(ctx context.Context, v VehicleRecord) (int64, error) {
	if v.ID == 0 {
		r, e := s.db.ExecContext(ctx, `INSERT INTO vehicle_records(vehicle_id,record_type,record_date,title,amount,mileage,status,detail,attachment_url) VALUES(?,?,?,?,?,?,?,?,?)`, v.VehicleID, v.RecordType, v.RecordDate, v.Title, v.Amount, v.Mileage, v.Status, v.Detail, v.AttachmentURL)
		if e != nil {
			return 0, e
		}
		return r.LastInsertId()
	}
	_, e := s.db.ExecContext(ctx, `UPDATE vehicle_records SET vehicle_id=?,record_type=?,record_date=?,title=?,amount=?,mileage=?,status=?,detail=?,attachment_url=? WHERE id=?`, v.VehicleID, v.RecordType, v.RecordDate, v.Title, v.Amount, v.Mileage, v.Status, v.Detail, v.AttachmentURL, v.ID)
	return v.ID, e
}
func (s *MySQLStore) DeleteVehicleRecord(ctx context.Context, id int64) error {
	_, e := s.db.ExecContext(ctx, `DELETE FROM vehicle_records WHERE id=?`, id)
	return e
}

func (s *MySQLStore) ListDevices(ctx context.Context) ([]Device, error) {
	rows, e := s.db.QueryContext(ctx, `SELECT d.id,d.device_no,d.imei,d.model,d.protocol,d.sim_no,d.iccid,d.organization_id,COALESCE(o.name,''),d.vehicle_id,COALESCE(v.plate_no,''),d.install_position,d.installer,d.installer_phone,d.photo_vin,d.photo_position,d.photo_vehicle,d.bound_vehicle_vin,COALESCE(DATE_FORMAT(d.bound_at,'%Y-%m-%d %H:%i:%s'),''),COALESCE(DATE_FORMAT(d.service_start_time,'%Y-%m-%d %H:%i:%s'),''),COALESCE(DATE_FORMAT(d.service_end_time,'%Y-%m-%d %H:%i:%s'),''),d.service_duration_months,d.device_key,d.inventory_status,COALESCE(DATE_FORMAT(d.inbound_date,'%Y-%m-%d'),''),d.remark,DATE_FORMAT(d.created_at,'%Y-%m-%d %H:%i:%s') FROM devices d LEFT JOIN organizations o ON o.id=d.organization_id LEFT JOIN vehicles v ON v.id=d.vehicle_id ORDER BY d.id DESC`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Device{}
	for rows.Next() {
		var v Device
		var org sql.NullInt64
		var vehicle sql.NullInt64
		if e = rows.Scan(&v.ID, &v.DeviceNo, &v.IMEI, &v.Model, &v.Protocol, &v.SIMNo, &v.ICCID, &org, &v.OrganizationName, &vehicle, &v.PlateNo, &v.InstallPosition, &v.Installer, &v.InstallerPhone, &v.PhotoVIN, &v.PhotoPosition, &v.PhotoVehicle, &v.BoundVehicleVIN, &v.BoundAt, &v.ServiceStartTime, &v.ServiceEndTime, &v.ServiceDurationMonths, &v.DeviceKey, &v.InventoryStatus, &v.InboundDate, &v.Remark, &v.CreatedAt); e != nil {
			return nil, e
		}
		if org.Valid {
			v.OrganizationID = &org.Int64
		}
		if vehicle.Valid {
			v.VehicleID = &vehicle.Int64
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *MySQLStore) SaveDevice(ctx context.Context, v Device) (int64, error) {
	var date any = nil
	if v.InboundDate != "" {
		date = v.InboundDate
	}
	if v.ID == 0 {
		if v.InventoryStatus == "" {
			v.InventoryStatus = "pending_production"
		}
		r, e := s.db.ExecContext(ctx, `INSERT INTO devices(device_no,imei,model,protocol,sim_no,iccid,organization_id,device_key,inventory_status,inbound_date,remark) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, v.DeviceNo, v.IMEI, v.Model, v.Protocol, v.SIMNo, v.ICCID, v.OrganizationID, v.DeviceKey, v.InventoryStatus, date, v.Remark)
		if e != nil {
			return 0, e
		}
		return r.LastInsertId()
	}
	// 状态只能经由专用状态确认接口流转，普通资料编辑不得绕过状态机。
	_, e := s.db.ExecContext(ctx, `UPDATE devices SET device_no=?,imei=?,model=?,protocol=?,sim_no=?,iccid=?,organization_id=?,device_key=?,remark=? WHERE id=?`, v.DeviceNo, v.IMEI, v.Model, v.Protocol, v.SIMNo, v.ICCID, v.OrganizationID, v.DeviceKey, v.Remark, v.ID)
	return v.ID, e
}
func (s *MySQLStore) DeleteDevice(ctx context.Context, id int64) error {
	var vehicleID sql.NullInt64
	var status string
	if err := s.db.QueryRowContext(ctx, `SELECT vehicle_id,inventory_status FROM devices WHERE id=?`, id).Scan(&vehicleID, &status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("设备不存在")
		}
		return err
	}
	if vehicleID.Valid {
		return errors.New("设备已绑定车辆，请先解除绑定关系后再删除")
	}
	if status == "in_use" || status == "pending_confirmation" || status == "pending_repair" {
		return errors.New("使用中、待确认或待维修设备不能删除")
	}
	var count int
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM device_maintenance WHERE device_id=?`, id).Scan(&count)
	if count > 0 {
		return errors.New("该设备存在运维记录，不能直接删除")
	}
	_, e := s.db.ExecContext(ctx, `DELETE FROM devices WHERE id=?`, id)
	return e
}

func (s *MySQLStore) TransitionDeviceStatus(ctx context.Context, id int64, target string) error {
	var current string
	if err := s.db.QueryRowContext(ctx, `SELECT inventory_status FROM devices WHERE id=?`, id).Scan(&current); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("设备不存在")
		}
		return err
	}
	valid := map[string]bool{"pending_use": true, "pending_repair": true, "disabled": true, "scrapped": true}
	if !valid[target] {
		return errors.New("目标设备状态不正确")
	}
	if current == "in_use" {
		return errors.New("使用中的设备不能直接修改状态，请先从车辆解绑")
	}
	if current == target {
		return fmt.Errorf("不允许将设备状态从 %s 变更为 %s", current, target)
	}
	_, err := s.db.ExecContext(ctx, `UPDATE devices SET inventory_status=? WHERE id=?`, target, id)
	return err
}

func (s *MySQLStore) BatchTransitionDeviceStatus(ctx context.Context, ids []int64, target string) error {
	if len(ids) == 0 {
		return errors.New("请选择设备")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	valid := map[string]bool{"pending_use": true, "pending_repair": true, "disabled": true, "scrapped": true}
	if !valid[target] {
		return errors.New("目标设备状态不正确")
	}
	for _, id := range ids {
		var current string
		if err = tx.QueryRowContext(ctx, `SELECT inventory_status FROM devices WHERE id=? FOR UPDATE`, id).Scan(&current); err != nil {
			return err
		}
		if current == "in_use" {
			return fmt.Errorf("设备 %d 正在使用中，不能修改状态", id)
		}
		if current != target {
			if _, err = tx.ExecContext(ctx, `UPDATE devices SET inventory_status=? WHERE id=?`, target, id); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

func (s *MySQLStore) MigrateDevices(ctx context.Context, ids []int64, organizationID int64) error {
	if len(ids) == 0 || organizationID == 0 {
		return errors.New("请选择设备和目标机构")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, id := range ids {
		var vehicleID sql.NullInt64
		if err = tx.QueryRowContext(ctx, `SELECT vehicle_id FROM devices WHERE id=? FOR UPDATE`, id).Scan(&vehicleID); err != nil {
			return err
		}
		if vehicleID.Valid {
			return fmt.Errorf("设备 %d 已绑定车辆，无法单独迁移，请通过车辆变更机构处理", id)
		}
		if _, err = tx.ExecContext(ctx, `UPDATE devices SET organization_id=? WHERE id=?`, organizationID, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *MySQLStore) RenewDevices(ctx context.Context, ids []int64, durationMonths int) error {
	if len(ids) == 0 || durationMonths <= 0 {
		return errors.New("请选择设备和续费时长")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, id := range ids {
		var status string
		var vehicleID sql.NullInt64
		if err = tx.QueryRowContext(ctx, `SELECT inventory_status,vehicle_id FROM devices WHERE id=? FOR UPDATE`, id).Scan(&status, &vehicleID); err != nil {
			return err
		}
		if !vehicleID.Valid && status != "pending_use" {
			return fmt.Errorf("设备 %d 必须为已绑定设备或待使用设备才可续费", id)
		}
		if !vehicleID.Valid {
			if _, err = tx.ExecContext(ctx, `UPDATE devices SET service_duration_months=service_duration_months+? WHERE id=?`, durationMonths, id); err != nil {
				return err
			}
			continue
		}
		if _, err = tx.ExecContext(ctx, `UPDATE devices SET service_start_time=COALESCE(service_start_time,NOW()),service_end_time=DATE_ADD(IF(service_end_time IS NULL OR service_end_time<NOW(),NOW(),service_end_time),INTERVAL ? MONTH),service_duration_months=service_duration_months+?,inventory_status=IF(inventory_status='disabled','in_use',inventory_status) WHERE id=?`, durationMonths, durationMonths, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *MySQLStore) ListMaintenance(ctx context.Context) ([]Maintenance, error) {
	rows, e := s.db.QueryContext(ctx, `SELECT m.id,m.device_id,d.device_no,m.maintenance_type,m.issue_description,m.handling_result,m.handler,m.status,COALESCE(DATE_FORMAT(m.handled_at,'%Y-%m-%d %H:%i:%s'),''),DATE_FORMAT(m.created_at,'%Y-%m-%d %H:%i:%s') FROM device_maintenance m JOIN devices d ON d.id=m.device_id ORDER BY m.id DESC`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Maintenance{}
	for rows.Next() {
		var v Maintenance
		if e = rows.Scan(&v.ID, &v.DeviceID, &v.DeviceNo, &v.MaintenanceType, &v.IssueDescription, &v.HandlingResult, &v.Handler, &v.Status, &v.HandledAt, &v.CreatedAt); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *MySQLStore) SaveMaintenance(ctx context.Context, v Maintenance) (int64, error) {
	var handled any = nil
	if v.HandledAt != "" {
		handled = v.HandledAt
	}
	if v.ID == 0 {
		r, e := s.db.ExecContext(ctx, `INSERT INTO device_maintenance(device_id,maintenance_type,issue_description,handling_result,handler,status,handled_at) VALUES(?,?,?,?,?,?,?)`, v.DeviceID, v.MaintenanceType, v.IssueDescription, v.HandlingResult, v.Handler, v.Status, handled)
		if e != nil {
			return 0, e
		}
		return r.LastInsertId()
	}
	_, e := s.db.ExecContext(ctx, `UPDATE device_maintenance SET device_id=?,maintenance_type=?,issue_description=?,handling_result=?,handler=?,status=?,handled_at=? WHERE id=?`, v.DeviceID, v.MaintenanceType, v.IssueDescription, v.HandlingResult, v.Handler, v.Status, handled, v.ID)
	return v.ID, e
}
func (s *MySQLStore) DeleteMaintenance(ctx context.Context, id int64) error {
	_, e := s.db.ExecContext(ctx, `DELETE FROM device_maintenance WHERE id=?`, id)
	return e
}

func (s *MySQLStore) ListFinanceCompanies(ctx context.Context) ([]FinanceCompany, error) {
	rows, e := s.db.QueryContext(ctx, `SELECT id,name,code,contact_name,contact_phone,address,status,remark,DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s') FROM finance_companies ORDER BY id DESC`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []FinanceCompany{}
	for rows.Next() {
		var v FinanceCompany
		if e = rows.Scan(&v.ID, &v.Name, &v.Code, &v.ContactName, &v.ContactPhone, &v.Address, &v.Status, &v.Remark, &v.CreatedAt); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *MySQLStore) SaveFinanceCompany(ctx context.Context, v FinanceCompany) (int64, error) {
	v.Name = strings.TrimSpace(v.Name)
	v.Code = strings.TrimSpace(v.Code)
	if v.Name == "" || v.Code == "" {
		return 0, errors.New("公司名称和编码不能为空")
	}
	if v.ID == 0 {
		r, e := s.db.ExecContext(ctx, `INSERT INTO finance_companies(name,code,contact_name,contact_phone,address,status,remark) VALUES(?,?,?,?,?,?,?)`, v.Name, v.Code, v.ContactName, v.ContactPhone, v.Address, v.Status, v.Remark)
		if e != nil {
			return 0, e
		}
		return r.LastInsertId()
	}
	_, e := s.db.ExecContext(ctx, `UPDATE finance_companies SET name=?,code=?,contact_name=?,contact_phone=?,address=?,status=?,remark=? WHERE id=?`, v.Name, v.Code, v.ContactName, v.ContactPhone, v.Address, v.Status, v.Remark, v.ID)
	return v.ID, e
}
func (s *MySQLStore) DeleteFinanceCompany(ctx context.Context, id int64) error {
	var n int
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM finance_products WHERE company_id=?`, id).Scan(&n)
	if n > 0 {
		return errors.New("该金融公司存在金融产品，不能删除")
	}
	_, e := s.db.ExecContext(ctx, `DELETE FROM finance_companies WHERE id=?`, id)
	return e
}
func (s *MySQLStore) ListFinanceProducts(ctx context.Context) ([]FinanceProduct, error) {
	rows, e := s.db.QueryContext(ctx, `SELECT p.id,p.company_id,c.name,p.name,p.code,p.product_type,p.annual_rate,p.term_months,p.status,p.remark,DATE_FORMAT(p.created_at,'%Y-%m-%d %H:%i:%s') FROM finance_products p JOIN finance_companies c ON c.id=p.company_id ORDER BY p.id DESC`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []FinanceProduct{}
	for rows.Next() {
		var v FinanceProduct
		if e = rows.Scan(&v.ID, &v.CompanyID, &v.CompanyName, &v.Name, &v.Code, &v.ProductType, &v.AnnualRate, &v.TermMonths, &v.Status, &v.Remark, &v.CreatedAt); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *MySQLStore) SaveFinanceProduct(ctx context.Context, v FinanceProduct) (int64, error) {
	if strings.TrimSpace(v.Name) == "" || strings.TrimSpace(v.Code) == "" || v.CompanyID == 0 {
		return 0, errors.New("金融公司、产品名称和编码不能为空")
	}
	if v.ID == 0 {
		r, e := s.db.ExecContext(ctx, `INSERT INTO finance_products(company_id,name,code,product_type,annual_rate,term_months,status,remark) VALUES(?,?,?,?,?,?,?,?)`, v.CompanyID, v.Name, v.Code, v.ProductType, v.AnnualRate, v.TermMonths, v.Status, v.Remark)
		if e != nil {
			return 0, e
		}
		return r.LastInsertId()
	}
	_, e := s.db.ExecContext(ctx, `UPDATE finance_products SET company_id=?,name=?,code=?,product_type=?,annual_rate=?,term_months=?,status=?,remark=? WHERE id=?`, v.CompanyID, v.Name, v.Code, v.ProductType, v.AnnualRate, v.TermMonths, v.Status, v.Remark, v.ID)
	return v.ID, e
}
func (s *MySQLStore) DeleteFinanceProduct(ctx context.Context, id int64) error {
	_, e := s.db.ExecContext(ctx, `DELETE FROM finance_products WHERE id=?`, id)
	return e
}
func (s *MySQLStore) ListCollectionCompanies(ctx context.Context) ([]CollectionCompany, error) {
	rows, e := s.db.QueryContext(ctx, `SELECT id,name,code,contact_name,contact_phone,service_area,address,status,remark,DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s') FROM collection_companies ORDER BY id DESC`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []CollectionCompany{}
	for rows.Next() {
		var v CollectionCompany
		if e = rows.Scan(&v.ID, &v.Name, &v.Code, &v.ContactName, &v.ContactPhone, &v.ServiceArea, &v.Address, &v.Status, &v.Remark, &v.CreatedAt); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *MySQLStore) SaveCollectionCompany(ctx context.Context, v CollectionCompany) (int64, error) {
	if strings.TrimSpace(v.Name) == "" || strings.TrimSpace(v.Code) == "" {
		return 0, errors.New("公司名称和编码不能为空")
	}
	if v.ID == 0 {
		r, e := s.db.ExecContext(ctx, `INSERT INTO collection_companies(name,code,contact_name,contact_phone,service_area,address,status,remark) VALUES(?,?,?,?,?,?,?,?)`, v.Name, v.Code, v.ContactName, v.ContactPhone, v.ServiceArea, v.Address, v.Status, v.Remark)
		if e != nil {
			return 0, e
		}
		return r.LastInsertId()
	}
	_, e := s.db.ExecContext(ctx, `UPDATE collection_companies SET name=?,code=?,contact_name=?,contact_phone=?,service_area=?,address=?,status=?,remark=? WHERE id=?`, v.Name, v.Code, v.ContactName, v.ContactPhone, v.ServiceArea, v.Address, v.Status, v.Remark, v.ID)
	return v.ID, e
}
func (s *MySQLStore) DeleteCollectionCompany(ctx context.Context, id int64) error {
	_, e := s.db.ExecContext(ctx, `DELETE FROM collection_companies WHERE id=?`, id)
	return e
}
