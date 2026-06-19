package dal

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

func TestCreateAndGetCertification(t *testing.T) {
	db := setupTestDB(t)
	svc := NewCertificationService(db)

	userID := uint64(1)
	cat := "climbing"
	typeID := "rock"
	cert := Certification{
		ID:           uuid.New(),
		UserID:       &userID,
		Provider:     "Falk",
		Category:     cat,
		Type:         typeID,
		StartDate:    time.Now(),
		EndDate:      time.Now().AddDate(1, 0, 0),
		Participants: datatypes.JSONSlice[uint64]{userID},
	}
	created, err := svc.CreateCertification(cert)
	if err != nil {
		t.Fatalf("CreateCertification failed: %v", err)
	}
	certs, err := svc.GetCertificationsByUser(1)
	if err != nil {
		t.Fatalf("GetCertificationsByUser failed: %v", err)
	}
	if len(certs) != 1 {
		t.Fatalf("expected 1 cert, got %d", len(certs))
	}
	if certs[0].Provider != "Falk" {
		t.Errorf("expected provider Falk, got %s", certs[0].Provider)
	}
	_ = created
}

func TestUpdateCertification(t *testing.T) {
	db := setupTestDB(t)
	svc := NewCertificationService(db)

	userID := uint64(1)
	cat := "climbing"
	typeID := "rock"
	cert := Certification{
		ID:           uuid.New(),
		UserID:       &userID,
		Provider:     "Falk",
		Category:     cat,
		Type:         typeID,
		StartDate:    time.Now(),
		EndDate:      time.Now().AddDate(1, 0, 0),
		Participants: datatypes.JSONSlice[uint64]{userID},
	}
	created, err := svc.CreateCertification(cert)
	if err != nil {
		t.Fatalf("CreateCertification failed: %v", err)
	}
	created.Provider = "Dansk Træklatrenævn"
	updated, err := svc.UpdateCertification(1, created)
	if err != nil {
		t.Fatalf("UpdateCertification failed: %v", err)
	}
	certs, _ := svc.GetCertificationsByUser(1)
	if certs[0].Provider != "Dansk Træklatrenævn" {
		t.Errorf("expected updated provider, got %s", certs[0].Provider)
	}
	_ = updated
}

func TestDeleteCertification(t *testing.T) {
	db := setupTestDB(t)
	svc := NewCertificationService(db)

	userID := uint64(1)
	cat := "climbing"
	typeID := "rock"
	cert := Certification{
		ID:           uuid.New(),
		UserID:       &userID,
		Provider:     "Falk",
		Category:     cat,
		Type:         typeID,
		StartDate:    time.Now(),
		EndDate:      time.Now().AddDate(1, 0, 0),
		Participants: datatypes.JSONSlice[uint64]{userID},
	}
	created, err := svc.CreateCertification(cert)
	if err != nil {
		t.Fatalf("CreateCertification failed: %v", err)
	}
	err = svc.DeleteCertification(created.ID, 1)
	if err != nil {
		t.Fatalf("DeleteCertification failed: %v", err)
	}
	certs, _ := svc.GetCertificationsByUser(1)
	if len(certs) != 0 {
		t.Errorf("expected 0 certs, got %d", len(certs))
	}
}
