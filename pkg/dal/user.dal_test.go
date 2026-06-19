package dal

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func strPtr(s string) *string { return &s }

func TestUserDal_CreateUser(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)

	db.Migrator().AutoMigrate(&User{}, &Activity{})

	userDal := NewUserDal(db)
	user := &User{Name: "Test User", Email: "test@example.com", Auth0Sub: strPtr("auth0|1")}

	result := userDal.CreateUser(user)
	require.NotNil(t, result)
	require.NoError(t, result.Error)
}

func TestUserDal_FindUserById(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)

	db.Migrator().AutoMigrate(&User{}, &Activity{})

	userDal := NewUserDal(db)
	user := &User{Name: "Test User", Email: "test@example.com", Auth0Sub: strPtr("auth0|1")}
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
	user := &User{Name: "Test User", Email: "test@example.com", Auth0Sub: strPtr("auth0|1")}
	userResult := userDal.CreateUser(user)
	require.NotNil(t, userResult)
	require.NoError(t, userResult.Error)

	dbUser, err := userDal.FindUserByEmail("test@example.com")
	require.NoError(t, err)
	require.NotNil(t, dbUser)

	require.Equal(t, "Test User", dbUser.Name)
	require.Equal(t, "test@example.com", dbUser.Email)
}

func TestUserDal_FindUserByAuth0Sub(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)

	db.Migrator().AutoMigrate(&User{}, &Activity{})

	userDal := NewUserDal(db)
	user := &User{Name: "Test User", Email: "test@example.com", Auth0Sub: strPtr("auth0|abc")}
	userResult := userDal.CreateUser(user)
	require.NotNil(t, userResult)
	require.NoError(t, userResult.Error)

	dbUser, err := userDal.FindUserByAuth0Sub("auth0|abc")
	require.NoError(t, err)
	require.NotNil(t, dbUser)
	require.Equal(t, "Test User", dbUser.Name)
	require.NotNil(t, dbUser.Auth0Sub)
	require.Equal(t, "auth0|abc", *dbUser.Auth0Sub)

	_, err = userDal.FindUserByAuth0Sub("auth0|missing")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestUserDal_FindUsers(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)

	db.Migrator().AutoMigrate(&User{}, &Activity{})

	userDal := NewUserDal(db)
	user := &User{Name: "Test User", Email: "test@example.com", Auth0Sub: strPtr("auth0|1")}
	userResult := userDal.CreateUser(user)
	require.NotNil(t, userResult)
	require.NoError(t, userResult.Error)
	user = &User{Name: "Test User 2", Email: "test2@example.com", Auth0Sub: strPtr("auth0|2")}
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
