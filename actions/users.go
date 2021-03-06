package actions

import (
	"log"
	"logbogen/models"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/pkg/errors"
)

// UsersNew renders the form for creating a new User.
// This function is mapped to the path GET /users/new
func UsersNew(c buffalo.Context) error {
	u := models.User{}
	c.Set("user", u)
	return c.Render(200, r.HTML("users/new.html"))
}

// UsersCreate registers a new user with the application.
func UsersCreate(c buffalo.Context) error {

	u := &models.User{}
	if err := c.Bind(u); err != nil {
		return errors.WithStack(err)
	}
	u.Provider = "localuser"
	u.ProviderID = u.Email
	log.Printf("%#v", u)

	tx := c.Value("tx").(*pop.Connection)

	q := tx.Where("email = ?", u.ProviderID)
	exists, err := q.Exists("users")
	if err != nil {
		return errors.WithStack(err)
	}
	if exists {
		verrs := &validate.Errors{Errors: map[string][]string{"email": {"Email alread exists"}}}
		c.Set("user", u)
		c.Set("errors", verrs)
		return c.Render(422, r.Auto(c, u))
	}

	verrs, err := u.Create(tx)
	if err != nil {
		return errors.WithStack(err)
	}

	if verrs.HasAny() {
		c.Set("user", u)
		c.Set("errors", verrs)
		return c.Render(422, r.Auto(c, u))
	}

	c.Session().Set("current_user_id", u.ID)
	c.Flash().Add("success", "Welcome to Buffalo!")

	return c.Redirect(302, "/")
}

func UserShow(c buffalo.Context) error {
	// Allocate an empty user
	user := &models.User{}

	tx := c.Value("tx").(*pop.Connection)

	// To find the user the parameter user_id is used.
	if err := tx.Eager().Find(user, c.Param("user_id")); err != nil {
		return c.Error(404, err)
	}

	return c.Render(200, r.Auto(c, user))
}

func UserEdit(c buffalo.Context) error {
	// Allocate an empty user
	user := &models.User{}

	// To find the user the parameter user_id is used.
	if err := scope(c).Eager().Find(user, currentUser(c).ID); err != nil {
		return c.Error(404, err)
	}

	return c.Render(200, r.Auto(c, user))
}
