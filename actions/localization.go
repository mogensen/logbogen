package actions

import (
	"net/http"
	"time"

	"emperror.dev/errors"
	"github.com/gobuffalo/buffalo"
)

func SwitchLanguage(c buffalo.Context) error {
	f := struct {
		Language string `form:"lang"`
		URL      string `form:"url"`
	}{}
	if err := c.Bind(&f); err != nil {
		return errors.WithStack(err)
	}

	// Set new current language using a cookie
	cookie := http.Cookie{
		Name:   "lang",
		Value:  f.Language,
		MaxAge: int((time.Hour * 24 * 265).Seconds()),
		Path:   "/",
	}
	http.SetCookie(c.Response(), &cookie)

	// Update language for the flash message
	T.Refresh(c, f.Language)

	c.Flash().Add("success", T.Translate(c, "users.language-changed", f))

	return c.Redirect(302, f.URL)
}
