package services

import (
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

	userMap, err := s.getUserMap()
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
	userMap, err := s.getUserMap()
	if err != nil {
		return types.Certification{}, err
	}
	return types.CertificationFromDB(cert, userMap), nil
}

func (s *CertificationService) CreateCertification(cert *types.Certification) (types.Certification, error) {
	created, err := s.dal.CreateCertification(cert.ToDB())
	if err != nil {
		return types.Certification{}, err
	}
	userMap, err := s.getUserMap()
	if err != nil {
		return types.Certification{}, err
	}
	return types.CertificationFromDB(created, userMap), nil
}

func (s *CertificationService) UpdateCertification(userID uint64, cert *types.Certification) (types.Certification, error) {
	updated, err := s.dal.UpdateCertification(userID, cert.ToDB())
	userMap, err := s.getUserMap()
	if err != nil {
		return types.Certification{}, err
	}
	return types.CertificationFromDB(updated, userMap), err
}

func (s *CertificationService) DeleteCertification(certID uuid.UUID, userID uint64) error {
	return s.dal.DeleteCertification(certID, userID)
}

func (s *CertificationService) getUserMap() (map[uint64]types.User, error) {
	users, err := s.userDal.FindUsers()
	if err != nil {
		return nil, err
	}
	userMap := make(map[uint64]types.User)
	for _, user := range users {
		userMap[uint64(user.ID)] = *types.UserFromDal(&user)
	}
	return userMap, nil
}
