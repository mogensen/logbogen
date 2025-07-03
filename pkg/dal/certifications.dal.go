package dal

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Certification struct defines the Certification Model
// OtherParticipants is stored as a comma-separated string
// You may want to use a JSON array for more flexibility
// Add gorm.Model for ID, CreatedAt, UpdatedAt, DeletedAt
// If you want to use your own ID, you can specify it

type Certification struct {
	ID           uuid.UUID `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	UserID       *uint64                     `gorm:"not null"`
	Category     string                      `gorm:"not null"`
	Type         string                      `gorm:"not null"`
	Provider     string                      `gorm:"not null"`
	StartDate    time.Time                   `gorm:"not null"`
	EndDate      time.Time                   `gorm:"not null"`
	Participants datatypes.JSONSlice[uint64] `gorm:"type:json"`
}

// CertificationService handles all database operations for certifications
// Similar to ActivityService

type CertificationService struct {
	db *gorm.DB
}

// NewCertificationService creates a new CertificationService instance
func NewCertificationService(db *gorm.DB) *CertificationService {
	return &CertificationService{db: db}
}

// CreateCertification inserts a new certification for a user
func (s *CertificationService) CreateCertification(cert Certification) (Certification, error) {
	result := s.db.Create(&cert)
	return cert, result.Error
}

func (s *CertificationService) GetCertification(userID uint64, certID uuid.UUID) (Certification, error) {
	var cert Certification
	result := s.db.First(&cert, "id = ? AND user_id = ?", certID, userID)
	return cert, result.Error
}

// GetCertificationsByUser returns all certifications for a user
func (s *CertificationService) GetCertificationsByUser(userID uint64) ([]Certification, error) {
	var certs []Certification
	result := s.db.Order("start_date DESC").Find(&certs, "user_id = ?", userID)
	return certs, result.Error
}

// UpdateCertification updates an existing certification
func (s *CertificationService) UpdateCertification(userID uint64, cert Certification) (Certification, error) {
	result := s.db.Model(&Certification{}).Where("id = ? AND user_id = ?", cert.ID, userID).Updates(&cert)
	if result.Error != nil {
		return cert, result.Error
	}
	result = s.db.First(&cert, "id = ? AND user_id = ?", cert.ID, userID)
	return cert, result.Error
}

// DeleteCertification deletes a certification by ID and user
func (s *CertificationService) DeleteCertification(certID uuid.UUID, userID uint64) error {
	result := s.db.Unscoped().Delete(&Certification{}, "id = ? AND user_id = ?", certID, userID)
	return result.Error
}
