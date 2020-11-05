package actions

import (
	"logbogen/models"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/helpers/hctx"
	"github.com/gobuffalo/packr/v2"
	"github.com/gofrs/uuid"
)

var r *render.Engine
var assetsBox = packr.New("app:assets", "../public")

func init() {
	r = render.New(render.Options{
		// HTML layout to be used for all HTML requests:
		HTMLLayout: "application.plush.html",

		// Box containing all of the templates:
		TemplatesBox: packr.New("app:templates", "../templates"),
		AssetsBox:    assetsBox,

		// Add template helpers here:
		Helpers: render.Helpers{
			"isActive": func(name string, help hctx.HelperContext) string {
				if cp, ok := help.Value("current_route").(buffalo.RouteInfo); ok {
					if strings.HasPrefix(cp.PathName, name) {
						return "active"
					}
				}
				return "inactive"
			},
			"checkboxChecked": func(id uuid.UUID, slice models.Users) string {
				for _, c := range slice {
					if id == c.ID {
						return "checked"
					}
				}
				return ""
			},
			"selected": func(id uuid.UUID, slice models.Users) string {
				for _, c := range slice {
					if id == c.ID {
						return "selected='selected'"
					}
				}
				return ""
			},
			"image": func(img *models.UsersImage) bool {
				if img != nil && img.ID != uuid.Nil {
					return true
				}
				return false
			},
			"firstUsers": func(n int, users []models.User) []models.User {
				if len(users) > n {
					return users[:n]
				}
				return users
			},
			// for non-bootstrap form helpers uncomment the lines
			// below and import "github.com/gobuffalo/helpers/forms"
			// forms.FormKey:     forms.Form,
			// forms.FormForKey:  forms.FormFor,
		},
	})
}
