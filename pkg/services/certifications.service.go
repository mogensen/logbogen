package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/mogensen/logbook/pkg/config"
	"github.com/mogensen/logbook/pkg/dal"
	"github.com/mogensen/logbook/pkg/types"
)

type CertificationService struct {
	dal     *dal.CertificationService
	userDal dal.UserDal
}

func (s *CertificationService) GetTypes(category string) []types.CertificationType {
	if category == "" {
		// If no category is specified, return all categories and their types
		return config.AllCertificationTypes
	}
	categoryTypes := []types.CertificationType{}
	for _, activityType := range config.AllCertificationTypes {
		if activityType.Category == category {
			categoryTypes = append(categoryTypes, activityType)
		}
	}
	return categoryTypes
}

func NewCertificationService(dal *dal.CertificationService, userDal dal.UserDal) *CertificationService {
	return &CertificationService{dal: dal, userDal: userDal}
}

func (s *CertificationService) ListUserCertifications(userID uint64) ([]types.Certification, error) {
	certifications, err := s.dal.GetCertificationsByUser(userID)
	if err != nil {
		return nil, err
	}

	userMap, err := s.participantMap(certifications...)
	if err != nil {
		return nil, err
	}

	result := make([]types.Certification, len(certifications))
	for i, cert := range certifications {
		result[i] = types.CertificationFromDB(cert, userMap)
	}
	return result, nil
}

func (s *CertificationService) GetCertification(userID uint64, certID uuid.UUID) (types.Certification, error) {
	cert, err := s.dal.GetCertification(userID, certID)
	if err != nil {
		return types.Certification{}, err
	}
	userMap, err := s.participantMap(cert)
	if err != nil {
		return types.Certification{}, err
	}
	return types.CertificationFromDB(cert, userMap), nil
}

func (s *CertificationService) CreateCertification(cert *types.Certification) (types.Certification, error) {
	if err := s.cleanOwnerAndParticipants(cert); err != nil {
		return types.Certification{}, err
	}
	created, err := s.dal.CreateCertification(cert.ToDB())
	if err != nil {
		return types.Certification{}, err
	}
	userMap, err := s.participantMap(created)
	if err != nil {
		return types.Certification{}, err
	}
	return types.CertificationFromDB(created, userMap), nil
}

func (s *CertificationService) UpdateCertification(userID uint64, cert *types.Certification) (types.Certification, error) {
	cert.UserID = &userID
	if err := s.cleanOwnerAndParticipants(cert); err != nil {
		return types.Certification{}, err
	}
	updated, err := s.dal.UpdateCertification(userID, cert.ToDB())
	if err != nil {
		return types.Certification{}, err
	}
	userMap, err := s.participantMap(updated)
	if err != nil {
		return types.Certification{}, err
	}
	return types.CertificationFromDB(updated, userMap), nil
}

func (s *CertificationService) DeleteCertification(certID uuid.UUID, userID uint64) error {
	return s.dal.DeleteCertification(certID, userID)
}

// cleanOwnerAndParticipants replaces the participants with the validated ones, without the owner
func (s *CertificationService) cleanOwnerAndParticipants(cert *types.Certification) error {
	if cert.UserID == nil {
		return errors.New("certification has no owner")
	}
	ids, err := cleanParticipants(s.userDal, *cert.UserID, cert.ParticipantsIDs)
	if err != nil {
		return err
	}
	cert.ParticipantsIDs = ids
	return nil
}

// participantMap looks up only the users that take part in the certifications, which is all the pages show
func (s *CertificationService) participantMap(certs ...dal.Certification) (map[uint64]types.User, error) {
	ids := []uint64{}
	for _, cert := range certs {
		ids = append(ids, cert.Participants...)
	}
	if len(ids) == 0 {
		return map[uint64]types.User{}, nil
	}

	users, err := s.userDal.FindUsersByIDs(ids)
	if err != nil {
		return nil, err
	}
	userMap := make(map[uint64]types.User, len(users))
	for _, user := range users {
		userMap[uint64(user.ID)] = *types.UserFromDal(&user)
	}
	return userMap, nil
}
