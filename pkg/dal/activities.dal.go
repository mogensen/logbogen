package dal

import (
	"time"

	"github.com/google/uuid"
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
	Category     string                      `gorm:"not null"`
	Role         string                      `gorm:"not null"`
	Comment      string                      `gorm:"not null"`
	Participants datatypes.JSONSlice[uint64] `gorm:"type:json"`
	CreatedAt    time.Time                   `gorm:"not null"`
	UpdatedAt    time.Time                   `gorm:"not null"`
}

type ActivityDal interface {
	CreateActivity(activity *Activity) (Activity, error)
	FindActivity(conds ...interface{}) (Activity, error)
	FindActivityToClone(ActivityIden string, userIden uint64) (Activity, error)
	FindActivityByUser(ActivityIden string, userIden uint64) (Activity, error)
	FindActivitiesByUser(userIden uint64) ([]Activity, error)
	DeleteActivity(ActivityIden string, userIden uint64) error
	UpdateActivity(ActivityIden string, userIden uint64, data interface{}) (Activity, error)
	FindPendingActivitiesByUser(userIden uint64) ([]Activity, error)
}

// ActivityService handles all database operations for activities
type ActivityService struct {
	db *gorm.DB
}

// NewActivityService creates a new ActivityService instance
func NewActivityService(db *gorm.DB) *ActivityService {
	return &ActivityService{db: db}
}

// CreateActivity creates an Activity entry in the Activity's table
func (s *ActivityService) CreateActivity(activity *Activity) (Activity, error) {
	result := s.db.Create(activity)
	return *activity, result.Error
}

// FindActivity finds an Activity with given condition
func (s *ActivityService) FindActivity(conds ...interface{}) (Activity, error) {
	var activity Activity
	result := s.db.Model(&Activity{}).Take(&activity, conds...)
	return activity, result.Error
}

// FindActivityToClone finds an Activity with given Activity and user identifier for cloning
func (s *ActivityService) FindActivityToClone(ActivityIden string, userIden uint64) (Activity, error) {
	var activity Activity
	result := s.db.Model(&Activity{}).Order("date DESC").
		Where(datatypes.JSONArrayQuery("participants").Contains(userIden)).
		Take(&activity, "id = ? AND user != ?", ActivityIden, userIden)
	return activity, result.Error
}

// FindActivityByUser finds an Activity with given Activity and user identifier
func (s *ActivityService) FindActivityByUser(ActivityIden string, userIden uint64) (Activity, error) {
	return s.FindActivity("id = ? AND user = ?", ActivityIden, userIden)
}

// FindActivitiesByUser finds the Activities with user's identifier given
func (s *ActivityService) FindActivitiesByUser(userIden uint64) ([]Activity, error) {
	var activities []Activity
	result := s.db.Model(&Activity{}).Order("date DESC").Find(&activities, "user = ?", userIden)
	return activities, result.Error
}

// DeleteActivity deletes an Activity from Activities' table with the given Activity and user identifier
func (s *ActivityService) DeleteActivity(ActivityIden string, userIden uint64) error {
	result := s.db.Unscoped().Delete(&Activity{}, "id = ? AND user = ?", ActivityIden, userIden)
	return result.Error
}

// UpdateActivity allows to update the Activity with the given ActivityID and userID
func (s *ActivityService) UpdateActivity(ActivityIden string, userIden uint64, data interface{}) (Activity, error) {
	result := s.db.Model(&Activity{}).Where("id = ? AND user = ?", ActivityIden, userIden).Updates(data)
	if result.Error != nil {
		return Activity{}, result.Error
	}
	return s.FindActivityByUser(ActivityIden, userIden)
}

// FindPendingActivitiesByUser finds any activities created by another user but with the current user as participant
func (s *ActivityService) FindPendingActivitiesByUser(userIden uint64) ([]Activity, error) {
	var activities []Activity
	result := s.db.Model(&Activity{}).Order("date DESC").
		Where(datatypes.JSONArrayQuery("participants").Contains(userIden)).
		Find(&activities, "user != ?", userIden)
	return activities, result.Error
}
