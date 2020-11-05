package grifts

import (
	"fmt"
	"logbogen/models"
	"math/rand"
	"time"

	"github.com/markbates/grift/grift"
)

var _ = grift.Namespace("demo", func() {

	grift.Desc("seed", "Seeds a database")
	grift.Add("seed", func(c *grift.Context) error {
		for i := 0; i < 20; i++ {
			createUser(fmt.Sprintf("demo-%d", i), "demo")
		}

		users := &models.Users{}

		if err := models.DB.All(users); err != nil {
			return err
		}

		userss := *users
		for i := 0; i < 200; i++ {
			u := userss[rand.Intn(len(userss))]

			nrOfParticipants := rand.Intn(len(userss))
			if nrOfParticipants == 0 {
				nrOfParticipants = 2
			}
			userss = Shuffle(userss)

			err := createActivity(u, userss[:nrOfParticipants])
			if err != nil {
				return err
			}
		}

		return nil
	})

})

func Shuffle(vals models.Users) models.Users {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	ret := make(models.Users, len(vals))
	n := len(vals)
	for i := 0; i < n; i++ {
		randIndex := r.Intn(len(vals))
		ret[i] = vals[randIndex]
		vals = append(vals[:randIndex], vals[randIndex+1:]...)
	}
	return ret
}
