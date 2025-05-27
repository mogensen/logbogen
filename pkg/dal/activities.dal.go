package dal

import (
	"time"

	"github.com/google/uuid"
	"github.com/mogensen/logbook/pkg/database"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ClimbingActivity struct defines the Climbing Activity Model
type ClimbingActivity struct {
	gorm.Model
	ID           uuid.UUID                   `gorm:"not null"`
	User         *uint64                     `gorm:"index,not null"`
	Date         time.Time                   `gorm:"not null"`
	Lat          float64                     `gorm:"not null"`
	Lng          float64                     `gorm:"not null"`
	Location     string                      `gorm:"not null"`
	Type         string                      `gorm:"not null"`
	OtherType    string                      `gorm:"not null default ''"`
	Role         string                      `gorm:"not null"`
	Comment      string                      `gorm:"not null"`
	Participants datatypes.JSONSlice[uint64] `gorm:"type:json"`
	CreatedAt    time.Time                   `gorm:"not null"`
	UpdatedAt    time.Time                   `gorm:"not null"`
}

// CreateActivity create a Activity entry in the Activity's table
func CreateClimbingActivity(activity *ClimbingActivity) *gorm.DB {
	return database.DB.Create(activity)
}

// FindClimbingActivity finds a ClimbingActivity with given condition
func FindClimbingActivity(dest *ClimbingActivity, conds ...interface{}) *gorm.DB {
	return database.DB.Model(&ClimbingActivity{}).Take(dest, conds...)
}

// FindClimbingActivityByUser finds a ClimbingActivity with given ClimbingActivity and user identifier
func FindClimbingActivityToClone(dest *ClimbingActivity, ClimbingActivityIden string, userIden uint64) *gorm.DB {
	return database.DB.Model(&ClimbingActivity{}).Order("date DESC").
		Where(datatypes.JSONArrayQuery("participants").Contains(userIden)).
		Take(dest, "id = ? AND user != ?", ClimbingActivityIden, userIden)
}

// FindClimbingActivityByUser finds a ClimbingActivity with given ClimbingActivity and user identifier
func FindClimbingActivityByUser(dest *ClimbingActivity, ClimbingActivityIden string, userIden uint64) *gorm.DB {
	return FindClimbingActivity(dest, "id = ? AND user = ?", ClimbingActivityIden, userIden)
}

// FindClimbingActivitiesByUser finds the ClimbingActivitys with user's identifier given
func FindClimbingActivitiesByUser(dest *[]ClimbingActivity, userIden uint64) *gorm.DB {
	return database.DB.Model(&ClimbingActivity{}).Order("date DESC").Find(dest, "user = ?", userIden)
}

// DeleteClimbingActivity deletes a ClimbingActivity from ClimbingActivitys' table with the given ClimbingActivity and user identifier
func DeleteClimbingActivity(ClimbingActivityIden string, userIden uint64) *gorm.DB {
	return database.DB.Unscoped().Delete(&ClimbingActivity{}, "id = ? AND user = ?", ClimbingActivityIden, userIden)
}

// UpdateClimbingActivity allows to update the ClimbingActivity with the given ClimbingActivityID and userID
func UpdateClimbingActivity(ClimbingActivityIden string, userIden uint64, data interface{}) *gorm.DB {
	return database.DB.Model(&ClimbingActivity{}).Where("id = ? AND user = ?", ClimbingActivityIden, userIden).Updates(data)
}

// FindPendingActivitiesByUser finds any activities created by another user but with the current user as participant
func FindPendingActivitiesByUser(dest *[]ClimbingActivity, userIden uint64) *gorm.DB {
	return database.DB.Model(&ClimbingActivity{}).Order("date DESC").
		Where(datatypes.JSONArrayQuery("participants").Contains(userIden)).
		Find(dest, "user != ?", userIden)
}
