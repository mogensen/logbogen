package dal

import (
	"time"

	"github.com/google/uuid"
	"github.com/mogensen/logbook/pkg/database"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Activity struct defines the Activity Model
type Activity struct {
	gorm.Model
	ID           uuid.UUID                   `gorm:"not null"`
	User         *uint64                     `gorm:"index,not null"`
	Date         time.Time                   `gorm:"not null"`
	Lat          float64                     `gorm:"not null"`
	Lng          float64                     `gorm:"not null"`
	Location     string                      `gorm:"not null"`
	Type         string                      `gorm:"not null"`
	OtherType    string                      `gorm:"not null default ''"`
	Category     string                      `gorm:"not null"`
	Role         string                      `gorm:"not null"`
	Comment      string                      `gorm:"not null"`
	Participants datatypes.JSONSlice[uint64] `gorm:"type:json"`
	CreatedAt    time.Time                   `gorm:"not null"`
	UpdatedAt    time.Time                   `gorm:"not null"`
}

// CreateActivity creates an Activity entry in the Activity's table
func CreateActivity(activity *Activity) *gorm.DB {
	return database.DB.Create(activity)
}

// FindActivity finds an Activity with given condition
func FindActivity(dest *Activity, conds ...interface{}) *gorm.DB {
	return database.DB.Model(&Activity{}).Take(dest, conds...)
}

// FindActivityToClone finds an Activity with given Activity and user identifier for cloning
func FindActivityToClone(dest *Activity, ActivityIden string, userIden uint64) *gorm.DB {
	return database.DB.Model(&Activity{}).Order("date DESC").
		Where(datatypes.JSONArrayQuery("participants").Contains(userIden)).
		Take(dest, "id = ? AND user != ?", ActivityIden, userIden)
}

// FindActivityByUser finds an Activity with given Activity and user identifier
func FindActivityByUser(dest *Activity, ActivityIden string, userIden uint64) *gorm.DB {
	return FindActivity(dest, "id = ? AND user = ?", ActivityIden, userIden)
}

// FindActivitiesByUser finds the Activities with user's identifier given
func FindActivitiesByUser(dest *[]Activity, userIden uint64) *gorm.DB {
	return database.DB.Model(&Activity{}).Order("date DESC").Find(dest, "user = ?", userIden)
}

// DeleteActivity deletes an Activity from Activities' table with the given Activity and user identifier
func DeleteActivity(ActivityIden string, userIden uint64) *gorm.DB {
	return database.DB.Unscoped().Delete(&Activity{}, "id = ? AND user = ?", ActivityIden, userIden)
}

// UpdateActivity allows to update the Activity with the given ActivityID and userID
func UpdateActivity(ActivityIden string, userIden uint64, data interface{}) *gorm.DB {
	return database.DB.Model(&Activity{}).Where("id = ? AND user = ?", ActivityIden, userIden).Updates(data)
}

// FindPendingActivitiesByUser finds any activities created by another user but with the current user as participant
func FindPendingActivitiesByUser(dest *[]Activity, userIden uint64) *gorm.DB {
	return database.DB.Model(&Activity{}).Order("date DESC").
		Where(datatypes.JSONArrayQuery("participants").Contains(userIden)).
		Find(dest, "user != ?", userIden)
}
