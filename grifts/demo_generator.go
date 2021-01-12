package grifts

import (
	"encoding/base64"
	"fmt"
	"logbogen/models"
	"logbogen/services"
	"math/rand"
	"time"

	"emperror.dev/errors"
	"github.com/Pallinder/go-randomdata"
	"github.com/brianvoe/gofakeit"
	"github.com/ipsn/go-adorable"

	"golang.org/x/crypto/bcrypt"
)

func createUser(username, password string) error {
	ph, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user := models.User{
		Name:               randomdata.FirstName(randomdata.RandomGender) + " " + randomdata.LastName(),
		Email:              username + "@logbogen.nu",
		Provider:           "localuser",
		ProviderID:         username,
		PasswordHash:       string(ph),
		MemberNumber:       fmt.Sprintf("%d-%s-%d", gofakeit.Year(), randomdata.StringSample("TC", "TI"), randomdata.Number(999)),
		Organization:       randomdata.StringSample("Dansk Træklatreforening", "FDF", "Fjeldgruppen", "DGI"),
		Achievement:        &models.Achievement{},
		ClimbingActivities: &models.Climbingactivities{},
	}
	if err := models.DB.Create(&user); err != nil {
		return errors.WithStack(err)
	}

	avatar := adorable.Random()

	pImage := models.UsersImage{
		ImageData: []byte(base64.StdEncoding.EncodeToString(avatar)),
		UserID:    user.ID,
	}
	models.DB.Create(&pImage)
	return nil
}

func createActivity(user models.User, participants []models.User) error {
	activityTime := randomTimestamp(4, 0)

	randFloats := func(min, max float64) float64 {
		return min + rand.Float64()*(max-min)
	}

	act := &models.Climbingactivity{
		UserID:    user.ID,
		Date:      activityTime,
		Lat:       randFloats(55.885219, 56.3947998),
		Lng:       randFloats(10.147212, 10.2941546),
		Type:      randType(),
		OtherType: randOtherType(),
		Role:      randRole(),
		Comment:   "Fantastisk klatre tur til toppen af det hele. Også kaffe og kage..",
	}

	geo, _ := services.ReverseGeocode(act.Lat, act.Lng)
	act.Location = geo.SimpleDisplayName()
	act.Participants = participants
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
