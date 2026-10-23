package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mogensen/logbook/pkg/config"
	"github.com/mogensen/logbook/pkg/dal"
	"github.com/mogensen/logbook/pkg/mocks"
	"github.com/mogensen/logbook/pkg/types"
	"github.com/stretchr/testify/require"
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
	userDalMock.On("FindUsersByIDs", []uint64{2}).
		Return([]dal.User{{ID: 2, Name: "Test User", Email: "test@example.com"}}, nil)
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
		ParticipantsIDs: []uint64{userID, 2},
	}
	// Create
	created, err := svc.CreateCertification(&cert)
	if err != nil {
		t.Fatalf("CreateCertification failed: %v", err)
	}
	require.Equal(t, []uint64{2}, created.ParticipantsIDs, "the owner is not a participant")
	require.Len(t, created.Participants, 1)
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

func TestCertificationService_UpdateMissingCertificationFails(t *testing.T) {
	userID := uint64(1)
	cat := config.AllCertificationCategories[0]
	typeObj := config.AllCertificationTypes[0]
	svc := NewCertificationService(setupTestDAL(t), &mocks.UserDalMock{})

	_, err := svc.UpdateCertification(userID, &types.Certification{
		ID:         uuid.New(),
		UserID:     &userID,
		CategoryID: cat.ID,
		TypeID:     typeObj.ID,
	})
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestCertificationService_ParticipantsAreValidated(t *testing.T) {
	ownerID := uint64(1)
	cat := config.AllCertificationCategories[0]
	typeObj := config.AllCertificationTypes[0]
	newCert := func(participants ...uint64) *types.Certification {
		return &types.Certification{
			ID: uuid.New(), UserID: &ownerID, CategoryID: cat.ID, TypeID: typeObj.ID, ParticipantsIDs: participants,
		}
	}

	t.Run("an unknown participant is rejected and nothing is stored", func(t *testing.T) {
		userDal := &mocks.UserDalMock{}
		userDal.On("FindUsersByIDs", []uint64{2, 99}).Return([]dal.User{{ID: 2}}, nil)
		svc := NewCertificationService(setupTestDAL(t), userDal)

		_, err := svc.CreateCertification(newCert(2, 99))
		require.ErrorContains(t, err, "unknown participant")
		certs, err := svc.ListUserCertifications(ownerID)
		require.NoError(t, err)
		require.Empty(t, certs)
	})

	t.Run("an update with an unknown participant is rejected and keeps the stored ones", func(t *testing.T) {
		userDal := &mocks.UserDalMock{}
		userDal.On("FindUsersByIDs", []uint64{2}).Return([]dal.User{{ID: 2, Name: "Two"}}, nil)
		userDal.On("FindUsersByIDs", []uint64{99}).Return([]dal.User{}, nil)
		svc := NewCertificationService(setupTestDAL(t), userDal)
		created, err := svc.CreateCertification(newCert(2))
		require.NoError(t, err)

		created.ParticipantsIDs = []uint64{99}
		_, err = svc.UpdateCertification(ownerID, &created)
		require.ErrorContains(t, err, "unknown participant")
		got, err := svc.GetCertification(ownerID, created.ID)
		require.NoError(t, err)
		require.Equal(t, []uint64{2}, got.ParticipantsIDs)
	})

	t.Run("too many participants are rejected", func(t *testing.T) {
		ids := make([]uint64, 0, maxParticipants+1)
		for i := uint64(0); i <= maxParticipants; i++ {
			ids = append(ids, i+2)
		}
		svc := NewCertificationService(setupTestDAL(t), &mocks.UserDalMock{})

		_, err := svc.CreateCertification(newCert(ids...))
		require.ErrorContains(t, err, "too many participants")
	})

	t.Run("a certification without participants needs no user lookup", func(t *testing.T) {
		svc := NewCertificationService(setupTestDAL(t), &mocks.UserDalMock{})

		created, err := svc.CreateCertification(newCert())
		require.NoError(t, err)
		_, err = svc.ListUserCertifications(ownerID)
		require.NoError(t, err)
		require.Empty(t, created.Participants)
	})
}
