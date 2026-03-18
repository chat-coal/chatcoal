package services

import (
	"chatcoal/database"
	"chatcoal/models"
	"errors"

	"gorm.io/gorm"
)

var validTargetTypes = map[string]bool{
	"user":       true,
	"message":    true,
	"dm_message": true,
	"forum_post": true,
}

func CreateReport(reporterID models.Snowflake, targetType string, targetID models.Snowflake, reason string) (*models.Report, error) {
	if !validTargetTypes[targetType] {
		return nil, errors.New("invalid target type")
	}

	// Verify the target exists
	switch targetType {
	case "user":
		var count int64
		database.Database.Model(&models.User{}).Where("id = ?", targetID).Count(&count)
		if count == 0 {
			return nil, errors.New("user not found")
		}
	case "message":
		var count int64
		database.Database.Model(&models.Message{}).Where("id = ?", targetID).Count(&count)
		if count == 0 {
			return nil, errors.New("message not found")
		}
	case "dm_message":
		var count int64
		database.Database.Model(&models.DMMessage{}).Where("id = ?", targetID).Count(&count)
		if count == 0 {
			return nil, errors.New("message not found")
		}
	case "forum_post":
		var count int64
		database.Database.Model(&models.ForumPost{}).Where("id = ?", targetID).Count(&count)
		if count == 0 {
			return nil, errors.New("forum post not found")
		}
	}

	// Check for duplicate report
	var existing int64
	database.Database.Model(&models.Report{}).
		Where("reporter_id = ? AND target_type = ? AND target_id = ?", reporterID, targetType, targetID).
		Count(&existing)
	if existing > 0 {
		return nil, errors.New("already reported")
	}

	report := models.Report{
		ReporterID: reporterID,
		TargetType: targetType,
		TargetID:   targetID,
		Reason:     reason,
	}
	if err := database.Database.Create(&report).Error; err != nil {
		return nil, err
	}
	return &report, nil
}

type ReportWithReporter struct {
	models.Report
	Reporter *models.User `json:"reporter,omitempty" gorm:"foreignKey:ReporterID"`
}

func GetReports(status string, before models.Snowflake) ([]ReportWithReporter, error) {
	var reports []ReportWithReporter
	q := database.Database.Model(&models.Report{}).
		Preload("Reporter", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, display_name, username, avatar_url")
		}).
		Order("created_at DESC").
		Limit(50)

	if status != "" {
		q = q.Where("status = ?", status)
	}
	if before != 0 {
		q = q.Where("id < ?", before)
	}

	err := q.Find(&reports).Error
	return reports, err
}

func UpdateReportStatus(reportID models.Snowflake, status string) (*models.Report, error) {
	var report models.Report
	if err := database.Database.First(&report, "id = ?", reportID).Error; err != nil {
		return nil, err
	}
	report.Status = status
	if err := database.Database.Save(&report).Error; err != nil {
		return nil, err
	}
	return &report, nil
}
