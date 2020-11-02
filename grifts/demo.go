package grifts

import (
	"fmt"
	"logbogen/models"
	"math/rand"

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

		for i := 0; i < 200; i++ {
			u := (*users)[rand.Intn(len(*users))]
			err := createActivity(u)
			if err != nil {
				return err
			}
		}

		return nil
	})

})
