package grifts

import (
	"logbogen/models"
	"math/rand"
	"time"

	"emperror.dev/errors"
	"github.com/Pallinder/go-randomdata"
	"github.com/brianvoe/gofakeit"
	"github.com/gobuffalo/nulls"

	"golang.org/x/crypto/bcrypt"
)

func createUser(username, password string) error {
	ph, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user := models.User{
		Name:         randomdata.FirstName(randomdata.RandomGender) + " " + randomdata.LastName(),
		Email:        nulls.NewString(username + "@logbogen.nu"),
		Provider:     "localuser",
		ProviderID:   username,
		PasswordHash: string(ph),
	}
	if err := models.DB.Create(&user); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func createActivity(user models.User) error {
	activityTime := randomTimestamp(4, 0)

	randFloats := func(min, max float64) float64 {
		return min + rand.Float64()*(max-min)
	}

	act := &models.Climbingactivity{
		UserID:  user.ID,
		Date:    activityTime,
		Lat:     randFloats(54.8000145534, 57.730016588),
		Lng:     randFloats(8.08997684086, 12.6900061378),
		Type:    randType(),
		Role:    randRole(),
		Comment: "Fantastisk klatre tur til toppen af det hele. Også kaffe og kage..",
	}
	models.DB.Create(act)

	return nil
}

// StartYear and EndYear is relative to now
//  eg: randomTimestamp(4, 2) gives a random time between two and four years afo
func randomTimestamp(startYear, endYear int) time.Time {
	start := time.Now().AddDate(-startYear, 0, 0)
	end := time.Now().AddDate(-endYear, 0, 0)
	return gofakeit.DateRange(start, end)
}
