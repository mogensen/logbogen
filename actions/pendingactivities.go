package actions

import (
	"fmt"
	"logbogen/models"
	"net/http"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/x/responder"
	"github.com/gofrs/uuid"
)

// PendingActivitiesList lists all Climbingactivities that the current user is registered on.
// But that the user has not created him self
// This function is mapped to the path
// GET /pendingactivities
func PendingActivitiesList(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	pca := &models.ParticipantsClimbingactivities{}

	tx.Where("user_id = (?)", currentUser(c).ID).All(pca)

	pendingactivities := &models.Climbingactivities{}

	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".
	ids := []uuid.UUID{}
	for _, v := range *pca {
		ids = append(ids, v.ActivityID)
	}

	// Retrieve all Climbingactivities from the DB not owned by current user,
	// but where there exists a ParticipantsClimbingactivities pointing to current user
	q := tx.Where("climbingactivities.user_id != (?)", currentUser(c).ID).Where("id in (?)", ids).Eager("User").Order("Date").PaginateFromParams(c.Params())
	if err := q.All(pendingactivities); err != nil {
		return err
	}

	// Retrieve all Climbingactivities from the DB
	climbingactivities := &models.Climbingactivities{}
	if err := scope(c).All(climbingactivities); err != nil {
		return err
	}

	// Remove already logged climbs
	res := models.Climbingactivities{}
	for _, pending := range *pendingactivities {
		found := false
		for _, act := range *climbingactivities {
			sameDate := pending.Date.Truncate(24 * time.Hour).Equal(act.Date.Truncate(24 * time.Hour))
			sameType := pending.Type == act.Type
			if sameDate && sameType {
				found = true
			}
		}
		if !found {
			res = append(res, pending)
		}
	}

	return responder.Wants("html", func(c buffalo.Context) error {
		// Add the paginator to the context so it can be used in the template.
		// c.Set("pagination", q.Paginator)

		c.Set("climbingactivities", &res)
		return c.Render(http.StatusOK, r.HTML("/climbingactivities/index.plush.html"))
	}).Wants("json", func(c buffalo.Context) error {
		return c.Render(200, r.JSON(&res))
	}).Wants("xml", func(c buffalo.Context) error {
		return c.Render(200, r.XML(&res))
	}).Respond(c)
}
