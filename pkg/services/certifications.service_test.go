package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mogensen/logbook/pkg/dal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDAL(t *testing.T) *dal.CertificationService {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	if err := db.AutoMigrate(&dal.Certification{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return dal.NewCertificationService(db)
}

func TestCertificationService_CRUD(t *testing.T) {
	dalSvc := setupTestDAL(t)
	svc := NewCertificationService(dalSvc)
	userID := uint64(1)
	cert := dal.Certification{
		ID:                uuid.New(),
		UserID:            &userID,
		Provider:          "Falk",
		StartDate:         time.Now(),
		EndDate:           time.Now().AddDate(1, 0, 0),
		OtherParticipants: "Alice,Bob",
	}
	// Create
	created, err := svc.CreateCertification(&cert)
	if err != nil {
		t.Fatalf("CreateCertification failed: %v", err)
	}
	// List
	certs, err := svc.ListUserCertifications(1)
	if err != nil || len(certs) != 1 {
		t.Fatalf("ListUserCertifications failed: %v", err)
	}
	// Update
	update := map[string]interface{}{"Provider": "Dansk Træklatrenævn"}
	updated, err := svc.UpdateCertification(created.ID, 1, update)
	if err != nil {
		t.Fatalf("UpdateCertification failed: %v", err)
	}
	if updated.Provider != "Dansk Træklatrenævn" {
		t.Errorf("expected updated provider, got %s", updated.Provider)
	}
	// Delete
	err = svc.DeleteCertification(created.ID, 1)
	if err != nil {
		t.Fatalf("DeleteCertification failed: %v", err)
	}
	certs, _ = svc.ListUserCertifications(1)
	if len(certs) != 0 {
		t.Errorf("expected 0 certs, got %d", len(certs))
	}
}
