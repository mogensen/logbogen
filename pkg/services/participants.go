package services

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mogensen/logbook/pkg/dal"
)

// maxParticipants bounds the participants list of a single activity or certification
const maxParticipants = 50

// cleanParticipants de-duplicates the ids, drops the owner and rejects unknown users
func cleanParticipants(userDal dal.UserDal, ownerID uint64, ids []uint64) ([]uint64, error) {
	seen := make(map[uint64]struct{}, len(ids))
	unique := make([]uint64, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok || id == ownerID {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	if len(unique) > maxParticipants {
		return nil, fiber.NewError(fiber.StatusBadRequest, "too many participants")
	}
	if len(unique) == 0 {
		return unique, nil
	}

	users, err := userDal.FindUsersByIDs(unique)
	if err != nil {
		return nil, err
	}
	if len(users) != len(unique) {
		return nil, fiber.NewError(fiber.StatusBadRequest, "unknown participant")
	}
	return unique, nil
}
