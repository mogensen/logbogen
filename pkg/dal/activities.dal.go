package dal

import (
	"database/sql/driver"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mogensen/logbook/pkg/database"
	"gorm.io/gorm"
)

type Participants []uint64

func (o *Participants) Scan(src any) error {
	bytes, ok := src.([]byte)
	if !ok {
		return errors.New("src value cannot cast to []byte")
	}
	ids := strings.SplitSeq(string(bytes), ",")
	for v := range ids {
		if v == "" {
			continue
		}
		id, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return err
		}
		*o = append(*o, id)
	}

	return nil
}

func (o Participants) Value() (driver.Value, error) {
	if len(o) == 0 {
		return nil, nil
	}
	strs := make([]string, len(o))
	for i, v := range o {
		strs[i] = strconv.FormatUint(v, 10)
	}
	return strings.Join(strs, ","), nil
}

// ClimbingActivity struct defines the Climbing Activity Model
type ClimbingActivity struct {
	gorm.Model
	ID           uuid.UUID    `gorm:"not null"`
	User         *uint        `gorm:"index,not null"`
	Date         time.Time    `gorm:"not null"`
	Lat          float64      `gorm:"not null"`
	Lng          float64      `gorm:"not null"`
	Location     string       `gorm:"not null"`
	Type         string       `gorm:"not null"`
	OtherType    string       `gorm:"not null default ''"`
	Role         string       `gorm:"not null"`
	Comment      string       `gorm:"not null"`
	Participants Participants `gorm:"type:VARCHAR(1024)"`
	CreatedAt    time.Time    `gorm:"not null"`
	UpdatedAt    time.Time    `gorm:"not null"`
}

// CreateActivity create a Activity entry in the Activity's table
func CreateClimbingActivity(activity *ClimbingActivity) *gorm.DB {
	return database.DB.Create(activity)
}

// FindClimbingActivity finds a ClimbingActivity with given condition
func FindClimbingActivity(dest interface{}, conds ...interface{}) *gorm.DB {
	return database.DB.Model(&ClimbingActivity{}).Take(dest, conds...)
}

// FindClimbingActivityByUser finds a ClimbingActivity with given ClimbingActivity and user identifier
func FindClimbingActivityByUser(dest interface{}, ClimbingActivityIden interface{}, userIden interface{}) *gorm.DB {
	return FindClimbingActivity(dest, "id = ? AND user = ?", ClimbingActivityIden, userIden)
}

// FindClimbingActivitiesByUser finds the ClimbingActivitys with user's identifier given
func FindClimbingActivitiesByUser(dest interface{}, userIden interface{}) *gorm.DB {
	return database.DB.Model(&ClimbingActivity{}).Order("date DESC").Find(dest, "user = ?", userIden)
}

// DeleteClimbingActivity deletes a ClimbingActivity from ClimbingActivitys' table with the given ClimbingActivity and user identifier
func DeleteClimbingActivity(ClimbingActivityIden interface{}, userIden interface{}) *gorm.DB {
	return database.DB.Unscoped().Delete(&ClimbingActivity{}, "id = ? AND user = ?", ClimbingActivityIden, userIden)
}

// UpdateClimbingActivity allows to update the ClimbingActivity with the given ClimbingActivityID and userID
func UpdateClimbingActivity(ClimbingActivityIden interface{}, userIden interface{}, data interface{}) *gorm.DB {
	return database.DB.Model(&ClimbingActivity{}).Where("id = ? AND user = ?", ClimbingActivityIden, userIden).Updates(data)
}
