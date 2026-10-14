package services

import (
	"testing"

	"github.com/mogensen/logbook/pkg/dal"
	"github.com/mogensen/logbook/pkg/mocks"
	"github.com/stretchr/testify/require"
)

func TestCleanParticipants(t *testing.T) {
	ownerID := uint64(1)

	t.Run("drops owner and duplicates", func(t *testing.T) {
		userDal := &mocks.UserDalMock{}
		userDal.On("FindUsersByIDs", []uint64{2, 3}).Return([]dal.User{{}, {}}, nil)
		got, err := cleanParticipants(userDal, ownerID, []uint64{2, 1, 3, 2})
		require.NoError(t, err)
		require.Equal(t, []uint64{2, 3}, got)
	})

	t.Run("rejects unknown user", func(t *testing.T) {
		userDal := &mocks.UserDalMock{}
		userDal.On("FindUsersByIDs", []uint64{2, 99}).Return([]dal.User{{}}, nil)
		_, err := cleanParticipants(userDal, ownerID, []uint64{2, 99})
		require.Error(t, err)
		require.Contains(t, err.Error(), "unknown participant")
	})

	t.Run("rejects too many participants", func(t *testing.T) {
		ids := make([]uint64, 0, maxParticipants+1)
		for i := uint64(0); i <= maxParticipants; i++ {
			ids = append(ids, i+2)
		}
		userDal := &mocks.UserDalMock{}

		_, err := cleanParticipants(userDal, ownerID, ids)
		require.Error(t, err)
		require.Contains(t, err.Error(), "too many participants")
	})

	t.Run("empty list needs no lookup", func(t *testing.T) {
		userDal := &mocks.UserDalMock{}
		got, err := cleanParticipants(userDal, ownerID, nil)
		require.NoError(t, err)
		require.Empty(t, got)
	})
}
