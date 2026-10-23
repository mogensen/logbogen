package dal

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUserDal_CreateUser(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)

	db.Migrator().AutoMigrate(&User{}, &Activity{})

	userDal := NewUserDal(db)
	user := &User{Name: "Test User", Email: "test@example.com", FirebaseUID: new("uid-1")}

	result := userDal.CreateUser(user)
	require.NotNil(t, result)
	require.NoError(t, result.Error)
}

func TestUserDal_FindUserById(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)

	db.Migrator().AutoMigrate(&User{}, &Activity{})

	userDal := NewUserDal(db)
	user := &User{Name: "Test User", Email: "test@example.com", FirebaseUID: new("uid-1")}
	userResult := userDal.CreateUser(user)
	require.NotNil(t, userResult)
	require.NoError(t, userResult.Error)

	dbUser, err := userDal.FindUserById(1)
	require.NoError(t, err)
	require.NotNil(t, dbUser)

	require.Equal(t, "Test User", dbUser.Name)
	require.Equal(t, "test@example.com", dbUser.Email)
}

func TestUserDal_FindUserByEmail(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)

	db.Migrator().AutoMigrate(&User{}, &Activity{})

	userDal := NewUserDal(db)
	user := &User{Name: "Test User", Email: "test@example.com", FirebaseUID: new("uid-1")}
	userResult := userDal.CreateUser(user)
	require.NotNil(t, userResult)
	require.NoError(t, userResult.Error)

	dbUser, err := userDal.FindUserByEmail("test@example.com")
	require.NoError(t, err)
	require.NotNil(t, dbUser)

	require.Equal(t, "Test User", dbUser.Name)
	require.Equal(t, "test@example.com", dbUser.Email)
}

func TestUserDal_FindUserByFirebaseUID(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)

	db.Migrator().AutoMigrate(&User{}, &Activity{})

	userDal := NewUserDal(db)
	user := &User{Name: "Test User", Email: "test@example.com", FirebaseUID: new("uid-abc")}
	userResult := userDal.CreateUser(user)
	require.NotNil(t, userResult)
	require.NoError(t, userResult.Error)

	dbUser, err := userDal.FindUserByFirebaseUID("uid-abc")
	require.NoError(t, err)
	require.NotNil(t, dbUser)
	require.Equal(t, "Test User", dbUser.Name)
	require.NotNil(t, dbUser.FirebaseUID)
	require.Equal(t, "uid-abc", *dbUser.FirebaseUID)

	_, err = userDal.FindUserByFirebaseUID("uid-missing")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestUserDal_FindUsers(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)

	db.Migrator().AutoMigrate(&User{}, &Activity{})

	userDal := NewUserDal(db)
	user := &User{Name: "Test User", Email: "test@example.com", FirebaseUID: new("uid-1")}
	userResult := userDal.CreateUser(user)
	require.NotNil(t, userResult)
	require.NoError(t, userResult.Error)
	user = &User{Name: "Test User 2", Email: "test2@example.com", FirebaseUID: new("uid-2")}
	userResult = userDal.CreateUser(user)
	require.NotNil(t, userResult)
	require.NoError(t, userResult.Error)

	users, err := userDal.FindUsers()
	require.NoError(t, err)
	require.Equal(t, 2, len(users))
	require.Equal(t, "Test User", users[0].Name)
	require.Equal(t, "test@example.com", users[0].Email)
	require.Equal(t, "Test User 2", users[1].Name)
	require.Equal(t, "test2@example.com", users[1].Email)
}

func TestUserDal_FindUsersByIDs(t *testing.T) {
	db := setupTestDB(t)
	userDal := NewUserDal(db)
	for i, email := range []string{"a@example.com", "b@example.com", "c@example.com"} {
		require.NoError(t, userDal.CreateUser(&User{Name: email, Email: email, FirebaseUID: new(email)}).Error, i)
	}

	users, err := userDal.FindUsersByIDs([]uint64{1, 3, 42})
	require.NoError(t, err)
	require.Len(t, users, 2)

	none, err := userDal.FindUsersByIDs(nil)
	require.NoError(t, err)
	require.Empty(t, none)
}
