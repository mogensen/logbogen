package actions

import (
	"fmt"
	"logbogen/models"
	"net/http"

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

	climbingactivities := &models.Climbingactivities{}

	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".
	ids := []uuid.UUID{}
	for _, v := range *pca {
		ids = append(ids, v.ActivityID)
	}
	q := tx.Where("climbingactivities.user_id != (?)", currentUser(c).ID).Where("id in (?)", ids).Eager("User").Order("Date").PaginateFromParams(c.Params())

	// // Retrieve all Climbingactivities from the DB
	if err := q.All(climbingactivities); err != nil {
		return err
	}

	return responder.Wants("html", func(c buffalo.Context) error {
		// Add the paginator to the context so it can be used in the template.
		// c.Set("pagination", q.Paginator)

		c.Set("climbingactivities", climbingactivities)
		return c.Render(http.StatusOK, r.HTML("/climbingactivities/index.plush.html"))
	}).Wants("json", func(c buffalo.Context) error {
		return c.Render(200, r.JSON(climbingactivities))
	}).Wants("xml", func(c buffalo.Context) error {
		return c.Render(200, r.XML(climbingactivities))
	}).Respond(c)
}
