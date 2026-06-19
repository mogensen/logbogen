package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mogensen/logbook/pkg/config"
	"github.com/mogensen/logbook/pkg/dal"
	"github.com/mogensen/logbook/pkg/mocks"
	"github.com/mogensen/logbook/pkg/types"
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
	userDalMock := &mocks.UserDalMock{}
	userID := uint64(1)
	cat := config.AllCertificationCategories[0]
	typeObj := config.AllCertificationTypes[0]
	userDalMock.On("FindUsers").
		Return([]dal.User{{Model: gorm.Model{ID: uint(userID)}, Name: "Test User", Email: "test@example.com"}}, nil)
	svc := NewCertificationService(dalSvc, userDalMock)
	cert := types.Certification{
		ID:              uuid.New(),
		UserID:          &userID,
		CategoryID:      cat.ID,
		Category:        cat,
		TypeID:          typeObj.ID,
		Type:            typeObj,
		Provider:        "Falk",
		StartDate:       types.Date(time.Now()),
		EndDate:         types.Date(time.Now().AddDate(1, 0, 0)),
		ParticipantsIDs: []uint64{userID},
	}
	// Create
	created, err := svc.CreateCertification(&cert)
	if err != nil {
		t.Fatalf("CreateCertification failed: %v", err)
	}
	// List
	certs, err := svc.ListUserCertifications(userID)
	if err != nil || len(certs) != 1 {
		t.Fatalf("ListUserCertifications failed: %v", err)
	}
	// Update
	created.Provider = "Dansk Træklatrenævn"
	updated, err := svc.UpdateCertification(userID, &created)
	if err != nil {
		t.Fatalf("UpdateCertification failed: %v", err)
	}
	if updated.Provider != "Dansk Træklatrenævn" {
		t.Errorf("expected updated provider, got %s", updated.Provider)
	}
	// Delete
	err = svc.DeleteCertification(created.ID, userID)
	if err != nil {
		t.Fatalf("DeleteCertification failed: %v", err)
	}
	certs, _ = svc.ListUserCertifications(userID)
	if len(certs) != 0 {
		t.Errorf("expected 0 certs, got %d", len(certs))
	}
}
