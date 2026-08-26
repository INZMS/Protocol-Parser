package basicinfo

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

// StartServiceJobs starts the device-service automation owned by the backend.
// Expiration runs shortly after midnight; renewal reminders run at 10:00 local time.
func (s *MySQLStore) StartServiceJobs(ctx context.Context) {
	reminderDays := 30
	if value, err := strconv.Atoi(os.Getenv("DEVICE_SERVICE_REMINDER_DAYS")); err == nil && value > 0 {
		reminderDays = value
	}
	// 启动时先补跑一次，避免服务停机期间错过到期或提醒时间点。
	go func() {
		jobCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()
		if err := s.disableExpiredDevices(jobCtx); err != nil {
			log.Printf("设备到期停用补跑失败: %v", err)
		}
		if err := s.createRenewalReminders(jobCtx, reminderDays); err != nil {
			log.Printf("设备续费提醒补跑失败: %v", err)
		}
	}()
	go s.runDaily(ctx, 0, 5, func(jobCtx context.Context) error { return s.disableExpiredDevices(jobCtx) })
	go s.runDaily(ctx, 10, 0, func(jobCtx context.Context) error { return s.createRenewalReminders(jobCtx, reminderDays) })
}

func (s *MySQLStore) runDaily(ctx context.Context, hour, minute int, job func(context.Context) error) {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
		if !next.After(now) {
			next = next.AddDate(0, 0, 1)
		}
		timer := time.NewTimer(time.Until(next))
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			jobCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
			if err := job(jobCtx); err != nil {
				log.Printf("设备服务自动任务执行失败: %v", err)
			}
			cancel()
		}
	}
}

func (s *MySQLStore) disableExpiredDevices(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `UPDATE devices SET inventory_status='disabled' WHERE inventory_status='in_use' AND service_end_time IS NOT NULL AND service_end_time<CURRENT_TIMESTAMP`)
	return err
}

func (s *MySQLStore) createRenewalReminders(ctx context.Context, reminderDays int) error {
	_, err := s.db.ExecContext(ctx, `INSERT IGNORE INTO system_notifications(organization_id,device_id,type,title,content,notification_key)
		SELECT d.organization_id,d.id,'device_renewal','设备服务续费提醒',
		CONCAT('【续费提醒】贵机构名下的设备（设备号：',d.device_no,'）将于 ',DATE_FORMAT(d.service_end_time,'%Y-%m-%d %H:%i:%s'),' 到期。为保证服务不中断，请及时续费。'),
		CONCAT('device-renewal:',d.id,':',DATE_FORMAT(d.service_end_time,'%Y%m%d%H%i%s'))
		FROM devices d WHERE d.inventory_status='in_use' AND d.service_end_time>=CURRENT_TIMESTAMP
		AND d.service_end_time<DATE_ADD(CURRENT_TIMESTAMP,INTERVAL ? DAY)`, reminderDays)
	if err != nil {
		return fmt.Errorf("生成设备续费提醒失败: %w", err)
	}
	return nil
}
